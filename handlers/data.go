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
	Season    string           `json:"season" binding:"required"` // 季节
	Scene     string           `json:"scene" binding:"required"`  // 场景
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

	// 校验数据
	anomalyResult := CheckDataAnomaly(reqBody.Data, reqBody.Scene, reqBody.Season)

	if !anomalyResult.IsNormal {
		// log.Printf("检测到异常数据: %v", anomalyResult.AnomalyFields)
		// log.Printf("异常详情: %v", anomalyResult.AnomalyDetails)
		utils.Respond(c, gin.H{
			"anomaly_fields":  anomalyResult.AnomalyFields,
			"anomaly_details": anomalyResult.AnomalyDetails,
		}, utils.ErrDataAnomaly)
		device.Status = 3 // 设置设备状态为异常
		if err := h.DB.Save(device).Error; err != nil {
			utils.Respond(c, nil, utils.ErrInternalServer)
			return
		}
		return
	}

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
			if before, err := time.Parse(time.RFC3339, before); err == nil {
				query = query.Where("created_at < ?", before)
			}
			if after, err := time.Parse(time.RFC3339, after); err == nil {
				query = query.Where("created_at > ?", after)
			}
			return query
		},
	)(c)
}

func (h *DataHandler) MyData(c *gin.Context) {
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
			if before, err := time.Parse(time.RFC3339, before); err == nil {
				query = query.Where("created_at < ?", before)
			}
			if after, err := time.Parse(time.RFC3339, after); err == nil {
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
	if startTime, err := time.Parse(time.RFC3339, req.StartTime); err == nil {
		query = query.Where("created_at > ?", startTime)
	}
	if endTime, err := time.Parse(time.RFC3339, req.EndTime); err == nil {
		query = query.Where("created_at < ?", endTime)
	}

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

// 在 DataHandler 中添加 MyAnalysis 函数
func (h *DataHandler) MyAnalysis(c *gin.Context) {
	var req DataAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, utils.ErrMissingParam)
		return
	}

	// 获取当前用户
	currentUser, exists := c.Get("CurrentUser")
	if !exists {
		utils.Respond(c, nil, utils.ErrUnauthorized)
		return
	}
	user, ok := currentUser.(*models.User)
	if !ok {
		utils.Respond(c, nil, utils.ErrUnauthorized)
		return
	}

	// 查询设备并检查所有者权限
	var device models.Device
	if err := h.DB.Preload("Owner").First(&device, "device_id = ?", req.DeviceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.Respond(c, nil, utils.ErrUnknownDevice)
		} else {
			utils.Respond(c, nil, utils.ErrInternalServer)
		}
		return
	}

	// 检查设备所有者权限
	if device.OwnerID == nil || *device.OwnerID != user.UUID {
		utils.Respond(c, nil, utils.ErrForbidden)
		return
	}

	// 构建查询条件 - 使用设备的UUID而不是DeviceID
	query := h.DB.Model(&models.Data{})
	query = query.Where("my_device_id = ?", device.UUID) // 使用设备的UUID
	if startTime, err := time.Parse(time.RFC3339, req.StartTime); err == nil {
		query = query.Where("created_at > ?", startTime)
	}
	if endTime, err := time.Parse(time.RFC3339, req.EndTime); err == nil {
		query = query.Where("created_at < ?", endTime)
	}

	// 查询数据
	var dataList []models.Data
	if err := query.Find(&dataList).Error; err != nil {
		log.Println("查询数据失败：", err)
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
		log.Println("创建HTTP请求失败：", err)
		utils.Respond(c, nil, utils.ErrExternalService)
		return
	}
	reqHttp.Header.Set("Content-Type", "application/json")
	reqHttp.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.AiApiKey))

	resp, err := http.DefaultClient.Do(reqHttp)
	if err != nil {
		log.Println("AI API请求失败：", err)
		utils.Respond(c, nil, utils.ErrExternalService)
		return
	}
	defer resp.Body.Close()

	var aiResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		log.Println("解析AI响应失败：", err)
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

	utils.Respond(c, gin.H{
		"detail":     content,
		"data_count": len(dataList),
		"device":     device.DeviceID,
	}, utils.ErrOK)
}

// 场景和季节的数据范围定义
type DataRange struct {
	Min float32
	Max float32
}

type SeasonData struct {
	Temperature DataRange
	Humidity    DataRange
	FreshAir    DataRange
	Ozone       DataRange
	NitroDio    DataRange
	Methanal    DataRange
	Pm25        DataRange
	CarbMomo    DataRange
	Bacteria    DataRange
	Radon       DataRange
}

type SceneConfig map[string]map[string]SeasonData

