package main

import (
	"ssat_backend_rebuild/handlers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	// 初始化处理器
	deviceHandler := &handlers.DeviceHandler{DB: db}

	// API版本组
	apiRouter := router.Group("/api")
	{
		// 设备相关路由
		devices := apiRouter.Group("/devices")
		{
			devices.GET("/", deviceHandler.GetDevices)
			devices.GET("/:id", deviceHandler.GetDevice)
			devices.DELETE("/:id", deviceHandler.DeleteDevice)
		}
	}
}
