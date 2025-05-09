package setup

import (
	"ssat_backend_rebuild/handlers"
	"ssat_backend_rebuild/middlewares"
	"ssat_backend_rebuild/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB, dbMongo *mongo.Client, config Config) {
	// 初始化处理器
	deviceHandler := &handlers.DeviceHandler{
		BaseHandler: handlers.BaseHandler[models.Device]{DB: db},
	}
	// userHandler := &handlers.BaseHandler[models.User]{DB: db}
	authHandler := &handlers.AuthHandler{
		DB:         db,
		JWTSecret:  config.JWTConfig.Secret,
		JWTExpires: config.JWTConfig.Expires,
	}
	authMiddleware := &middlewares.AuthMiddleware{
		DB:        db,
		JWTSecret: config.JWTConfig.Secret,
	}

	apiRouter := router.Group("/")
	{
		// 设备相关路由
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

		// users := apiRouter.Group("/users")
		// users.Use(authMiddleware.AuthRequired())
		// {
		// 	users.GET("/my_profile", userHandler.)
		// 	users.GET("/",
		// 		authMiddleware.AuthRequired(),
		// 		authMiddleware.PermRequired(utils.ReadUsers),
		// 		userHandler.GetUsers)
		// 	users.GET("/:uuid",
		// 		authMiddleware.AuthRequired(),
		// 		authMiddleware.PermRequired(utils.ReadUsers),
		// 		userHandler.GetUser)
		// 	users.PUT("/:uuid",
		// 		authMiddleware.AuthRequired(),
		// 		authMiddleware.PermRequired(utils.WriteUsers),
		// 		userHandler.UpdateUser)
		// 	users.DELETE("/:uuid",
		// 		authMiddleware.AuthRequired(),
		// 		authMiddleware.PermRequired(utils.WriteUsers),
		// 		userHandler.DeleteUser)
		// }

		auth := apiRouter.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authMiddleware.AuthRequired(), authHandler.Logout)
			auth.POST("/register", authHandler.Register)
		}
	}
}
