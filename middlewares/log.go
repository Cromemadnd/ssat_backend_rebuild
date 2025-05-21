package middlewares

import (
	"ssat_backend_rebuild/models"
	"ssat_backend_rebuild/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LogMiddleware struct {
	DB *gorm.DB
}

func (m *LogMiddleware) WithLogging(logType uint8) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		subject := &uuid.UUID{}
		switch logType {
		case 0:
			subject = &c.MustGet("CurrentDevice").(*models.Device).UUID
		case 1:
			subject = &c.MustGet("CurrentUser").(*models.User).UUID
		case 2:
			subject = &c.MustGet("CurrentAdminUser").(*models.Admin).UUID
		}

		logEntry := models.Log{
			LogType: logType,
			Subject: subject,
			Path:    c.Request.URL.Path,
			Method:  c.Request.Method,
			IP:      c.ClientIP(),
			Status:  c.MustGet("status").(*utils.ErrorCode),
		}
		m.DB.Create(&logEntry)
	}
}
