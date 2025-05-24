package handlers

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	AiApiUrl            string
	AiApiKey            string
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
var deviceTimers = make(map[string]*time.Timer)

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
	now := time.Now()
	device.LastReceived = &now

	// 更新设备的状态
	device.Status = 1
	deviceID := device.DeviceID
	if timer, ok := deviceTimers[deviceID]; ok {
		timer.Stop() // 停止旧定时器
	}
	deviceTimers[deviceID] = time.AfterFunc(time.Minute, func() {
		// 1分钟后执行
		h.DB.Model(&models.Device{}).Where("device_id = ?", deviceID).Update("status", 0)
	})

	if err := h.DB.Save(device).Error; err != nil {
		utils.Respond(c, nil, utils.ErrInternalServer)
		return
	}

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

type DataAnalysisRequest struct {
	DeviceID  string `json:"device_id" binding:"required"`   // 设备id
	Type      string `json:"report_type" binding:"required"` // 分析类型
	StartTime string `json:"start_time" binding:"required"`  // 开始时间（ISO8601格式）
	EndTime   string `json:"end_time" binding:"required"`    // 结束时间（ISO8601格式）
	Model     string `json:"model" binding:"required"`       // 使用的模型
}

func getAIPrompt(req DataAnalysisRequest, dataList []models.Data) string {
	// 多段统计摘要
	summary := "本时段各统计段数据如下(每一项数据中的四个数据点分别代表平均值、方差、最小值、最大值)\n"
	for i, data := range dataList {
		avg := data.Avg
		variance := data.Var
		min := data.Min
		max := data.Max
		summary += fmt.Sprintf(
			"#%d 温度[%.2f,%.2f,%.2f,%.2f] 湿度[%.2f,%.2f,%.2f,%.2f] 新风[%.2f,%.2f,%.2f,%.2f] 臭氧[%.2f,%.2f,%.2f,%.2f] 二氧化氮[%.2f,%.2f,%.2f,%.2f] 甲醛[%.2f,%.2f,%.2f,%.2f] PM2.5[%.2f,%.2f,%.2f,%.2f] 一氧化碳[%.2f,%.2f,%.2f,%.2f] 细菌[%.2f,%.2f,%.2f,%.2f] 氡气[%.2f,%.2f,%.2f,%.2f];",
			i+1,
			avg.Temperature, variance.Temperature, min.Temperature, max.Temperature,
			avg.Humidity, variance.Humidity, min.Humidity, max.Humidity,
			avg.FreshAir, variance.FreshAir, min.FreshAir, max.FreshAir,
			avg.Ozone, variance.Ozone, min.Ozone, max.Ozone,
			avg.NitroDio, variance.NitroDio, min.NitroDio, max.NitroDio,
			avg.Methanal, variance.Methanal, min.Methanal, max.Methanal,
			avg.Pm25, variance.Pm25, min.Pm25, max.Pm25,
			avg.CarbMomo, variance.CarbMomo, min.CarbMomo, max.CarbMomo,
			avg.Bacteria, variance.Bacteria, min.Bacteria, max.Bacteria,
			avg.Radon, variance.Radon, min.Radon, max.Radon,
		)
	}

	switch req.Type {
	case "0":
		return fmt.Sprintf(
			"请作为环境数据分析专家，针对以下统计数据：\n%s\n请对这些统计数据进行简单分析，归纳主要特征（如均值、极值、波动等），并给出简明结论。",
			summary)
	case "1":
		return fmt.Sprintf(
			"请作为环境异常检测专家，针对以下统计数据：\n%s\n请对这些统计数据进行异常监测，指出是否存在异常数据点或异常趋势，并简要说明可能的原因。",
			summary)
	case "2":
		return fmt.Sprintf(
			"请作为环境趋势预测专家，针对以下统计数据：\n%s\n请分析主要数据的变化趋势，并预测未来一段时间内的走势。请给出趋势判断和预测依据。",
			summary)
	case "3":
		return fmt.Sprintf(
			"请作为环境数据综合分析专家，针对以下统计数据：\n%s\n请进行综合分析，包括数据特征总结、异常检测、趋势预测等内容，给出详细的分析报告和建议。",
			summary)
	default:
		return "请提供有效的分析类型。"
	}
}

func (h *DataHandler) Analysis(c *gin.Context) {
	var req DataAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, utils.ErrMissingParam)
		return
	}

	// 构建查询条件
	query := h.DB.Model(&models.Data{})
	query = query.Where("my_device_id = ?", req.DeviceID)
	query = query.Where("created_at > ?", req.StartTime)
	query = query.Where("created_at < ?", req.EndTime)

	// 查询数据
	var dataList []models.Data
	if err := query.Find(&dataList).Error; err != nil {
		log.Println("解析响应失败：", err)
		utils.Respond(c, nil, utils.ErrInternalServer)
		return
	}
	if len(dataList) == 0 {
		utils.Respond(c, nil, utils.ErrNoData)
		return
	}

	// 生成分析提示
	prompt := getAIPrompt(req, dataList)
	aiReq := map[string]interface{}{
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": prompt,
			},
		},
		"model": req.Model,
	}
	body, _ := json.Marshal(aiReq)

	reqHttp, err := http.NewRequest("POST", h.AiApiUrl, bytes.NewReader(body))
	if err != nil {
		log.Println("解析响应失败：", err)
		utils.Respond(c, nil, utils.ErrExternalService)
		return
	}
	reqHttp.Header.Set("Content-Type", "application/json")
	reqHttp.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.AiApiKey))
	// log.Println("请求头：", reqHttp.Header)
	// log.Println("请求体：", string(body))

	resp, err := http.DefaultClient.Do(reqHttp)
	if err != nil {
		utils.Respond(c, nil, utils.ErrExternalService)
		return
	}
	defer resp.Body.Close()

	var aiResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		utils.Respond(c, nil, utils.ErrInternalServer)
		return
	}

	// 解析响应
	choices, ok := aiResp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		utils.Respond(c, aiResp, utils.ErrExternalService)
		return
	}
	choiceMap, ok := choices[0].(map[string]interface{})
	if !ok {
		utils.Respond(c, aiResp, utils.ErrExternalService)
		return
	}
	message, ok := choiceMap["message"].(map[string]interface{})
	if !ok {
		utils.Respond(c, aiResp, utils.ErrExternalService)
		return
	}
	content, ok := message["content"].(string)
	if !ok {
		utils.Respond(c, aiResp, utils.ErrExternalService)
		return
	}
	// log.Println("AI响应：", content)

	utils.Respond(c, gin.H{
		"detail":     content,
		"data_count": len(dataList),
	}, utils.ErrOK)
}
