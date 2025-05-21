package models

import (
	"time"
)

// 分析请求模型
type AnalysisRequest struct {
	DeviceID   string `json:"device_id" binding:"required"`
	Model      string `json:"model" binding:"required,oneof=DeepSeek-V3 DeepSeek-R1"` 
	ReportType string `json:"report_type" binding:"required,oneof=simple anomaly trend comprehensive"`
	TimeStart  int64  `json:"time_start" binding:"required"`
	TimeEnd    int64  `json:"time_end" binding:"required"`
}

// 分析报告摘要
type AnalysisSummary struct {
	Status        string `json:"status"`
	AnomalyCount  int    `json:"anomaly_count"`
	Recommendation string `json:"recommendation"`
	Alert         bool   `json:"alert"`
	AlertMessage  string `json:"alert_message,omitempty"`
}

// 图表数据点
type ChartDataPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
	Label     string  `json:"label,omitempty"`
	IsAnomaly bool    `json:"is_anomaly,omitempty"`
}

// 图表数据集合
type ChartDataset struct {
	Label string           `json:"label"`
	Data  []ChartDataPoint `json:"data"`
}

// 图表数据
type ChartData struct {
	Type     string         `json:"type"`
	Datasets []ChartDataset `json:"datasets"`
}

// 详细分析项
type DetailedAnalysisItem struct {
	Metric      string  `json:"metric"`
	Value       float64 `json:"value"`
	Status      string  `json:"status"`
	Description string  `json:"description"`
}

// 分析结果模型
type AnalysisResult struct {
	UUID          string                `json:"uuid"`
	DeviceID      string                `json:"device_id"`
	Model         string                `json:"model"`
	ReportType    string                `json:"report_type"`
	GeneratedAt   time.Time             `json:"generated_at"`
	TimeRange     [2]int64              `json:"time_range"`
	Summary       AnalysisSummary       `json:"summary"`
	Charts        []ChartData           `json:"charts"`
	DetailedItems []DetailedAnalysisItem `json:"detailed_items"`
}

// 分析缓存键
type AnalysisCacheKey struct {
	DeviceID   string
	Model      string
	ReportType string
	TimeStart  int64
	TimeEnd    int64
}

func (k AnalysisCacheKey) String() string {
	return k.DeviceID + ":" + k.Model + ":" + k.ReportType + ":" + 
	       time.Unix(k.TimeStart, 0).Format("20060102") + ":" + 
	       time.Unix(k.TimeEnd, 0).Format("20060102")
} 