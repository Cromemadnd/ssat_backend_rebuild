package utils

import (
	"errors"
	"fmt"
	"ssat_backend_rebuild/models"
	"time"
)

// AIModelService 接口定义了与AI模型交互的方法
type AIModelService interface {
	// 处理分析
	ProcessAnalysis(deviceData []models.DataEntry, request models.AnalysisRequest) (*models.AnalysisResult, error)
}

// DeepSeekV3Model 实现DeepSeek-V3模型的分析功能
type DeepSeekV3Model struct{}

// ProcessAnalysis 处理DeepSeek-V3的分析请求
func (m *DeepSeekV3Model) ProcessAnalysis(deviceData []models.DataEntry, request models.AnalysisRequest) (*models.AnalysisResult, error) {
	switch request.ReportType {
	case "simple":
		return m.processSimpleAnalysis(deviceData, request)
	case "anomaly":
		return m.processAnomalyDetection(deviceData, request)
	case "trend":
		return m.processTrendPrediction(deviceData, request)
	case "comprehensive":
		return m.processComprehensiveAnalysis(deviceData, request)
	default:
		return nil, errors.New("不支持的报告类型")
	}
}

// DeepSeekR1Model 实现DeepSeek-R1模型的分析功能
type DeepSeekR1Model struct{}

// ProcessAnalysis 处理DeepSeek-R1的分析请求
func (m *DeepSeekR1Model) ProcessAnalysis(deviceData []models.DataEntry, request models.AnalysisRequest) (*models.AnalysisResult, error) {
	switch request.ReportType {
	case "simple":
		return m.processSimpleAnalysis(deviceData, request)
	case "anomaly":
		return m.processAnomalyDetection(deviceData, request)
	case "trend":
		return m.processTrendPrediction(deviceData, request)
	case "comprehensive":
		return m.processComprehensiveAnalysis(deviceData, request)
	default:
		return nil, errors.New("不支持的报告类型")
	}
}

// ModelFactory 返回一个基于模型名称的AI模型实例
func ModelFactory(modelName string) (AIModelService, error) {
	switch modelName {
	case "DeepSeek-V3":
		return &DeepSeekV3Model{}, nil
	case "DeepSeek-R1":
		return &DeepSeekR1Model{}, nil
	default:
		return nil, errors.New("不支持的模型类型")
	}
}

// --- DeepSeekV3Model的分析方法实现 ---

func (m *DeepSeekV3Model) processSimpleAnalysis(data []models.DataEntry, request models.AnalysisRequest) (*models.AnalysisResult, error) {
	// 简单分析实现
	result := createBaseAnalysisResult(request)
	// 计算简单统计数据
	result.Summary = models.AnalysisSummary{
		Status:         "正常",
		AnomalyCount:   0,
		Recommendation: "环境数据正常，保持当前设置",
		Alert:          false,
	}

	// 添加简单图表数据
	result.Charts = createBasicCharts(data)

	// 添加详细分析项
	result.DetailedItems = generateDetailedAnalysisItems(data, "simple")

	return result, nil
}

func (m *DeepSeekV3Model) processAnomalyDetection(data []models.DataEntry, request models.AnalysisRequest) (*models.AnalysisResult, error) {
	// 异常检测实现
	result := createBaseAnalysisResult(request)

	// 模拟异常检测
	anomalyCount := 0
	for _, entry := range data {
		// 简单示例：当温度超过30度或低于10度时，视为异常
		if entry.Temperature > 30 || entry.Temperature < 10 {
			anomalyCount++
		}
	}

	var status string
	if anomalyCount > 0 {
		status = "异常"
	} else {
		status = "正常"
	}

	var recommendation string
	if anomalyCount > 0 {
		recommendation = "建议检查温度控制系统"
	} else {
		recommendation = "环境数据正常，无需采取行动"
	}

	var alertMessage string
	if anomalyCount > 5 {
		alertMessage = "温度异常次数过多，请立即检查"
	}

	result.Summary = models.AnalysisSummary{
		Status:         status,
		AnomalyCount:   anomalyCount,
		Recommendation: recommendation,
		Alert:          anomalyCount > 5,
		AlertMessage:   alertMessage,
	}

	// 添加异常检测图表
	result.Charts = createAnomalyCharts(data)

	// 添加详细分析项
	result.DetailedItems = generateDetailedAnalysisItems(data, "anomaly")

	return result, nil
}

