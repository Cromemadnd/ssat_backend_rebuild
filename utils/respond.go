package utils

import (
	"github.com/gin-gonic/gin"
)

func Respond(c *gin.Context, data any, status ErrorCode) {
	c.AbortWithStatusJSON(status.HttpCode, gin.H{
		"status":  status.Code,
		"message": status.Message,
		"data":    data,
	})
}

// RespondWithError 返回错误响应
func RespondWithError(c *gin.Context, httpCode int, message string, err error) {
	errResponse := gin.H{
		"error":   message,
		"message": err.Error(),
	}
	c.AbortWithStatusJSON(httpCode, errResponse)
}
