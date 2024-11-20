// controllers/movie_controller
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
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Thêm movie mới
func AddMovie(c *gin.Context) {
	movieCollection := models.GetMovieCollection()
	var movie models.Movie
	if err := c.ShouldBindJSON(&movie); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie data"})
		return
	}

	// Gán ID mới cho movie
	movie.ID = primitive.NewObjectID()
	movie.Status = 1
	// Gán ngày giờ hiện tại cho trường Createat
	movie.CreatedAt = time.Now()
	movie.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := movieCollection.InsertOne(ctx, movie)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting movie"})
		return
	}

	// Xóa cache tổng thể của danh sách movies
	err = dbs.RedisClient.Del(ctx, "movies_cache").Err()
	if err != nil {
		log.Printf("Error deleting movies cache: %v", err)
	}

	c.JSON(http.StatusOK, movie)
}

// Lấy tất cả movies
func GetAllMovies(c *gin.Context) {
	movieCollection := models.GetMovieCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kiểm tra Redis cache
	cachedMovies, err := dbs.RedisClient.Get(ctx, "movies_cache").Result()
	if err == redis.Nil {
		var movies []models.Movie
		cursor, err := movieCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching movies"})
			return
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var movie models.Movie
			if err := cursor.Decode(&movie); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding movie"})
				return
			}
			movies = append(movies, movie)
		}

		// Lưu cache
		jsonData, _ := json.Marshal(movies)
		err = dbs.RedisClient.Set(ctx, "movies_cache", jsonData, 10*time.Minute).Err()
		if err != nil {
			log.Printf("Error caching movies: %v", err)
		}

		c.JSON(http.StatusOK, movies)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching cache"})
	} else {
		var movies []models.Movie
		json.Unmarshal([]byte(cachedMovies), &movies)
		c.JSON(http.StatusOK, movies)
	}
}

// Lấy một movie theo ID
func GetMovieByID(c *gin.Context) {
	movieCollection := models.GetMovieCollection()
	id := c.Param("id")
	movieID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kiểm tra Redis cache
	cacheKey := "movie_cache_" + id
	cachedMovie, err := dbs.RedisClient.Get(ctx, cacheKey).Result()
	if err == redis.Nil {

		var movie models.Movie
		err = movieCollection.FindOne(ctx, bson.M{"_id": movieID}).Decode(&movie)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching movie"})
			}
			return
		}

		jsonData, _ := json.Marshal(movie)
		err = dbs.RedisClient.Set(ctx, cacheKey, jsonData, 10*time.Minute).Err()
		if err != nil {
			log.Printf("Error caching movie: %v", err)
		}

		c.JSON(http.StatusOK, movie)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching cache"})
	} else {
		var movie models.Movie
		json.Unmarshal([]byte(cachedMovie), &movie)
		c.JSON(http.StatusOK, movie)
	}
}

// Cập nhật movie
func UpdateMovie(c *gin.Context) {
	movieCollection := models.GetMovieCollection()
	var movie models.Movie
	if err := c.ShouldBindJSON(&movie); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie data"})
		return
	}

	// Gán ngày giờ hiện tại cho trường Updateat
	movie.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": movie.ID}
	update := bson.M{"$set": movie}
	_, err := movieCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating movie"})
		return
	}

	// Xóa cache liên quan đến movie đã được cập nhật
	cacheKey := "movie_cache_" + movie.ID.Hex()
	err = dbs.RedisClient.Del(ctx, cacheKey).Err()
	if err != nil {
		log.Printf("Error deleting cache for updated movie: %v", err)
	}

	// Xóa cache tổng thể của danh sách movies
	err = dbs.RedisClient.Del(ctx, "movies_cache").Err()
	if err != nil {
		log.Printf("Error deleting movies cache: %v", err)
	}

	c.Status(http.StatusOK)
}

// Xóa movie
func DeleteMovie(c *gin.Context) {
	movieCollection := models.GetMovieCollection()
	id := c.Param("id")
	movieID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = movieCollection.DeleteOne(ctx, bson.M{"_id": movieID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting movie"})
		return
	}

	// Xóa cache liên quan đến movie đã xóa
	cacheKey := "movie_cache_" + id
	err = dbs.RedisClient.Del(ctx, cacheKey).Err()
	if err != nil {
		log.Printf("Error deleting cache for deleted movie: %v", err)
	}

	// Xóa cache tổng thể của danh sách movies
	err = dbs.RedisClient.Del(ctx, "movies_cache").Err()
	if err != nil {
		log.Printf("Error deleting movies cache: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Movie deleted successfully"})
}

// Lấy thông tin Movie cùng với các Episodes liên quan
func GetMovieWithEpisodes(c *gin.Context) {
	movieCollection := models.GetMovieCollection()
	id := c.Param("id")
	movieID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Kiểm tra Redis cache
	cacheKey := "movie_with_episodes_cache_" + id
	cachedData, err := dbs.RedisClient.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		// Nếu không có cache, thực hiện truy vấn MongoDB
		pipeline := mongo.Pipeline{
			bson.D{{"$match", bson.D{{"_id", movieID}}}},
			bson.D{{"$lookup", bson.D{
				{"from", "episodes"},
				{"localField", "_id"},
				{"foreignField", "movie_id"},
				{"as", "episodes"},
			}}},
			bson.D{
				{"$lookup", bson.D{
					{"from", "categories"},
					{"localField", "category"},
					{"foreignField", "_id"},
					{"as", "categoryDetails"},
				}},
			},
		}

		cursor, err := movieCollection.Aggregate(ctx, pipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error aggregating data"})
			return
		}
		defer cursor.Close(ctx)

		if !cursor.Next(ctx) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found or no episodes available"})
			return
		}

		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding result"})
			return
		}

		// Lưu vào Redis cache
		jsonData, _ := json.Marshal(result)
		err = dbs.RedisClient.Set(ctx, cacheKey, jsonData, 10*time.Minute).Err()
		if err != nil {
			log.Printf("Error caching movie with episodes: %v", err)
		}

		c.JSON(http.StatusOK, result)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching cache"})
	} else {
		// Dữ liệu có sẵn trong cache
		var result bson.M
		json.Unmarshal([]byte(cachedData), &result)
		c.JSON(http.StatusOK, result)
	}
}

func CreateMoviesBulk(c *gin.Context) {
	var payload struct {
		Movies []models.Movie `json:"movies"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	movieCollection := models.GetMovieCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Convert to slice of interface{}
	var movies []interface{}
	for _, movie := range payload.Movies {
		movie.ID = primitive.NewObjectID()
		movie.CreatedAt = time.Now()
		movie.UpdatedAt = time.Now()
		movies = append(movies, movie)
	}

	_, err := movieCollection.InsertMany(ctx, movies)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert movies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Movies created successfully"})
}