func (m *DeepSeekV3Model) processTrendPrediction(data []models.DataEntry, request models.AnalysisRequest) (*models.AnalysisResult, error) {
	// 趋势预测实现
	result := createBaseAnalysisResult(request)

	result.Summary = models.AnalysisSummary{
		Status:         "稳定",
		AnomalyCount:   0,
		Recommendation: "预测环境指标将保持稳定趋势",
		Alert:          false,
	}

	// 添加趋势预测图表
	result.Charts = createTrendCharts(data)

	// 添加详细分析项
	result.DetailedItems = generateDetailedAnalysisItems(data, "trend")

	return result, nil
}

func (m *DeepSeekV3Model) processComprehensiveAnalysis(data []models.DataEntry, request models.AnalysisRequest) (*models.AnalysisResult, error) {
	// 综合分析实现
	result := createBaseAnalysisResult(request)

	result.Summary = models.AnalysisSummary{
		Status:         "正常",
		AnomalyCount:   0,
		Recommendation: "综合环境指标良好，建议定期维护设备",
		Alert:          false,
	}

	// 添加综合分析图表
	result.Charts = createComprehensiveCharts(data)

	// 添加详细分析项
	result.DetailedItems = generateDetailedAnalysisItems(data, "comprehensive")

	return result, nil
}

// --- DeepSeekR1Model的分析方法实现 ---

func (m *DeepSeekR1Model) processSimpleAnalysis(data []models.DataEntry, request models.AnalysisRequest) (*models.AnalysisResult, error) {
	// DeepSeek-R1的简单分析实现（与V3类似但可能有不同逻辑）
	result := createBaseAnalysisResult(request)

	result.Summary = models.AnalysisSummary{
		Status:         "正常",
		AnomalyCount:   0,
		Recommendation: "环境数据符合预期，继续当前操作",
		Alert:          false,
	}

	// 添加简单图表数据（R1版本）
	result.Charts = createBasicCharts(data)

	// 添加详细分析项（R1版本）
	result.DetailedItems = generateDetailedAnalysisItems(data, "simple")

	return result, nil
}

func (m *DeepSeekR1Model) processAnomalyDetection(data []models.DataEntry, request models.AnalysisRequest) (*models.AnalysisResult, error) {
	// R1模型的异常检测实现
	result := createBaseAnalysisResult(request)

	// R1模型的异常检测逻辑
	anomalyCount := 0
	for _, entry := range data {
		// R1模型可能使用不同的阈值或算法
		if entry.Humidity < 20 || entry.Humidity > 70 {
			anomalyCount++
		}
	}

	// 使用if-else替代三元运算符
	var status string
	if anomalyCount > 0 {
		status = "异常"
	} else {
		status = "正常"
	}

	var recommendation string
	if anomalyCount > 0 {
		recommendation = "建议调整湿度控制系统"
	} else {
		recommendation = "环境湿度正常，保持现状"
	}

	var alertMessage string
	if anomalyCount > 3 {
		alertMessage = "湿度异常次数较多，请注意调整"
	}

	result.Summary = models.AnalysisSummary{
		Status:         status,
		AnomalyCount:   anomalyCount,
		Recommendation: recommendation,
		Alert:          anomalyCount > 3,
		AlertMessage:   alertMessage,
	}

	// 添加异常检测图表（R1版本）
	result.Charts = createAnomalyCharts(data)

	// 添加详细分析项（R1版本）
	result.DetailedItems = generateDetailedAnalysisItems(data, "anomaly")

	return result, nil
}

