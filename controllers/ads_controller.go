// controllers/ads_controller.go
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

// Khai báo biến collection cho ads
var adsCollection *mongo.Collection

// Khởi tạo adsCollection
func InitializeadsCollection() {
	if dbs.DB == nil {
		log.Fatal("Database not initialized")
	}
	adsCollection = dbs.DB.Collection("adss")
}

// Thêm ads mới
func Addads(c *gin.Context) {
	var ads models.Ads
	if err := c.ShouldBindJSON(&ads); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ads data"})
		return
	}

	ads.ID = primitive.NewObjectID()
	ads.Status = 1 
	ads.CreatedAt = time.Now()
	ads.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := adsCollection.InsertOne(ctx, ads)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting ads"})
		return
	}

	// Xóa cache Redis sau khi thêm ads mới
	dbs.RedisClient.Del(ctx, "adss")

	c.JSON(http.StatusOK, ads)
}

// Lấy tất cả adss
func GetAlladss(c *gin.Context) {
	var adss []models.Ads

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kiểm tra dữ liệu cache từ Redis
	cachedadss, err := dbs.RedisClient.Get(ctx, "adss").Result()
	if err == nil && cachedadss != "" {
		// Nếu có cache, giải mã từ Redis cache
		json.Unmarshal([]byte(cachedadss), &adss)
		c.JSON(http.StatusOK, adss)
		return
	}

	// Nếu không có cache, lấy từ MongoDB
	cursor, err := adsCollection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching adss"})
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var ads models.Ads
		if err := cursor.Decode(&ads); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding ads"})
			return
		}
		adss = append(adss, ads)
	}

	// Lưu dữ liệu vào Redis cache
	adssJSON, _ := json.Marshal(adss)
	dbs.RedisClient.Set(ctx, "adss", string(adssJSON), 30*time.Minute)

	c.JSON(http.StatusOK, adss)
}

// Lấy một ads theo ID
func GetadsByID(c *gin.Context) {
	id := c.Query("id")
	adsID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ads ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kiểm tra cache từ Redis
	cachedads, err := dbs.RedisClient.Get(ctx, "ads_"+id).Result()
	if err == nil && cachedads != "" {
		// Nếu có cache, trả về cache
		var ads models.Ads
		json.Unmarshal([]byte(cachedads), &ads)
		c.JSON(http.StatusOK, ads)
		return
	}

	// Nếu không có cache, lấy từ MongoDB
	var ads models.Ads
	err = adsCollection.FindOne(ctx, bson.M{"_id": adsID}).Decode(&ads)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "ads not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching ads"})
		}
		return
	}

	// Lưu dữ liệu vào Redis cache
	adsJSON, _ := json.Marshal(ads)
	dbs.RedisClient.Set(ctx, "ads_"+id, string(adsJSON), 30*time.Minute)

	c.JSON(http.StatusOK, ads)
}

// Cập nhật ads
func Updateads(c *gin.Context) {
	var ads models.Ads
	if err := c.ShouldBindJSON(&ads); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ads data"})
		return
	}
	ads.UpdatedAt = time.Now()
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": ads.ID}
	update := bson.M{"$set": ads}

	_, err := adsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating ads"})
		return
	}

	// Xóa cache Redis sau khi cập nhật
	dbs.RedisClient.Del(ctx, "ads_"+ads.ID.Hex())
	dbs.RedisClient.Del(ctx, "adss")

	c.JSON(http.StatusOK, gin.H{"message": "ads updated successfully"})
}

// Xóa ads
func Deleteads(c *gin.Context) {
	id := c.Query("id")
	adsID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ads ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = adsCollection.DeleteOne(ctx, bson.M{"_id": adsID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting ads"})
		return
	}

	// Xóa cache Redis sau khi xóa
	dbs.RedisClient.Del(ctx, "ads_"+id)
	dbs.RedisClient.Del(ctx, "adss")

	c.JSON(http.StatusOK, gin.H{"message": "ads deleted successfully"})
}
