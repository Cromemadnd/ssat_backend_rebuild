package setup

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SetupMongo(config MongoConfig) *mongo.Collection {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := fmt.Sprintf("mongodb://%s:%d", config.Host, config.Port)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri).SetConnectTimeout(10*time.Second))
	if err != nil {
		log.Fatalf("MongoDB连接失败: %v", err)
		return nil
	}
	// 测试连接
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("MongoDB连接测试失败: %v", err)
		return nil
	}

	return client.Database(config.DBName).Collection(config.Collection)
}