func (m *DeepSeekR1Model) processTrendPrediction(data []models.DataEntry, request models.AnalysisRequest) (*models.AnalysisResult, error) {
	// R1模型的趋势预测实现
	result := createBaseAnalysisResult(request)

	result.Summary = models.AnalysisSummary{
		Status:         "上升",
		AnomalyCount:   0,
		Recommendation: "预测环境指标呈上升趋势，建议关注",
		Alert:          false,
	}

	// 添加趋势预测图表（R1版本）
	result.Charts = createTrendCharts(data)

	// 添加详细分析项（R1版本）
	result.DetailedItems = generateDetailedAnalysisItems(data, "trend")

	return result, nil
}

func (m *DeepSeekR1Model) processComprehensiveAnalysis(data []models.DataEntry, request models.AnalysisRequest) (*models.AnalysisResult, error) {
	// R1模型的综合分析实现
	result := createBaseAnalysisResult(request)

	result.Summary = models.AnalysisSummary{
		Status:         "良好",
		AnomalyCount:   0,
		Recommendation: "综合环境指标良好，建议优化空气质量",
		Alert:          false,
	}

	// 添加综合分析图表（R1版本）
	result.Charts = createComprehensiveCharts(data)

	// 添加详细分析项（R1版本）
	result.DetailedItems = generateDetailedAnalysisItems(data, "comprehensive")

	return result, nil
}

// --- 辅助函数 ---

// 创建基础分析结果
func createBaseAnalysisResult(request models.AnalysisRequest) *models.AnalysisResult {
	return &models.AnalysisResult{
		DeviceID:      request.DeviceID,
		Model:         request.Model,
		ReportType:    request.ReportType,
		GeneratedAt:   time.Unix(request.TimeStart, 0),
		TimeRange:     [2]int64{request.TimeStart, request.TimeEnd},
		Charts:        make([]models.ChartData, 0),
		DetailedItems: make([]models.DetailedAnalysisItem, 0),
	}
}

// 创建基本图表
func createBasicCharts(data []models.DataEntry) []models.ChartData {
	// 生成所有环境指标的图表
	charts := []models.ChartData{}

	// 温度图表
	charts = append(charts, createSingleMetricChart(data, "温度", func(entry models.DataEntry) float64 {
		return float64(entry.Temperature)
	}))

	// 湿度图表
	charts = append(charts, createSingleMetricChart(data, "湿度", func(entry models.DataEntry) float64 {
		return float64(entry.Humidity)
	}))

	// 新鲜空气图表
	charts = append(charts, createSingleMetricChart(data, "新鲜空气", func(entry models.DataEntry) float64 {
		return float64(entry.FreshAir)
	}))

	// 臭氧图表
	charts = append(charts, createSingleMetricChart(data, "臭氧", func(entry models.DataEntry) float64 {
		return float64(entry.Ozone)
	}))

	// PM2.5图表
	charts = append(charts, createSingleMetricChart(data, "PM2.5", func(entry models.DataEntry) float64 {
		return float64(entry.Pm25)
	}))

	return charts
}

// 创建单指标图表
func createSingleMetricChart(data []models.DataEntry, label string, valueExtractor func(models.DataEntry) float64) models.ChartData {
	dataset := models.ChartDataset{
		Label: label,
		Data:  make([]models.ChartDataPoint, 0, len(data)),
	}

	for i, entry := range data {
		timestamp := int64(i) * 3600 // 每小时一个点

		dataset.Data = append(dataset.Data, models.ChartDataPoint{
			Timestamp: timestamp,
			Value:     valueExtractor(entry),
		})
	}

	return models.ChartData{
		Type:     "line",
		Datasets: []models.ChartDataset{dataset},
	}
}

// 创建异常检测图表
func createAnomalyCharts(data []models.DataEntry) []models.ChartData {
	// 基于基本图表添加异常标记
	charts := createBasicCharts(data)

	// 添加异常标记
	for i, entry := range data {
		// 温度异常判断
		if entry.Temperature > 30 || entry.Temperature < 10 {
			charts[0].Datasets[0].Data[i].IsAnomaly = true
			charts[0].Datasets[0].Data[i].Label = "温度异常"
		}

		// 湿度异常判断
		if entry.Humidity < 20 || entry.Humidity > 70 {
			charts[1].Datasets[0].Data[i].IsAnomaly = true
			charts[1].Datasets[0].Data[i].Label = "湿度异常"
		}

		// 臭氧异常判断
		if entry.Ozone > 0.1 {
			charts[3].Datasets[0].Data[i].IsAnomaly = true
			charts[3].Datasets[0].Data[i].Label = "臭氧异常"
		}

		// PM2.5异常判断
		if entry.Pm25 > 75 {
			charts[4].Datasets[0].Data[i].IsAnomaly = true
			charts[4].Datasets[0].Data[i].Label = "PM2.5异常"
		}
	}

	return charts
}

