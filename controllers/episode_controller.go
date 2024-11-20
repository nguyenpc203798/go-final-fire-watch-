// controllers/episode_controller
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

// Khai báo biến collection cho episode
var episodeCollection *mongo.Collection

// Khởi tạo episodeCollection
func InitializeEpisodeCollection() {
	if dbs.DB == nil {
		log.Fatal("Database not initialized")
	}
	episodeCollection = dbs.DB.Collection("episodes")
}

// Thêm tập mới (Episode)
func AddEpisode(c *gin.Context) {
	movieCollection := models.GetMovieCollection()
	var episode models.Episode
	if err := c.ShouldBindJSON(&episode); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	episode.ID = primitive.NewObjectID()
	episode.CreatedAt = time.Now()
	episode.UpdatedAt = time.Now()
	episode.Status = 1

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := episodeCollection.InsertOne(ctx, episode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi thêm tập phim"})
		return
	}

	// Tự động thêm ID của tập phim mới vào trường Episode của model Movie
	_, err = movieCollection.UpdateOne(
		ctx,
		bson.M{"_id": episode.MovieID}, // Tìm phim có MovieID khớp với episode.MovieID
		bson.M{"$push": bson.M{"episode": episode.ID}}, // Thêm episode.ID vào trường Episode
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi cập nhật tập phim vào Movie"})
		log.Print(err)
		return
	}

	// Xóa cache Redis sau khi thêm tập mới
	dbs.RedisClient.Del(ctx, "episodes")
	dbs.RedisClient.Del(ctx, "movies_cache")

	c.JSON(http.StatusOK, episode)
}

// Lấy tất cả các tập phim (Episodes)
func GetAllEpisodes(c *gin.Context) {
	var episodes []models.Episode
	log.Printf("Episode: %+v", episodes)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kiểm tra cache Redis
	cachedEpisodes, err := dbs.RedisClient.Get(ctx, "episodes").Result()
	if err == nil && cachedEpisodes != "" {
		// Nếu có cache, giải mã và trả về
		json.Unmarshal([]byte(cachedEpisodes), &episodes)
		c.JSON(http.StatusOK, episodes)
		return
	}

	// Nếu không có cache, lấy từ MongoDB
	cursor, err := episodeCollection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi lấy danh sách tập phim"})
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var episode models.Episode
		if err := cursor.Decode(&episode); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi đọc dữ liệu tập phim"})
			return
		}
		episodes = append(episodes, episode)
	}

	// Lưu dữ liệu vào Redis cache
	episodesJSON, _ := json.Marshal(episodes)
	dbs.RedisClient.Set(ctx, "episodes", string(episodesJSON), 30*time.Minute)

	c.JSON(http.StatusOK, episodes)
}

// Lấy tập phim theo ID
func GetEpisodeByID(c *gin.Context) {
	id := c.Query("id")
	episodeID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tập phim không hợp lệ"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kiểm tra cache Redis
	cachedEpisode, err := dbs.RedisClient.Get(ctx, "episode_"+id).Result()
	if err == nil && cachedEpisode != "" {
		var episode models.Episode
		json.Unmarshal([]byte(cachedEpisode), &episode)
		c.JSON(http.StatusOK, episode)
		return
	}

	// Nếu không có cache, lấy từ MongoDB
	var episode models.Episode
	err = episodeCollection.FindOne(ctx, bson.M{"_id": episodeID}).Decode(&episode)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy tập phim"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi tìm tập phim"})
		}
		return
	}

	// Lưu dữ liệu vào Redis cache
	episodeJSON, _ := json.Marshal(episode)
	dbs.RedisClient.Set(ctx, "episode_"+id, string(episodeJSON), 30*time.Minute)

	c.JSON(http.StatusOK, episode)
}

// Cập nhật tập phim
func UpdateEpisode(c *gin.Context) {
	var episode models.Episode
	if err := c.ShouldBindJSON(&episode); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	episode.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": episode.ID}
	update := bson.M{"$set": episode}

	_, err := episodeCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi cập nhật tập phim"})
		return
	}

	// Xóa cache Redis sau khi cập nhật
	dbs.RedisClient.Del(ctx, "episode_"+episode.ID.Hex())
	dbs.RedisClient.Del(ctx, "episodes")

	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật tập phim thành công"})
}

// Xóa tập phim
func DeleteEpisode(c *gin.Context) {
	movieCollection := models.GetMovieCollection()
	id := c.Param("id")
	episodeID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid episode ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = episodeCollection.DeleteOne(ctx, bson.M{"_id": episodeID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting episode"})
		return
	}

	// Cập nhật tất cả các bộ phim, xóa episodeID khỏi mảng episode của Movie
	_, err = movieCollection.UpdateMany(
		ctx,
		bson.M{"episode": episodeID}, // Tìm những bộ phim chứa episodeID
		bson.M{"$pull": bson.M{"episode": episodeID}}, // Xóa episodeID khỏi mảng episode
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating movies"})
		return
	}

	// Xóa cache Redis sau khi xóa
	dbs.RedisClient.Del(ctx, "episode_"+id)
	dbs.RedisClient.Del(ctx, "episodes")
	dbs.RedisClient.Del(ctx, "movies_cache")

	c.JSON(http.StatusOK, gin.H{"message": "episode deleted successfully"})
}
