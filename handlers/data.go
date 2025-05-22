package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"ssat_backend_rebuild/models"
	"ssat_backend_rebuild/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type DataHandler struct {
	MongoCollection     *mongo.Collection
	MongoToSQLThreshold int
	BaseHandler[models.Data]
}

type DataUploadRequest struct {
	DeviceID  string           `json:"device_id" binding:"required"`
	Timestamp int64            `json:"timestamp" binding:"required"`
	Data      models.DataEntry `json:"data" binding:"required"`
	Signature string           `json:"signature" binding:"required"`
}

type MongoData struct {
	DeviceID  string           `json:"device_id" bson:"device_id"`
	Timestamp int64            `json:"timestamp" bson:"timestamp"`
	Data      models.DataEntry `json:"data" bson:"data"`
	Processed bool             `json:"processed" bson:"processed"`
}

var DataCache = cache.New(5*time.Minute, 10*time.Minute)

// 计算最大值、最小值、平均值、方差
func CalcStats(data []float32) (max, min, avg, variance float32) {
	if len(data) == 0 {
		return
	}
	max, min = data[0], data[0]
	var sum float32
	for _, v := range data {
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
		sum += v
	}
	avg = sum / float32(len(data))
	for _, v := range data {
		variance += (v - avg) * (v - avg)
	}
	variance /= float32(len(data))
	return
}

func CalcStatsForMongoData(data []MongoData) (maxEntry, minEntry, avgEntry, varEntry models.DataEntry) {
	var temps, hums, freshAirs, ozones, nitroDios, methanals, pm25s, carbMomos, bacterias, radons []float32
	for _, d := range data {
		temps = append(temps, d.Data.Temperature)
		hums = append(hums, d.Data.Humidity)
		freshAirs = append(freshAirs, d.Data.FreshAir)
		ozones = append(ozones, d.Data.Ozone)
		nitroDios = append(nitroDios, d.Data.NitroDio)
		methanals = append(methanals, d.Data.Methanal)
		pm25s = append(pm25s, d.Data.Pm25)
		carbMomos = append(carbMomos, d.Data.CarbMomo)
		bacterias = append(bacterias, d.Data.Bacteria)
		radons = append(radons, d.Data.Radon)
	}
	maxEntry.Temperature, minEntry.Temperature, avgEntry.Temperature, varEntry.Temperature = CalcStats(temps)
	maxEntry.Humidity, minEntry.Humidity, avgEntry.Humidity, varEntry.Humidity = CalcStats(hums)
	maxEntry.FreshAir, minEntry.FreshAir, avgEntry.FreshAir, varEntry.FreshAir = CalcStats(freshAirs)
	maxEntry.Ozone, minEntry.Ozone, avgEntry.Ozone, varEntry.Ozone = CalcStats(ozones)
	maxEntry.NitroDio, minEntry.NitroDio, avgEntry.NitroDio, varEntry.NitroDio = CalcStats(nitroDios)
	maxEntry.Methanal, minEntry.Methanal, avgEntry.Methanal, varEntry.Methanal = CalcStats(methanals)
	maxEntry.Pm25, minEntry.Pm25, avgEntry.Pm25, varEntry.Pm25 = CalcStats(pm25s)
	maxEntry.CarbMomo, minEntry.CarbMomo, avgEntry.CarbMomo, varEntry.CarbMomo = CalcStats(carbMomos)
	maxEntry.Bacteria, minEntry.Bacteria, avgEntry.Bacteria, varEntry.Bacteria = CalcStats(bacterias)
	maxEntry.Radon, minEntry.Radon, avgEntry.Radon, varEntry.Radon = CalcStats(radons)
	return maxEntry, minEntry, avgEntry, varEntry
}

