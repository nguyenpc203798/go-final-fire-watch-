// controllers/genre_controller.go
package controllers

import (
	"context"
	"encoding/json"
	"fire-watch/dbs"
	"fire-watch/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Thêm genre mới
func AddGenre(c *gin.Context) {
	genreCollection := models.GetGenreCollection()
	var genre models.Genre
	if err := c.ShouldBindJSON(&genre); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid genre data"})
		return
	}

	genre.ID = primitive.NewObjectID()
	genre.Status = 1

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := genreCollection.InsertOne(ctx, genre)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting genre"})
		return
	}

	// Xóa cache Redis sau khi thêm genre mới
	dbs.RedisClient.Del(ctx, "genres")

	c.JSON(http.StatusOK, genre)
}

// Lấy tất cả genres
func GetAllGenres(c *gin.Context) {
	genreCollection := models.GetGenreCollection()
	var genres []models.Genre

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kiểm tra dữ liệu cache từ Redis
	cachedgenres, err := dbs.RedisClient.Get(ctx, "genres").Result()
	if err == nil && cachedgenres != "" {
		// Nếu có cache, giải mã từ Redis cache
		json.Unmarshal([]byte(cachedgenres), &genres)
		c.JSON(http.StatusOK, genres)
		return
	}

	// Nếu không có cache, lấy từ MongoDB
	cursor, err := genreCollection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching genres"})
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var genre models.Genre
		if err := cursor.Decode(&genre); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding genre"})
			return
		}
		genres = append(genres, genre)
	}

	// Lưu dữ liệu vào Redis cache
	genresJSON, _ := json.Marshal(genres)
	dbs.RedisClient.Set(ctx, "genres", string(genresJSON), 30*time.Minute)

	c.JSON(http.StatusOK, genres)
}

// Lấy một genre theo ID
func GetGenreByID(c *gin.Context) {
	genreCollection := models.GetGenreCollection()
	id := c.Query("id")
	genreID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid genre ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kiểm tra cache từ Redis
	cachedgenre, err := dbs.RedisClient.Get(ctx, "genre_"+id).Result()
	if err == nil && cachedgenre != "" {
		// Nếu có cache, trả về cache
		var genre models.Genre
		json.Unmarshal([]byte(cachedgenre), &genre)
		c.JSON(http.StatusOK, genre)
		return
	}

	// Nếu không có cache, lấy từ MongoDB
	var genre models.Genre
	err = genreCollection.FindOne(ctx, bson.M{"_id": genreID}).Decode(&genre)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "genre not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching genre"})
		}
		return
	}

	// Lưu dữ liệu vào Redis cache
	genreJSON, _ := json.Marshal(genre)
	dbs.RedisClient.Set(ctx, "genre_"+id, string(genreJSON), 30*time.Minute)

	c.JSON(http.StatusOK, genre)
}

// Cập nhật genre
func UpdateGenre(c *gin.Context) {
	genreCollection := models.GetGenreCollection()
	var genre models.Genre
	if err := c.ShouldBindJSON(&genre); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid genre data"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": genre.ID}
	update := bson.M{"$set": genre}

	_, err := genreCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating genre"})
		return
	}

	// Xóa cache Redis sau khi cập nhật
	dbs.RedisClient.Del(ctx, "genre_"+genre.ID.Hex())
	dbs.RedisClient.Del(ctx, "genres")

	c.JSON(http.StatusOK, gin.H{"message": "genre updated successfully"})
}

// Xóa genre
func DeleteGenre(c *gin.Context) {
	genreCollection := models.GetGenreCollection()
	movieCollection := models.GetMovieCollection()
	id := c.Param("id")
	genreID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid genre ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = genreCollection.DeleteOne(ctx, bson.M{"_id": genreID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting genre"})
		return
	}

	// Cập nhật tất cả các bộ phim, xóa genreID khỏi mảng genre của Movie
	_, err = movieCollection.UpdateMany(
		ctx,
		bson.M{"genre": genreID}, // Tìm những bộ phim chứa genreID
		bson.M{"$pull": bson.M{"genre": genreID}}, // Xóa genreID khỏi mảng genre
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating movies"})
		return
	}

	// Xóa cache Redis sau khi xóa
	dbs.RedisClient.Del(ctx, "genre_"+id)
	dbs.RedisClient.Del(ctx, "genres")
	dbs.RedisClient.Del(ctx, "movies_cache")

	c.JSON(http.StatusOK, gin.H{"message": "genre deleted successfully"})
}