var SCENES = SceneConfig{
	"family": {
		"winter": SeasonData{
			Temperature: DataRange{16, 22},
			Humidity:    DataRange{30, 50},
			FreshAir:    DataRange{0.5, 1.0},
			Ozone:       DataRange{0.019, 0.031},
			NitroDio:    DataRange{0.017, 0.032},
			Methanal:    DataRange{0.021, 0.085},
			Pm25:        DataRange{15.214, 40.032},
			CarbMomo:    DataRange{0.601, 4.054},
			Bacteria:    DataRange{120, 800},
			Radon:       DataRange{1.245, 3.519},
		},
		"summer": SeasonData{
			Temperature: DataRange{24, 28},
			Humidity:    DataRange{40, 70},
			FreshAir:    DataRange{1.0, 1.5},
			Ozone:       DataRange{0.028, 0.054},
			NitroDio:    DataRange{0.021, 0.051},
			Methanal:    DataRange{0.032, 0.112},
			Pm25:        DataRange{10.012, 45.564},
			CarbMomo:    DataRange{0.512, 3.521},
			Bacteria:    DataRange{100, 600},
			Radon:       DataRange{1.065, 3.041},
		},
	},
	"lab": {
		"winter": SeasonData{
			Temperature: DataRange{15, 20},
			Humidity:    DataRange{40, 55},
			FreshAir:    DataRange{2.0, 3.0},
			Ozone:       DataRange{0.005, 0.015},
			NitroDio:    DataRange{0.005, 0.015},
			Methanal:    DataRange{0.014, 0.049},
			Pm25:        DataRange{6.140, 15.001},
			CarbMomo:    DataRange{0.122, 1.575},
			Bacteria:    DataRange{60, 400},
			Radon:       DataRange{0.654, 1.525},
		},
		"summer": SeasonData{
			Temperature: DataRange{20, 24},
			Humidity:    DataRange{45, 60},
			FreshAir:    DataRange{2.5, 4.0},
			Ozone:       DataRange{0.018, 0.028},
			NitroDio:    DataRange{0.017, 0.028},
			Methanal:    DataRange{0.014, 0.055},
			Pm25:        DataRange{5.235, 18.002},
			CarbMomo:    DataRange{0.185, 1.810},
			Bacteria:    DataRange{50, 350},
			Radon:       DataRange{0.540, 1.800},
		},
	},
	"greenhouse": {
		"winter": SeasonData{
			Temperature: DataRange{15, 25},
			Humidity:    DataRange{45, 75},
			FreshAir:    DataRange{1.0, 2.0},
			Ozone:       DataRange{0.015, 0.034},
			NitroDio:    DataRange{0.019, 0.031},
			Methanal:    DataRange{0.010, 0.041},
			Pm25:        DataRange{12.201, 35.203},
			CarbMomo:    DataRange{0.201, 2.530},
			Bacteria:    DataRange{120, 700},
			Radon:       DataRange{0.650, 2.510},
		},
		"summer": SeasonData{
			Temperature: DataRange{20, 30},
			Humidity:    DataRange{50, 80},
			FreshAir:    DataRange{1.5, 3.0},
			Ozone:       DataRange{0.011, 0.041},
			NitroDio:    DataRange{0.019, 0.041},
			Methanal:    DataRange{0.015, 0.052},
			Pm25:        DataRange{10.325, 38.914},
			CarbMomo:    DataRange{0.284, 2.857},
			Bacteria:    DataRange{100, 750},
			Radon:       DataRange{0.549, 2.875},
		},
	},
}

// 检查单个数值是否在范围内
func isInRange(value float32, dataRange DataRange) bool {
	return value >= dataRange.Min && value <= dataRange.Max
}

// 异常检测结果
type AnomalyResult struct {
	IsNormal       bool              `json:"is_normal"`
	AnomalyFields  []string          `json:"anomaly_fields"`
	AnomalyDetails map[string]string `json:"anomaly_details"`
}

// 判断数据是否正常的函数
func CheckDataAnomaly(data models.DataEntry, scene, season string) AnomalyResult {
	result := AnomalyResult{
		IsNormal:       true,
		AnomalyFields:  []string{},
		AnomalyDetails: make(map[string]string),
	}

	// 检查场景和季节是否存在
	sceneData, exists := SCENES[scene]
	if !exists {
		result.IsNormal = false
		result.AnomalyFields = append(result.AnomalyFields, "scene")
		result.AnomalyDetails["scene"] = "未知场景类型"
		return result
	}

	seasonData, exists := sceneData[season]
	if !exists {
		result.IsNormal = false
		result.AnomalyFields = append(result.AnomalyFields, "season")
		result.AnomalyDetails["season"] = "未知季节类型"
		return result
	}

	// 检查各项数据是否在正常范围内
	checks := map[string]struct {
		value  float32
		range_ DataRange
	}{
		"temperature": {data.Temperature, seasonData.Temperature},
		"humidity":    {data.Humidity, seasonData.Humidity},
		"fresh_air":   {data.FreshAir, seasonData.FreshAir},
		"ozone":       {data.Ozone, seasonData.Ozone},
		"nitro_dio":   {data.NitroDio, seasonData.NitroDio},
		"methanal":    {data.Methanal, seasonData.Methanal},
		"pm2_5":       {data.Pm25, seasonData.Pm25},
		"carb_momo":   {data.CarbMomo, seasonData.CarbMomo},
		"bacteria":    {data.Bacteria, seasonData.Bacteria},
		"radon":       {data.Radon, seasonData.Radon},
	}

	for fieldName, check := range checks {
		if !isInRange(check.value, check.range_) {
			result.IsNormal = false
			result.AnomalyFields = append(result.AnomalyFields, fieldName)
			result.AnomalyDetails[fieldName] = fmt.Sprintf(
				"数值%.3f超出正常范围[%.3f, %.3f]",
				check.value, check.range_.Min, check.range_.Max,
			)
		}
	}

	return result
}
