package main

import (
	"context"
	"fmt"
	"mongodb-practice/config"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.uber.org/zap"
)

func main() {

	// zap loggerの設定
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("zapの設定に失敗しました: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync() // バッファをフラッシュする

	config.LoadEnv()

	mongoURI := config.GetMongoDB_URL()
	if mongoURI == "" {
		logger.Fatal("環境変数 MONGODB_URI が設定されていません")
	}

	// Use the SetServerAPIOptions() method to set the version of the Stable API on the client
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(mongoURI).SetServerAPIOptions(serverAPI)

	// クライアントを作成し、サーバーに接続
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		logger.Fatal("MongoDBへの接続に失敗しました", zap.Error(err))
	}

	// クライアントの切断を遅延実行
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			logger.Error("MongoDBからの切断に失敗しました", zap.Error(err))
		}
	}()

	// コネクションのタイムアウトを設定
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// pingコマンドを送信して接続を確認
	if err := client.Database("admin").RunCommand(ctx, bson.D{{"ping", 1}}).Err(); err != nil {
		logger.Fatal("MongoDBへのpingに失敗しました", zap.Error(err))
	}

	logger.Info("MongoDBへの接続に成功しました")
}
