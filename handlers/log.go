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
		nil,
		func(c *gin.Context, query *gorm.DB) *gorm.DB {
			logType := c.Query("log_type")
			subject := c.Query("subject")
			before := c.Query("before")
			after := c.Query("after")

			if logType != "" {
				query = query.Where("log_type = ?", logType)
			}
			if subject != "" {
				query = query.Where("subject = ?", subject)
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

// 日志详情
func (h *LogHandler) Retrieve(c *gin.Context) {
	h.BaseHandler.Retrieve(
		nil,
		nil,
	)(c)
}
