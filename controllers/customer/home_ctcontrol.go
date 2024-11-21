// controllers/movie_controller.go
package controllers

import (
	"context"
	"encoding/json"
	"fire-watch/dbs"
	"fire-watch/models"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetAllMovies(c *gin.Context) ([]models.Movie, error) {
	// Lấy collection Movie từ MongoDB
	movieCollection := models.GetMovieCollection()

	// Context với timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Lấy tham số phân trang từ query
	pageStr := c.Query("page")
	limit := 6 // Số lượng bản ghi trên mỗi trang

	// Xử lý page
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1 // Nếu không có hoặc không hợp lệ, mặc định là trang 1
	}

	// Tính toán `$skip`
	skip := (page - 1) * limit

	// Kiểm tra cache từ Redis
	cacheKey := fmt.Sprintf("movieshome_page_%d", page)
	cachedMovies, err := dbs.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil && cachedMovies != "" {
		var movies []models.Movie
		if err := json.Unmarshal([]byte(cachedMovies), &movies); err == nil {
			return movies, nil
		}
	}

	// Trường hợp cache không có, lấy dữ liệu từ MongoDB
	var movies []models.Movie
	pipeline := mongo.Pipeline{
		bson.D{{"$match", bson.D{{"deleted", bson.D{{"$ne", "deleted"}}}}}},
		bson.D{{"$match", bson.D{{"status", bson.D{{"$ne", 2}}}}}},
		bson.D{{"$sort", bson.D{{"position", 1}}}}, // Sắp xếp theo trường `position`
		bson.D{{"$skip", skip}},                    // Bỏ qua số lượng bản ghi tương ứng với `skip`
		bson.D{{"$limit", limit}},                  // Lấy tối đa `limit` bản ghi
	}

	cursor, err := movieCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var movie models.Movie
		if err := cursor.Decode(&movie); err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}

	// Serialize dữ liệu movies để lưu cache
	moviesJSON, _ := json.Marshal(movies)

	// Lưu cache vào Redis với TTL 30 phút
	err = dbs.RedisClient.Set(ctx, cacheKey, string(moviesJSON), 30*time.Minute).Err()
	if err != nil {
		log.Printf("Error caching movies data: %v", err)
	}

	return movies, nil
}

func GetCategoriesWithMovies(c *gin.Context) ([]bson.M, error) {
	// Lấy collection Category từ MongoDB
	categoryCollection := models.GetCategoryCollection()

	// Context với timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Lấy tham số phân trang từ query
	pageStr := c.Query("page")
	limit := 6 // Số lượng bản ghi trên mỗi trang

	// Xử lý page
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1 // Nếu không có hoặc không hợp lệ, mặc định là trang 1
	}

	// Tính toán `$skip`
	skip := (page - 1) * limit

	// Kiểm tra cache từ Redis
	cacheKey := fmt.Sprintf("categorieswithmovie_%d", page)
	cachedData, err := dbs.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil && cachedData != "" {
		var categorieswithmovie []bson.M
		// Giải mã cache thành danh sách danh mục với phim
		if err := json.Unmarshal([]byte(cachedData), &categorieswithmovie); err == nil {
			// Trả về từ cache
			return categorieswithmovie, nil
		}
	}

	// Pipeline Aggregation
	pipeline := mongo.Pipeline{
		// Lọc các danh mục chưa bị xóa
		bson.D{{"$match", bson.D{{"deleted", bson.D{{"$ne", "deleted"}}}}}},
		// Lookup để lấy danh sách phim
		bson.D{{"$lookup", bson.D{
			{"from", "movies"},           // Collection liên kết (movies)
			{"localField", "_id"},        // Trường `_id` từ `categories`
			{"foreignField", "category"}, // Trường `category` từ `movies`
			{"as", "movies"},             // Kết quả sẽ được gắn vào `movies`
		}}},
		// Lọc các phim bị xóa hoặc không hoạt động
		bson.D{{"$addFields", bson.D{
			{"movies", bson.D{{"$filter", bson.D{
				{"input", "$movies"},
				{"as", "movie"},
				{"cond", bson.D{
					{"$and", bson.A{
						bson.D{{"$ne", bson.A{"$$movie.deleted", "deleted"}}},
						bson.D{{"$eq", bson.A{"$$movie.status", 1}}},
					}},
				}},
			}}}},
		}}},
		// Giới hạn số lượng phim trong mỗi danh mục
		bson.D{{"$addFields", bson.D{
			{"movies", bson.D{{"$slice", bson.A{"$movies", skip, limit}}}}, // Lấy phim từ vị trí `skip` đến `limit`
		}}},
		// Sắp xếp danh mục theo `created_at`
		bson.D{{"$sort", bson.D{{"created_at", -1}}}},
	}

	// Thực thi aggregation
	cursor, err := categoryCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Lấy kết quả và decode thành danh sách
	var categorieswithmovie []bson.M
	if err := cursor.All(ctx, &categorieswithmovie); err != nil {
		return nil, err
	}

	// Serialize dữ liệu categories để lưu cache
	categoriesJSON, _ := json.Marshal(categorieswithmovie)

	// Lưu vào Redis với TTL 30 phút
	err = dbs.RedisClient.Set(ctx, cacheKey, string(categoriesJSON), 30*time.Minute).Err()
	if err != nil {
		log.Printf("Error caching categories with movies: %v", err)
	}

	// Trả về danh mục với phim
	return categorieswithmovie, nil
}

func GetUserFromRedis(c *gin.Context) (map[string]interface{}, error) {
	// Lấy user_id từ query hoặc context
	userID := c.Query("user_id")
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	// Truy xuất Redis với user_id
	ctx := context.Background()
	userJSON, err := dbs.RedisClient.Get(ctx, "user:"+userID).Result()
	if err != nil {
		return nil, fmt.Errorf("error fetching user from Redis: %v", err)
	}

	// Chuyển đổi dữ liệu JSON thành map
	var user map[string]interface{}
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		return nil, fmt.Errorf("error decoding user data: %v", err)
	}

	return user, nil
}

func SearchMovies(c *gin.Context) ([]models.Movie, error) {
	// Lấy collection Movie từ MongoDB
	movieCollection := models.GetMovieCollection()

	// Context với timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Lấy từ khóa tìm kiếm từ query parameter
	searchQuery := c.Query("search")
	if searchQuery == "" {
		return nil, fmt.Errorf("Search query is empty")
	}

	// Tạo pipeline để tìm kiếm trong MongoDB
	pipeline := mongo.Pipeline{
		bson.D{{"$match", bson.D{
			{"deleted", bson.D{{"$ne", "deleted"}}},                       // Bỏ qua các phim bị xóa
			{"status", bson.D{{"$ne", 2}}},                                // Bỏ qua các phim không hoạt động
			{"title", bson.D{{"$regex", searchQuery}, {"$options", "i"}}}, // Tìm kiếm không phân biệt chữ hoa/thường
		}}},
		bson.D{{"$sort", bson.D{{"position", 1}}}}, // Sắp xếp theo vị trí tăng dần
	}

	// Lấy kết quả từ MongoDB
	cursor, err := movieCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var movies []models.Movie
	for cursor.Next(ctx) {
		var movie models.Movie
		if err := cursor.Decode(&movie); err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return movies, nil
}
