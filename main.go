package main

import (
	"log"
	"ssat_backend_rebuild/setup"

	"github.com/gin-gonic/gin"
)

// 连接数据库
func main() {
	// 连接数据库
	myConfig := setup.LoadConfig()
	db := setup.SetupSQL(myConfig.SQLConfig)
	dbMongo := setup.SetupMongo(myConfig.MongoConfig)

	// 设置Gin
	router := gin.Default()

	// 设置路由
	setup.SetupRoutes(router, db, dbMongo, myConfig)

	// 启动服务器
	if err := router.Run(myConfig.ServerAddr); err != nil {
		log.Fatal("服务器启动失败: ", err)
	}
}
