// controllers/slide_controller.go
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

// Khai báo biến collection cho slide
var slideCollection *mongo.Collection

// Khởi tạo slideCollection
func InitializeslideCollection() {
	if dbs.DB == nil {
		log.Fatal("Database not initialized")
	}
	slideCollection = dbs.DB.Collection("slides")
}

// Thêm slide mới
func Addslide(c *gin.Context) {
	var slide models.Slide
	if err := c.ShouldBindJSON(&slide); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slide data"})
		return
	}

	slide.ID = primitive.NewObjectID()
	slide.Status = 1 
	slide.CreatedAt = time.Now()
	slide.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := slideCollection.InsertOne(ctx, slide)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting slide"})
		return
	}

	// Xóa cache Redis sau khi thêm slide mới
	dbs.RedisClient.Del(ctx, "slides")

	c.JSON(http.StatusOK, slide)
}

// Lấy tất cả slides
func GetAllslides(c *gin.Context) {
	var slides []models.Slide

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kiểm tra dữ liệu cache từ Redis
	cachedslides, err := dbs.RedisClient.Get(ctx, "slides").Result()
	if err == nil && cachedslides != "" {
		// Nếu có cache, giải mã từ Redis cache
		json.Unmarshal([]byte(cachedslides), &slides)
		c.JSON(http.StatusOK, slides)
		return
	}

	// Nếu không có cache, lấy từ MongoDB
	cursor, err := slideCollection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching slides"})
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var slide models.Slide
		if err := cursor.Decode(&slide); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding slide"})
			return
		}
		slides = append(slides, slide)
	}

	// Lưu dữ liệu vào Redis cache
	slidesJSON, _ := json.Marshal(slides)
	dbs.RedisClient.Set(ctx, "slides", string(slidesJSON), 30*time.Minute)

	c.JSON(http.StatusOK, slides)
}

// Lấy một slide theo ID
func GetslideByID(c *gin.Context) {
	id := c.Query("id")
	slideID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slide ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kiểm tra cache từ Redis
	cachedslide, err := dbs.RedisClient.Get(ctx, "slide_"+id).Result()
	if err == nil && cachedslide != "" {
		// Nếu có cache, trả về cache
		var slide models.Slide
		json.Unmarshal([]byte(cachedslide), &slide)
		c.JSON(http.StatusOK, slide)
		return
	}

	// Nếu không có cache, lấy từ MongoDB
	var slide models.Slide
	err = slideCollection.FindOne(ctx, bson.M{"_id": slideID}).Decode(&slide)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "slide not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching slide"})
		}
		return
	}

	// Lưu dữ liệu vào Redis cache
	slideJSON, _ := json.Marshal(slide)
	dbs.RedisClient.Set(ctx, "slide_"+id, string(slideJSON), 30*time.Minute)

	c.JSON(http.StatusOK, slide)
}

// Cập nhật slide
func Updateslide(c *gin.Context) {
	var slide models.Slide
	if err := c.ShouldBindJSON(&slide); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slide data"})
		return
	}
	slide.UpdatedAt = time.Now()
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": slide.ID}
	update := bson.M{"$set": slide}

	_, err := slideCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating slide"})
		return
	}

	// Xóa cache Redis sau khi cập nhật
	dbs.RedisClient.Del(ctx, "slide_"+slide.ID.Hex())
	dbs.RedisClient.Del(ctx, "slides")

	c.JSON(http.StatusOK, gin.H{"message": "slide updated successfully"})
}

// Xóa slide
func Deleteslide(c *gin.Context) {
	id := c.Query("id")
	slideID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slide ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = slideCollection.DeleteOne(ctx, bson.M{"_id": slideID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting slide"})
		return
	}

	// Xóa cache Redis sau khi xóa
	dbs.RedisClient.Del(ctx, "slide_"+id)
	dbs.RedisClient.Del(ctx, "slides")

	c.JSON(http.StatusOK, gin.H{"message": "slide deleted successfully"})
}