// 创建趋势预测图表
func createTrendCharts(data []models.DataEntry) []models.ChartData {
	// 基于基本图表添加趋势线
	charts := createBasicCharts(data)

	// 添加趋势数据集
	trendTempDataset := models.ChartDataset{
		Label: "温度趋势",
		Data:  make([]models.ChartDataPoint, 0),
	}

	trendHumidityDataset := models.ChartDataset{
		Label: "湿度趋势",
		Data:  make([]models.ChartDataPoint, 0),
	}

	// 简单线性趋势（实际中可能使用更复杂的算法）
	dataLen := len(data)
	if dataLen > 0 {
		// 起始点
		trendTempDataset.Data = append(trendTempDataset.Data, models.ChartDataPoint{
			Timestamp: 0,
			Value:     float64(data[0].Temperature),
		})

		trendHumidityDataset.Data = append(trendHumidityDataset.Data, models.ChartDataPoint{
			Timestamp: 0,
			Value:     float64(data[0].Humidity),
		})

		// 结束点
		trendTempDataset.Data = append(trendTempDataset.Data, models.ChartDataPoint{
			Timestamp: int64(dataLen-1) * 3600,
			Value:     float64(data[dataLen-1].Temperature),
		})

		trendHumidityDataset.Data = append(trendHumidityDataset.Data, models.ChartDataPoint{
			Timestamp: int64(dataLen-1) * 3600,
			Value:     float64(data[dataLen-1].Humidity),
		})

		// 预测未来点
		trendTempDataset.Data = append(trendTempDataset.Data, models.ChartDataPoint{
			Timestamp: int64(dataLen+24) * 3600,                   // 预测24小时后
			Value:     float64(data[dataLen-1].Temperature) * 1.1, // 简单增长模型
			Label:     "预测值",
		})

		trendHumidityDataset.Data = append(trendHumidityDataset.Data, models.ChartDataPoint{
			Timestamp: int64(dataLen+24) * 3600,                 // 预测24小时后
			Value:     float64(data[dataLen-1].Humidity) * 0.95, // 简单下降模型
			Label:     "预测值",
		})
	}

	charts[0].Datasets = append(charts[0].Datasets, trendTempDataset)
	charts[1].Datasets = append(charts[1].Datasets, trendHumidityDataset)

	return charts
}

// 创建综合分析图表
func createComprehensiveCharts(data []models.DataEntry) []models.ChartData {
	// 结合简单、异常分析的图表
	basicCharts := createBasicCharts(data)
	anomalyCharts := createAnomalyCharts(data)
	// 不需要趋势图表变量
	// trendCharts := createTrendCharts(data)

	// 创建空气质量综合指数图表
	airQualityDataset := models.ChartDataset{
		Label: "空气质量指数",
		Data:  make([]models.ChartDataPoint, 0, len(data)),
	}

	for i, entry := range data {
		// 综合多个指标计算空气质量
		aqiValue := calculateAQI(entry)

		airQualityDataset.Data = append(airQualityDataset.Data, models.ChartDataPoint{
			Timestamp: int64(i) * 3600,
			Value:     aqiValue,
			Label:     getAQILevel(aqiValue),
		})
	}

	comprehensiveCharts := append(basicCharts, models.ChartData{
		Type:     "line",
		Datasets: []models.ChartDataset{airQualityDataset},
	})

	// 添加异常检测图表中的异常标记
	for i := range comprehensiveCharts[0].Datasets[0].Data {
		if i < len(anomalyCharts[0].Datasets[0].Data) {
			comprehensiveCharts[0].Datasets[0].Data[i].IsAnomaly = anomalyCharts[0].Datasets[0].Data[i].IsAnomaly
		}
	}

	for i := range comprehensiveCharts[1].Datasets[0].Data {
		if i < len(anomalyCharts[1].Datasets[0].Data) {
			comprehensiveCharts[1].Datasets[0].Data[i].IsAnomaly = anomalyCharts[1].Datasets[0].Data[i].IsAnomaly
		}
	}

	return comprehensiveCharts
}

