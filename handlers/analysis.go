package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"ssat_backend_rebuild/models"
	"ssat_backend_rebuild/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// AnalysisHandler 处理数据分析相关请求
type AnalysisHandler struct {
	DB              *gorm.DB
	MongoCollection *mongo.Collection
	Cache           *cache.Cache
	CacheDuration   time.Duration
	InProgressCache *cache.Cache // 用于跟踪正在进行的分析
	AIApiKey        string       // AI模型API密钥
}

// NewAnalysisHandler 创建一个新的分析处理器
func NewAnalysisHandler(db *gorm.DB, mongoCollection *mongo.Collection) *AnalysisHandler {
	// 从环境变量获取API密钥，如果没有设置则使用默认值"lbwnb"
	apiKey := os.Getenv("AI_API_KEY")
	if apiKey == "" {
		apiKey = "lbwnb"
	}

	return &AnalysisHandler{
		DB:              db,
		MongoCollection: mongoCollection,
		Cache:           cache.New(24*time.Hour, 1*time.Hour), // 缓存24小时，每小时清理一次过期项
		CacheDuration:   24 * time.Hour,
		InProgressCache: cache.New(5*time.Minute, 1*time.Minute), // 进行中的分析缓存5分钟
		AIApiKey:        apiKey,
	}
}

// Analyse 处理分析请求
// GET /data/analyse
func (h *AnalysisHandler) Analyse(c *gin.Context) {
	var req models.AnalysisRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "请求参数无效", err)
		return
	}

	// 参数验证
	if req.TimeStart >= req.TimeEnd {
		utils.RespondWithError(c, http.StatusBadRequest, "时间范围无效", errors.New("开始时间必须早于结束时间"))
		return
	}

	// 检查设备是否存在
	var device models.Device
	result := h.DB.Where("device_id = ?", req.DeviceID).First(&device)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			utils.RespondWithError(c, http.StatusNotFound, "设备不存在", result.Error)
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, "数据库查询错误", result.Error)
		}
		return
	}

	// 构建缓存键
	cacheKey := models.AnalysisCacheKey{
		DeviceID:   req.DeviceID,
		Model:      req.Model,
		ReportType: req.ReportType,
		TimeStart:  req.TimeStart,
		TimeEnd:    req.TimeEnd,
	}.String()

	// 检查是否已有相同分析正在进行
	if _, found := h.InProgressCache.Get(cacheKey); found {
		utils.RespondWithError(c, http.StatusTooManyRequests, "相同的分析请求正在处理中", errors.New("请稍后重试"))
		return
	}

	// 检查缓存中是否有结果
	if cachedResult, found := h.Cache.Get(cacheKey); found {
		c.JSON(http.StatusOK, cachedResult)
		return
	}

	// 标记分析正在进行
	h.InProgressCache.Set(cacheKey, true, cache.DefaultExpiration)
	defer h.InProgressCache.Delete(cacheKey)

	// 从MongoDB中获取设备数据
	deviceData, err := h.getDeviceDataFromMongo(req.DeviceID, req.TimeStart, req.TimeEnd)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "获取设备数据失败", err)
		return
	}

	if len(deviceData) == 0 {
		utils.RespondWithError(c, http.StatusNotFound, "指定时间范围内没有数据", errors.New("未找到数据"))
		return
	}

	// 获取相应的AI模型
	aiModel, err := utils.ModelFactory(req.Model)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "不支持的AI模型", err)
		return
	}

	// 附加API密钥信息到日志中
	c.Set("ai_api_key", h.AIApiKey)

	// 处理分析
	analysisResult, err := aiModel.ProcessAnalysis(deviceData, req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "分析处理失败", err)
		return
	}

	// 设置UUID
	analysisResult.UUID = uuid.New().String()

	// 将结果存入缓存
	h.Cache.Set(cacheKey, analysisResult, h.CacheDuration)

	c.JSON(http.StatusOK, analysisResult)
}

// getDeviceDataFromMongo 从MongoDB中获取指定设备在给定时间范围内的数据
func (h *AnalysisHandler) getDeviceDataFromMongo(deviceID string, timeStart, timeEnd int64) ([]models.DataEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	startTime := time.Unix(timeStart, 0)
	endTime := time.Unix(timeEnd, 0)

	// 构建查询条件
	filter := bson.M{
		"device_id": deviceID,
		"timestamp": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	}

	// 执行查询
	cursor, err := h.MongoCollection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("MongoDB查询失败: %w", err)
	}
	defer cursor.Close(ctx)

	// 从游标中获取数据
	type MongoDataEntry struct {
		Timestamp time.Time        `bson:"timestamp"`
		Data      models.DataEntry `bson:"data"`
	}

	var entries []MongoDataEntry
	if err := cursor.All(ctx, &entries); err != nil {
		return nil, fmt.Errorf("解析MongoDB数据失败: %w", err)
	}

	// 提取数据部分
	result := make([]models.DataEntry, len(entries))
	for i, entry := range entries {
		result[i] = entry.Data
	}

	return result, nil
}

// ClearCache 清除指定设备的分析缓存
func (h *AnalysisHandler) ClearCache(c *gin.Context) {
	deviceID := c.Param("device_id")
	if deviceID == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "缺少设备ID", errors.New("设备ID是必需的"))
		return
	}

	// 遍历缓存，删除包含特定设备ID的所有项
	for k := range h.Cache.Items() {
		if k[:len(deviceID)] == deviceID {
			h.Cache.Delete(k)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "设备缓存已清除"})
}