func (h *DataHandler) Upload(c *gin.Context) {
	// 解析请求体
	var reqBody DataUploadRequest
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		utils.Respond(c, nil, utils.ErrMissingParam)
		return
	}

	// 验证设备ID
	device := &models.Device{}
	if err := h.DB.First(device, "device_id = ?", reqBody.DeviceID).Error; err != nil {
		utils.Respond(c, nil, utils.ErrUnknownDevice)
		return
	}
	c.Set("CurrentDevice", device)

	// 校验时间戳
	if time.Since(time.Unix(reqBody.Timestamp, 0)) > time.Minute {
		utils.Respond(c, nil, utils.ErrExpiredRequest)
		return
	}

	// 验证签名
	hash := md5.Sum([]byte(fmt.Sprintf("%s:%d:%s", reqBody.DeviceID, reqBody.Timestamp, device.Secret)))
	hashedSignature := hex.EncodeToString(hash[:])
	if reqBody.Signature != hashedSignature {
		utils.Respond(c, nil, utils.ErrInvalidSignature)
		return
	}

	// 记录签名，防止重放攻击
	if _, found := DataCache.Get(reqBody.Signature); found {
		utils.Respond(c, nil, utils.ErrReplayAttack)
		return
	}
	DataCache.Set(reqBody.Signature, true, 2*time.Minute)

	// 创建数据记录
	mongoData := MongoData{
		DeviceID:  reqBody.DeviceID,
		Timestamp: reqBody.Timestamp,
		Data:      reqBody.Data,
		Processed: false,
	}
	if _, err := h.MongoCollection.InsertOne(c, mongoData); err != nil {
		utils.Respond(c, nil, utils.ErrInternalServer)
		return
	}

	filter := bson.M{"device_id": reqBody.DeviceID, "processed": false}
	count, err := h.MongoCollection.CountDocuments(c, filter)
	if err != nil {
		utils.Respond(c, nil, utils.ErrInternalServer)
		return
	}
	// log.Println(count)

	if count >= int64(h.MongoToSQLThreshold) {
		log.Println("数据超过阈值，开始处理")
		// 获取未处理的数据
		cursor, err := h.MongoCollection.Find(c, filter)
		if err != nil {
			utils.Respond(c, nil, utils.ErrInternalServer)
			return
		}
		defer cursor.Close(c)

		// 计算数据统计信息
		var mongoDataList []MongoData
		if err := cursor.All(c, &mongoDataList); err != nil {
			utils.Respond(c, nil, utils.ErrInternalServer)
			return
		}
		maxEntry, minEntry, avgEntry, varEntry := CalcStatsForMongoData(mongoDataList)

		// 根据 DeviceID 查询设备
		device := &models.Device{}
		if err := h.DB.First(device, "device_id = ?", reqBody.DeviceID).Error; err != nil {
			utils.Respond(c, nil, utils.ErrUnknownDevice)
			return
		}

		// 将统计学数据插入到 SQL 数据库
		data := models.Data{
			MyDevice: device,
			Avg:      avgEntry,
			Var:      varEntry,
			Min:      minEntry,
			Max:      maxEntry,
		}
		if err := h.DB.Create(&data).Error; err != nil {
			utils.Respond(c, nil, utils.ErrInternalServer)
			return
		}

		// 更新 MongoDB 中的数据为已处理
		if _, err := h.MongoCollection.UpdateMany(c, filter, map[string]any{"$set": map[string]any{"processed": true}}); err != nil {
			utils.Respond(c, nil, utils.ErrInternalServer)
			return
		}
	}

	// if err := h.DB.Create(&reqBody).Error; err != nil {
	// 	utils.Respond(c, nil, utils.ErrInternalServer)
	// 	return
	// }

	// 更新设备的最后接收时间
	*device.LastReceived = time.Now()
	if err := h.DB.Save(device).Error; err != nil {
		utils.Respond(c, nil, utils.ErrInternalServer)
		return
	}
	log.Println(device)

	utils.Respond(c, nil, utils.ErrOK)
}

func (h *DataHandler) List(c *gin.Context) {
	h.BaseHandler.List(
		nil,
		func(c *gin.Context, query *gorm.DB) *gorm.DB {
			// 通过设备ID查询数据
			deviceId := c.Query("device_id")
			before := c.Query("before")
			after := c.Query("after")

			if deviceId != "" {
				query = query.Where("my_device_id = ?", deviceId)
			}
			if before != "" {
				query = query.Where("created_at < ?", before)
			}
			if after != "" {
				query = query.Where("created_at > ?", after)
			}
			return query
		},
	)(c)
}