// 根据多个环境因素计算空气质量指数（示例实现）
func calculateAQI(entry models.DataEntry) float64 {
	// 简化的计算方式，实际应基于标准算法
	return float64(entry.FreshAir*0.3 + entry.Ozone*0.1 + entry.NitroDio*0.1 +
		entry.Methanal*0.1 + entry.Pm25*0.2 + entry.CarbMomo*0.1 +
		entry.Bacteria*0.05 + entry.Radon*0.05)
}

// 获取空气质量级别
func getAQILevel(aqi float64) string {
	if aqi < 50 {
		return "优"
	} else if aqi < 100 {
		return "良"
	} else if aqi < 150 {
		return "轻度污染"
	} else if aqi < 200 {
		return "中度污染"
	} else if aqi < 300 {
		return "重度污染"
	} else {
		return "严重污染"
	}
}

// 生成详细分析项
func generateDetailedAnalysisItems(data []models.DataEntry, reportType string) []models.DetailedAnalysisItem {
	if len(data) == 0 {
		return []models.DetailedAnalysisItem{}
	}

	items := []models.DetailedAnalysisItem{}

	// 计算平均温度
	var avgTemp float64
	for _, entry := range data {
		avgTemp += float64(entry.Temperature)
	}
	avgTemp /= float64(len(data))

	items = append(items, models.DetailedAnalysisItem{
		Metric:      "平均温度",
		Value:       avgTemp,
		Status:      getTemperatureStatus(avgTemp),
		Description: getTemperatureDescription(avgTemp),
	})

	// 计算平均湿度
	var avgHumidity float64
	for _, entry := range data {
		avgHumidity += float64(entry.Humidity)
	}
	avgHumidity /= float64(len(data))

	items = append(items, models.DetailedAnalysisItem{
		Metric:      "平均湿度",
		Value:       avgHumidity,
		Status:      getHumidityStatus(avgHumidity),
		Description: getHumidityDescription(avgHumidity),
	})

	// 计算平均PM2.5
	var avgPM25 float64
	for _, entry := range data {
		avgPM25 += float64(entry.Pm25)
	}
	avgPM25 /= float64(len(data))

	items = append(items, models.DetailedAnalysisItem{
		Metric:      "平均PM2.5",
		Value:       avgPM25,
		Status:      getPM25Status(avgPM25),
		Description: getPM25Description(avgPM25),
	})

	// 计算平均臭氧
	var avgOzone float64
	for _, entry := range data {
		avgOzone += float64(entry.Ozone)
	}
	avgOzone /= float64(len(data))

	items = append(items, models.DetailedAnalysisItem{
		Metric:      "平均臭氧",
		Value:       avgOzone,
		Status:      getOzoneStatus(avgOzone),
		Description: getOzoneDescription(avgOzone),
	})

	// 计算平均二氧化氮
	var avgNitroDio float64
	for _, entry := range data {
		avgNitroDio += float64(entry.NitroDio)
	}
	avgNitroDio /= float64(len(data))

	items = append(items, models.DetailedAnalysisItem{
		Metric:      "平均二氧化氮",
		Value:       avgNitroDio,
		Status:      getNitroDioStatus(avgNitroDio),
		Description: getNitroDioDescription(avgNitroDio),
	})

	// 根据报告类型添加额外的详细分析
	switch reportType {
	case "anomaly":
		// 添加异常统计
		tempAnomalyCount := 0
		humidityAnomalyCount := 0

		for _, entry := range data {
			if entry.Temperature > 30 || entry.Temperature < 10 {
				tempAnomalyCount++
			}
			if entry.Humidity < 20 || entry.Humidity > 70 {
				humidityAnomalyCount++
			}
		}

		// 使用if-else替代三元运算符
		var tempStatus string
		if tempAnomalyCount > 0 {
			tempStatus = "异常"
		} else {
			tempStatus = "正常"
		}

		items = append(items, models.DetailedAnalysisItem{
			Metric:      "温度异常次数",
			Value:       float64(tempAnomalyCount),
			Status:      tempStatus,
			Description: fmt.Sprintf("在%d个数据点中发现%d次温度异常", len(data), tempAnomalyCount),
		})

		var humidityStatus string
		if humidityAnomalyCount > 0 {
			humidityStatus = "异常"
		} else {
			humidityStatus = "正常"
		}

		items = append(items, models.DetailedAnalysisItem{
			Metric:      "湿度异常次数",
			Value:       float64(humidityAnomalyCount),
			Status:      humidityStatus,
			Description: fmt.Sprintf("在%d个数据点中发现%d次湿度异常", len(data), humidityAnomalyCount),
		})

	case "trend":
		// 添加趋势统计
		if len(data) > 1 {
			tempChange := float64(data[len(data)-1].Temperature - data[0].Temperature)
			humidityChange := float64(data[len(data)-1].Humidity - data[0].Humidity)

			// 温度趋势描述使用if-else
			var tempTrendDirection string
			if tempChange > 0 {
				tempTrendDirection = "上升"
			} else {
				tempTrendDirection = "下降"
			}

			items = append(items, models.DetailedAnalysisItem{
				Metric:      "温度变化趋势",
				Value:       tempChange,
				Status:      getTrendStatus(tempChange),
				Description: fmt.Sprintf("在观测期间，温度总体%s了%.1f度", tempTrendDirection, abs(tempChange)),
			})

			// 湿度趋势描述使用if-else
			var humidityTrendDirection string
			if humidityChange > 0 {
				humidityTrendDirection = "上升"
			} else {
				humidityTrendDirection = "下降"
			}

			items = append(items, models.DetailedAnalysisItem{
				Metric:      "湿度变化趋势",
				Value:       humidityChange,
				Status:      getTrendStatus(humidityChange),
				Description: fmt.Sprintf("在观测期间，湿度总体%s了%.1f%%", humidityTrendDirection, abs(humidityChange)),
			})
		}

	case "comprehensive":
		// 添加空气质量指数
		var avgAQI float64
		for _, entry := range data {
			avgAQI += calculateAQI(entry)
		}
		avgAQI /= float64(len(data))

		items = append(items, models.DetailedAnalysisItem{
			Metric:      "平均空气质量指数",
			Value:       avgAQI,
			Status:      getAQILevel(avgAQI),
			Description: fmt.Sprintf("空气质量为%s，%s", getAQILevel(avgAQI), getAQIDescription(avgAQI)),
		})

		// 添加PM2.5分析
		var avgPM25 float64
		for _, entry := range data {
			avgPM25 += float64(entry.Pm25)
		}
		avgPM25 /= float64(len(data))

		items = append(items, models.DetailedAnalysisItem{
			Metric:      "平均PM2.5",
			Value:       avgPM25,
			Status:      getPM25Status(avgPM25),
			Description: getPM25Description(avgPM25),
		})

		// 添加甲醛分析
		var avgMethanal float64
		for _, entry := range data {
			avgMethanal += float64(entry.Methanal)
		}
		avgMethanal /= float64(len(data))

		items = append(items, models.DetailedAnalysisItem{
			Metric:      "平均甲醛",
			Value:       avgMethanal,
			Status:      getMethanolStatus(avgMethanal),
			Description: getMethanolDescription(avgMethanal),
		})

		// 添加一氧化碳分析
		var avgCarbMomo float64
		for _, entry := range data {
			avgCarbMomo += float64(entry.CarbMomo)
		}
		avgCarbMomo /= float64(len(data))

		items = append(items, models.DetailedAnalysisItem{
			Metric:      "平均一氧化碳",
			Value:       avgCarbMomo,
			Status:      getCarbMomoStatus(avgCarbMomo),
			Description: getCarbMomoDescription(avgCarbMomo),
		})

		// 添加细菌分析
		var avgBacteria float64
		for _, entry := range data {
			avgBacteria += float64(entry.Bacteria)
		}
		avgBacteria /= float64(len(data))

		items = append(items, models.DetailedAnalysisItem{
			Metric:      "平均细菌浓度",
			Value:       avgBacteria,
			Status:      getBacteriaStatus(avgBacteria),
			Description: getBacteriaDescription(avgBacteria),
		})

		// 添加氡气分析
		var avgRadon float64
		for _, entry := range data {
			avgRadon += float64(entry.Radon)
		}
		avgRadon /= float64(len(data))

		items = append(items, models.DetailedAnalysisItem{
			Metric:      "平均氡气",
			Value:       avgRadon,
			Status:      getRadonStatus(avgRadon),
			Description: getRadonDescription(avgRadon),
		})
	}

	return items
}

