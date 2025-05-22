package handlers

import (
	"ssat_backend_rebuild/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LogHandler struct {
	BaseHandler[models.Log]
}

// 日志列表（可分页、可按条件筛选，简单版）
func (h *LogHandler) List(c *gin.Context) {
	h.BaseHandler.List(
		[]string{"uuid", "log_type", "subject", "created_at", "path"},
		func(c *gin.Context, query *gorm.DB) *gorm.DB {
			logType := c.Query("log_type")
			subject := c.Query("subject")

			if logType != "" {
				query = query.Where("log_type = ?", logType)
			}
			if subject != "" {
				query = query.Where("subject = ?", subject)
			}
			return query
		},
	)(c)
}

// 日志详情
func (h *LogHandler) Retrieve(c *gin.Context) {
	h.BaseHandler.Retrieve(
		nil,
		nil,
	)(c)
}
