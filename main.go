package main

import (
	"fmt"
	"log"
	"ssat_backend_rebuild/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func connectDB(config DBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=%s&parseTime=True&loc=Local",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Charset)

	// 连接到MySQL服务器（不指定数据库）
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 检查并创建数据库
	createDBQuery := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET %s", config.DBName, config.Charset)
	if err := db.Exec(createDBQuery).Error; err != nil {
		return nil, err
	}

	// 重新连接到指定的数据库
	dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
		config.Charset)

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置最大空闲连接数
	sqlDB.SetMaxIdleConns(10)
	// 设置最大打开连接数
	sqlDB.SetMaxOpenConns(100)
	// 设置连接最大可复用时间
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// 连接数据库
func main() {
	// 连接数据库
	db, err := connectDB(dbConfig)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 测试连接
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("获取数据库实例失败: %v", err)
	}

	err = sqlDB.Ping()
	if err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	fmt.Println("数据库连接成功!")
	db.AutoMigrate(&models.Device{})

	// 设置Gin
	router := gin.Default()

	// 设置路由
	SetupRoutes(router, db)

	// 启动服务器
	if err := router.Run(":8080"); err != nil {
		log.Fatal("服务器启动失败: ", err)
	}
}