// 获取温度状态评价
func getTemperatureStatus(temp float64) string {
	if temp < 10 {
		return "偏低"
	} else if temp > 30 {
		return "偏高"
	} else {
		return "正常"
	}
}

// 获取温度描述
func getTemperatureDescription(temp float64) string {
	if temp < 10 {
		return "温度过低，可能影响舒适度"
	} else if temp > 30 {
		return "温度过高，可能影响设备运行和人员舒适度"
	} else if temp >= 20 && temp <= 26 {
		return "温度适宜，符合舒适标准"
	} else {
		return "温度在可接受范围内"
	}
}

// 获取湿度状态评价
func getHumidityStatus(humidity float64) string {
	if humidity < 20 {
		return "过干"
	} else if humidity > 70 {
		return "过湿"
	} else {
		return "正常"
	}
}

// 获取湿度描述
func getHumidityDescription(humidity float64) string {
	if humidity < 20 {
		return "湿度过低，可能导致静电增加和干燥不适"
	} else if humidity > 70 {
		return "湿度过高，可能导致霉菌生长和设备受潮"
	} else if humidity >= 40 && humidity <= 60 {
		return "湿度适宜，符合舒适标准"
	} else {
		return "湿度在可接受范围内"
	}
}

// 获取趋势状态
func getTrendStatus(change float64) string {
	if change > 5 {
		return "显著上升"
	} else if change > 0 {
		return "轻微上升"
	} else if change < -5 {
		return "显著下降"
	} else if change < 0 {
		return "轻微下降"
	} else {
		return "稳定"
	}
}

