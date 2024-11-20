// dbs/mongo.go
package dbs

import (
	"context"
	"log"
	"os"
	"time"

	// Đảm bảo đã cài đặt thư viện này
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

func Connect() {

	// Sử dụng URI kết nối từ biến môi trường
	clientOptions := options.Client().ApplyURI(os.Getenv("DEV_DB_ADDR"))

	// Tạo MongoDB client
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal("Error creating MongoDB client:", err)
	}

	// Tạo context với timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Kết nối với MongoDB
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal("Error connecting to MongoDB:", err)
	}

	// Chọn database "fire-watch"
	DB = client.Database("fire-watch")
	log.Println("Connected to MongoDB, using database 'fire-watch'")
}
