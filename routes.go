package main

import (
	"ssat_backend_rebuild/handlers"
	"ssat_backend_rebuild/middlewares"
	"ssat_backend_rebuild/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB, config Config) {
	// 初始化处理器
	deviceHandler := &handlers.DeviceHandler{DB: db}
	userHandler := &handlers.UserHandler{DB: db}
	authHandler := &handlers.AuthHandler{
		DB:         db,
		JWTSecret:  config.JWTConfig.Secret,
		JWTExpires: config.JWTConfig.Expires,
	}
	authMiddleware := &middlewares.AuthMiddleware{
		DB:        db,
		JWTSecret: config.JWTConfig.Secret,
	}

	/*
		权限代码   写  读
		用户管理   7   6
		设备管理   5   4
		日志管理   3   2
		数据管理   1   0
	*/

	apiRouter := router.Group("/")
	{
		// 设备相关路由
		devices := apiRouter.Group("/device")
		{
			devices.GET("/my_devices", deviceHandler.GetMyDevices)
			devices.POST("/bind", deviceHandler.GetMyDevices)

			devices.GET("/", authMiddleware.AuthRequired(), deviceHandler.GetDevices)
			devices.POST("/", deviceHandler.CreateDevice)
			devices.GET("/:uuid", deviceHandler.GetDevice)
			devices.PUT("/:uuid", deviceHandler.UpdateDevice)
			devices.DELETE("/:uuid", deviceHandler.DeleteDevice)
		}

		users := apiRouter.Group("/user")
		{
			users.GET("/",
				authMiddleware.AuthRequired(),
				authMiddleware.PermRequired(utils.ReadUsers),
				userHandler.GetUsers)
		}

		auth := apiRouter.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authMiddleware.AuthRequired(), authHandler.Logout)
			auth.POST("/register", authHandler.Register)
		}
	}
}
