package setup

import (
	"ssat_backend_rebuild/handlers"
	"ssat_backend_rebuild/middlewares"
	"ssat_backend_rebuild/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB, dbMongo *mongo.Collection, config Config) {
	// 初始化处理器
	deviceHandler := &handlers.DeviceHandler{
		BaseHandler: handlers.BaseHandler[models.Device]{DB: db},
	}
	userHandler := &handlers.UserHandler{
		BaseHandler: handlers.BaseHandler[models.User]{DB: db},
	}
	dataHandler := &handlers.DataHandler{
		MongoCollection: dbMongo,
		BaseHandler:     handlers.BaseHandler[models.Data]{DB: db},
	}

	// userHandler := &handlers.BaseHandler[models.User]{DB: db}
	authHandler := &handlers.AuthHandler{
		DB:           db,
		JWTSecret:    config.JWTConfig.Secret,
		JWTExpires:   config.JWTConfig.Expires,
		WechatAppID:  config.WechatConfig.AppID,
		WechatSecret: config.WechatConfig.Secret,
	}
	authMiddleware := &middlewares.AuthMiddleware{
		DB:        db,
		JWTSecret: config.JWTConfig.Secret,
	}

	apiRouter := router.Group("/")
	{
		auth := apiRouter.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authMiddleware.AuthRequired(), authHandler.Logout)
			auth.POST("/register", authHandler.Register)
			auth.POST("/wechat_login", authHandler.WechatLogin)
		}

		devices := apiRouter.Group("/devices")
		devices.Use(authMiddleware.AuthRequired())
		{
			devices.GET("/my_devices", deviceHandler.MyDevices)
			devices.GET("/my_devices/:uuid", deviceHandler.RetrieveMyDevice)
			devices.POST("/:uuid/bind", deviceHandler.Bind)
			devices.POST("/:uuid/unbind", deviceHandler.Unbind)

			// 需要权限控制的设备操作
			devices.Use(authMiddleware.AdminOnly())
			{
				devices.GET("/", deviceHandler.List)
				devices.GET("/:uuid", deviceHandler.Retrieve)
				devices.POST("/", deviceHandler.Create)
				devices.PUT("/:uuid", deviceHandler.Update)
				devices.DELETE("/:uuid", deviceHandler.Destroy)
			}
		}

		users := apiRouter.Group("/users")
		users.Use(authMiddleware.AuthRequired())
		{
			users.GET("/my_profile", userHandler.MyProfile)
			users.Use(authMiddleware.AdminOnly())
			{
				users.GET("/", userHandler.List)
				users.GET("/:uuid", userHandler.Retrieve)
				users.PUT("/:uuid", userHandler.Update)
				users.DELETE("/:uuid", userHandler.Destroy)
			}
		}

		data := apiRouter.Group("/data")
		{
			data.POST("/upload", dataHandler.Upload)
			data.Use(authMiddleware.AuthRequired())
			{

			}
		}
	}
}
