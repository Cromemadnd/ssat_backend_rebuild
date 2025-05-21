package utils

import "github.com/gin-gonic/gin"

func Respond(c *gin.Context, data any, status ErrorCode) {
	c.Set("status", &status)
	c.AbortWithStatusJSON(status.HttpCode, gin.H{
		"status":  status.Code,
		"message": status.Message,
		"data":    data,
	})
}
