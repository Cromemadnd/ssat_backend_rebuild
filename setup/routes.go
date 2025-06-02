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
	// 添加 CORS 中间件
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		// c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 初始化处理器
	deviceHandler := &handlers.DeviceHandler{
		BaseHandler: handlers.BaseHandler[models.Device]{DB: db},
	}
	userHandler := &handlers.UserHandler{
		BaseHandler: handlers.BaseHandler[models.User]{DB: db},
	}
	dataHandler := &handlers.DataHandler{
		MongoCollection:     dbMongo,
		MongoToSQLThreshold: config.MongoToSQLThreshold,
		AiApiUrl:            config.AiApiUrl,
		AiApiKey:            config.AiApiKey,
		BaseHandler:         handlers.BaseHandler[models.Data]{DB: db},
	}
	logHandler := &handlers.LogHandler{
		BaseHandler: handlers.BaseHandler[models.Log]{DB: db},
	}
	announcementHandler := &handlers.AnnouncementHandler{
		BaseHandler: handlers.BaseHandler[models.Announcement]{DB: db},
	}
	ticketHandler := &handlers.TicketHandler{
		BaseHandler: handlers.BaseHandler[models.Ticket]{DB: db},
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
	logMiddleware := &middlewares.LogMiddleware{DB: db}

	apiRouter := router.Group("/")
	{
		auth := apiRouter.Group("/auth")
		{
			auth.POST("/login", authHandler.AdminLogin)
			auth.POST("/wechat_login", authHandler.WechatLogin)
		}

		devices := apiRouter.Group("/devices")
		{
			// 只允许普通用户访问
			devices.GET("/my_devices", authMiddleware.UserOnly(), deviceHandler.MyDevices)
			devices.GET("/my_devices/:uuid", authMiddleware.UserOnly(), deviceHandler.RetrieveMyDevice)
			devices.POST("/:uuid/bind", authMiddleware.UserOnly(), logMiddleware.WithLogging(1), deviceHandler.Bind)
			devices.POST("/:uuid/unbind", authMiddleware.UserOnly(), logMiddleware.WithLogging(1), deviceHandler.Unbind)

			// 只允许管理员访问
			devices.GET("/", authMiddleware.AdminOnly(), deviceHandler.List)
			devices.GET("/:uuid", authMiddleware.AdminOnly(), deviceHandler.Retrieve)
			devices.POST("/", authMiddleware.AdminOnly(), logMiddleware.WithLogging(2), deviceHandler.Create)
			devices.PUT("/:uuid", authMiddleware.AdminOnly(), logMiddleware.WithLogging(2), deviceHandler.Update)
			devices.DELETE("/:uuid", authMiddleware.AdminOnly(), logMiddleware.WithLogging(2), deviceHandler.Destroy)
		}

		users := apiRouter.Group("/users")
		{
			// 只允许普通用户访问
			users.GET("/my_profile", authMiddleware.UserOnly(), userHandler.MyProfile)

			// 只允许管理员访问
			users.GET("/", authMiddleware.AdminOnly(), userHandler.List)
			users.GET("/:uuid", authMiddleware.AdminOnly(), userHandler.Retrieve)
			users.DELETE("/:uuid", authMiddleware.AdminOnly(), logMiddleware.WithLogging(2), userHandler.Destroy)
		}

		data := apiRouter.Group("/data")
		{
			data.POST("/upload", logMiddleware.WithLogging(0), dataHandler.Upload)

			data.GET("/", authMiddleware.AdminOnly(), dataHandler.List)
			data.POST("/analysis", authMiddleware.AdminOnly(), logMiddleware.WithLogging(2), dataHandler.Analysis)
		}

		logs := apiRouter.Group("/logs")
		{
			logs.GET("/", authMiddleware.AdminOnly(), logHandler.List)
			// logs.GET("/:uuid", authMiddleware.AdminOnly(), logHandler.Retrieve)
		}

		announcements := apiRouter.Group("/announcements")
		{
			// 允许普通用户或者管理员访问
			announcements.GET("/", authMiddleware.UserOrAdmin(), announcementHandler.List)
			announcements.GET("/:uuid", authMiddleware.UserOrAdmin(), announcementHandler.Retrieve)

			// 只允许管理员访问
			announcements.POST("/", authMiddleware.AdminOnly(), logMiddleware.WithLogging(2), announcementHandler.Create)
			announcements.PUT("/:uuid", authMiddleware.AdminOnly(), logMiddleware.WithLogging(2), announcementHandler.Update)
			announcements.DELETE("/:uuid", authMiddleware.AdminOnly(), logMiddleware.WithLogging(2), announcementHandler.Destroy)
		}

		tickets := apiRouter.Group("/tickets")
		{
			// 允许普通用户访问
			tickets.GET("/my_tickets", authMiddleware.UserOnly(), ticketHandler.ListMyTickets)
			tickets.GET("/my_tickets/:uuid", authMiddleware.UserOnly(), ticketHandler.RetrieveMyTicket)
			tickets.POST("/", authMiddleware.UserOnly(), logMiddleware.WithLogging(1), ticketHandler.Create)
			tickets.POST("/:uuid/supply", authMiddleware.UserOnly(), logMiddleware.WithLogging(1), ticketHandler.Supply)
			tickets.POST("/:uuid/close", authMiddleware.UserOnly(), logMiddleware.WithLogging(1), ticketHandler.Close)

			// 只允许管理员访问
			tickets.GET("/", authMiddleware.AdminOnly(), ticketHandler.List)
			tickets.GET("/:uuid", authMiddleware.AdminOnly(), ticketHandler.Retrieve)
			tickets.POST("/:uuid/reply", authMiddleware.AdminOnly(), logMiddleware.WithLogging(2), ticketHandler.Reply)
		}
	}
}
