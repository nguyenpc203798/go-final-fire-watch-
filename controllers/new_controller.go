// controllers/new_controller.go
package controllers

import (
	"context"
	"encoding/json"
	"fire-watch/dbs"
	"fire-watch/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Khai báo biến collection cho new
var newCollection *mongo.Collection

// Khởi tạo newCollection
func InitializeNewCollection() {
	if dbs.DB == nil {
		log.Fatal("Database not initialized")
	}
	newCollection = dbs.DB.Collection("news")
}

// Thêm new mới
func Addnew(c *gin.Context) {
	var new models.New
	if err := c.ShouldBindJSON(&new); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid new data"})
		return
	}

	new.ID = primitive.NewObjectID()
	new.Status = 1 
	new.CreatedAt = time.Now()
	new.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := newCollection.InsertOne(ctx, new)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting new"})
		return
	}

	// Xóa cache Redis sau khi thêm new mới
	dbs.RedisClient.Del(ctx, "news")

	c.JSON(http.StatusOK, new)
}

// Lấy tất cả news
func GetAllNews(c *gin.Context) {
	var news []models.New

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kiểm tra dữ liệu cache từ Redis
	cachednews, err := dbs.RedisClient.Get(ctx, "news").Result()
	if err == nil && cachednews != "" {
		// Nếu có cache, giải mã từ Redis cache
		json.Unmarshal([]byte(cachednews), &news)
		c.JSON(http.StatusOK, news)
		return
	}

	// Nếu không có cache, lấy từ MongoDB
	cursor, err := newCollection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching news"})
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var new models.New
		if err := cursor.Decode(&new); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding new"})
			return
		}
		news = append(news, new)
	}

	// Lưu dữ liệu vào Redis cache
	newsJSON, _ := json.Marshal(news)
	dbs.RedisClient.Set(ctx, "news", string(newsJSON), 30*time.Minute)

	c.JSON(http.StatusOK, news)
}

// Lấy một new theo ID
func GetNewByID(c *gin.Context) {
	id := c.Query("id")
	newID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid new ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kiểm tra cache từ Redis
	cachednew, err := dbs.RedisClient.Get(ctx, "new_"+id).Result()
	if err == nil && cachednew != "" {
		// Nếu có cache, trả về cache
		var new models.New
		json.Unmarshal([]byte(cachednew), &new)
		c.JSON(http.StatusOK, new)
		return
	}

	// Nếu không có cache, lấy từ MongoDB
	var new models.New
	err = newCollection.FindOne(ctx, bson.M{"_id": newID}).Decode(&new)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "new not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching new"})
		}
		return
	}

	// Lưu dữ liệu vào Redis cache
	newJSON, _ := json.Marshal(new)
	dbs.RedisClient.Set(ctx, "new_"+id, string(newJSON), 30*time.Minute)

	c.JSON(http.StatusOK, new)
}

// Cập nhật new
func UpdateNew(c *gin.Context) {
	var new models.New
	if err := c.ShouldBindJSON(&new); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid new data"})
		return
	}
	new.UpdatedAt = time.Now()
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": new.ID}
	update := bson.M{"$set": new}

	_, err := newCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating new"})
		return
	}

	// Xóa cache Redis sau khi cập nhật
	dbs.RedisClient.Del(ctx, "new_"+new.ID.Hex())
	dbs.RedisClient.Del(ctx, "news")

	c.JSON(http.StatusOK, gin.H{"message": "new updated successfully"})
}

// Xóa new
func DeleteNew(c *gin.Context) {
	id := c.Query("id")
	newID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid new ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = newCollection.DeleteOne(ctx, bson.M{"_id": newID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting new"})
		return
	}

	// Xóa cache Redis sau khi xóa
	dbs.RedisClient.Del(ctx, "new_"+id)
	dbs.RedisClient.Del(ctx, "news")

	c.JSON(http.StatusOK, gin.H{"message": "new deleted successfully"})
}