// 获取空气质量描述
func getAQIDescription(aqi float64) string {
	if aqi < 50 {
		return "空气质量优秀，适合户外活动"
	} else if aqi < 100 {
		return "空气质量良好，可以正常进行户外活动"
	} else if aqi < 150 {
		return "敏感人群应减少户外活动"
	} else if aqi < 200 {
		return "所有人应减少户外活动"
	} else if aqi < 300 {
		return "所有人应避免户外活动"
	} else {
		return "所有人应停止户外活动，并建议佩戴防护口罩"
	}
}

// 获取PM2.5状态
func getPM25Status(pm25 float64) string {
	if pm25 < 35 {
		return "优"
	} else if pm25 < 75 {
		return "良"
	} else if pm25 < 115 {
		return "轻度污染"
	} else if pm25 < 150 {
		return "中度污染"
	} else if pm25 < 250 {
		return "重度污染"
	} else {
		return "严重污染"
	}
}

// 获取PM2.5描述
func getPM25Description(pm25 float64) string {
	if pm25 < 35 {
		return "PM2.5浓度优良，空气清新"
	} else if pm25 < 75 {
		return "PM2.5浓度良好，空气质量可接受"
	} else if pm25 < 115 {
		return "PM2.5轻度污染，敏感人群应减少户外活动"
	} else if pm25 < 150 {
		return "PM2.5中度污染，应减少户外活动时间"
	} else if pm25 < 250 {
		return "PM2.5重度污染，应避免户外活动"
	} else {
		return "PM2.5严重污染，建议佩戴口罩或使用空气净化器"
	}
}

