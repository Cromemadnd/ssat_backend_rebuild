package setup

import (
	"fmt"
	"log"
	"ssat_backend_rebuild/models"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func connectSQL(config SQLConfig) (*gorm.DB, error) {
	// 连接到MySQL服务器（不指定数据库）
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=%s&parseTime=True&loc=Local",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Charset)
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
	dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
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

func SetupSQL(config SQLConfig) *gorm.DB {
	db, err := connectSQL(config)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
		return nil
	}

	// 测试连接
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("获取数据库实例失败: %v", err)
		return nil
	}

	err = sqlDB.Ping()
	if err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
		return nil
	}

	fmt.Println("数据库连接成功!")
	err = db.AutoMigrate(&models.Device{}, &models.Admin{}, &models.User{}, &models.Data{}, &models.Log{})
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
		return nil
	}

	return db
}