// 获取臭氧状态
func getOzoneStatus(ozone float64) string {
	if ozone < 0.05 {
		return "优"
	} else if ozone < 0.1 {
		return "良"
	} else if ozone < 0.2 {
		return "轻度污染"
	} else {
		return "严重污染"
	}
}

// 获取臭氧描述
func getOzoneDescription(ozone float64) string {
	if ozone < 0.05 {
		return "臭氧浓度低，空气质量优良"
	} else if ozone < 0.1 {
		return "臭氧浓度在正常范围内"
	} else if ozone < 0.2 {
		return "臭氧浓度偏高，可能对呼吸系统造成影响"
	} else {
		return "臭氧浓度严重超标，建议减少户外活动"
	}
}

// 获取二氧化氮状态
func getNitroDioStatus(nitroDio float64) string {
	if nitroDio < 40 {
		return "优"
	} else if nitroDio < 80 {
		return "良"
	} else if nitroDio < 180 {
		return "轻度污染"
	} else {
		return "严重污染"
	}
}

// 获取二氧化氮描述
func getNitroDioDescription(nitroDio float64) string {
	if nitroDio < 40 {
		return "二氧化氮浓度低，空气质量优良"
	} else if nitroDio < 80 {
		return "二氧化氮浓度在正常范围内"
	} else if nitroDio < 180 {
		return "二氧化氮浓度偏高，可能对呼吸系统造成影响"
	} else {
		return "二氧化氮浓度严重超标，建议减少户外活动"
	}
}

// 获取甲醛状态
func getMethanolStatus(methanol float64) string {
	if methanol < 0.08 {
		return "优"
	} else if methanol < 0.1 {
		return "良"
	} else {
		return "超标"
	}
}

// 获取甲醛描述
func getMethanolDescription(methanol float64) string {
	if methanol < 0.08 {
		return "甲醛浓度低，空气质量优良"
	} else if methanol < 0.1 {
		return "甲醛浓度在正常范围内"
	} else {
		return "甲醛浓度超标，建议检查室内装修材料和通风系统"
	}
}

// 获取一氧化碳状态
func getCarbMomoStatus(carbMomo float64) string {
	if carbMomo < 2 {
		return "优"
	} else if carbMomo < 4 {
		return "良"
	} else if carbMomo < 10 {
		return "超标"
	} else {
		return "危险"
	}
}

// 获取一氧化碳描述
func getCarbMomoDescription(carbMomo float64) string {
	if carbMomo < 2 {
		return "一氧化碳浓度低，空气质量优良"
	} else if carbMomo < 4 {
		return "一氧化碳浓度在正常范围内"
	} else if carbMomo < 10 {
		return "一氧化碳浓度超标，建议检查燃气设备"
	} else {
		return "一氧化碳浓度危险，立即通风并寻求专业帮助"
	}
}

// 获取细菌状态
func getBacteriaStatus(bacteria float64) string {
	if bacteria < 500 {
		return "优"
	} else if bacteria < 1000 {
		return "良"
	} else {
		return "超标"
	}
}

// 获取细菌描述
func getBacteriaDescription(bacteria float64) string {
	if bacteria < 500 {
		return "细菌浓度低，空气质量优良"
	} else if bacteria < 1000 {
		return "细菌浓度在正常范围内"
	} else {
		return "细菌浓度超标，建议改善通风条件并进行消毒"
	}
}

// 获取氡气状态
func getRadonStatus(radon float64) string {
	if radon < 100 {
		return "优"
	} else if radon < 200 {
		return "良"
	} else if radon < 400 {
		return "超标"
	} else {
		return "危险"
	}
}

// 获取氡气描述
func getRadonDescription(radon float64) string {
	if radon < 100 {
		return "氡气浓度低，空气质量优良"
	} else if radon < 200 {
		return "氡气浓度在正常范围内"
	} else if radon < 400 {
		return "氡气浓度超标，建议改善通风条件"
	} else {
		return "氡气浓度危险，建议专业检测并采取治理措施"
	}
}

// 取绝对值
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
