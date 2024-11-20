// controllers/movie_controller.go
package controllers

import (
	"context"
	"encoding/json"
	"fire-watch/dbs"
	"fire-watch/models"
	"fire-watch/websocket"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Định nghĩa kích thước tối đa cho ảnh (2MB)
const maxImageSize = 2 * 1024 * 1024 // 2MB

// Hàm xử lý upload file ảnh một cách đồng thời
func processImage(c *gin.Context, file *multipart.FileHeader, allowedFormats map[string]bool, maxImageSize int64) (string, error) {
	// Kiểm tra định dạng ảnh
	if !allowedFormats[file.Header.Get("Content-Type")] {
		return "", fmt.Errorf("Định dạng ảnh không hợp lệ. Chỉ chấp nhận JPEG, WEBP và PNG.")
	}

	// Kiểm tra kích thước ảnh
	if file.Size > maxImageSize {
		return "", fmt.Errorf("Dung lượng ảnh quá lớn. Kích thước tối đa là 2MB.")
	}

	// Tạo đường dẫn lưu file và kiểm tra nếu file đã tồn tại
	imagePath := fmt.Sprintf("views/uploads/images/%s", file.Filename)
	if _, err := os.Stat(imagePath); err == nil {
		return "", fmt.Errorf("File ảnh đã tồn tại. Vui lòng upload file khác.")
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("Không thể kiểm tra file đã tồn tại")
	}

	// Lưu file ảnh
	if err := c.SaveUploadedFile(file, imagePath); err != nil {
		return "", fmt.Errorf("Lưu ảnh thất bại")
	}

	return file.Filename, nil
}

func AddMovie(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	// Tạo một đối tượng Movie mới
	var movie models.Movie
	collection := models.GetMovieCollection()

	// Lấy dữ liệu từ form (không bao gồm file)
	movie.Title = c.PostForm("title")
	movie.NameEng = c.PostForm("name_eng")
	movie.Description = c.PostForm("description")
	movie.Tags = c.PostForm("tags")
	movie.Slug = c.PostForm("slug")
	movie.Duration = c.PostForm("duration")
	movie.Trailer = c.PostForm("trailer")

	// Lấy và chuyển đổi dữ liệu từ các trường chọn
	status, _ := strconv.Atoi(c.PostForm("status"))
	hotmovie, _ := strconv.Atoi(c.PostForm("hotmovie"))
	maxquality, _ := strconv.Atoi(c.PostForm("maxquality"))
	season, _ := strconv.Atoi(c.PostForm("season"))
	numofep, _ := strconv.Atoi(c.PostForm("numofep"))
	year, _ := strconv.Atoi(c.PostForm("year"))

	movie.Status = status
	movie.Hotmovie = hotmovie
	movie.MaxQuality = maxquality
	movie.Season = season
	movie.Numofep = numofep
	movie.Year = year

	// Định nghĩa các định dạng ảnh được chấp nhận
	allowedFormats := map[string]bool{"image/jpeg": true, "image/png": true, "image/webp": true, "image/avif": true}

	// Xử lý file ảnh chính
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "File ảnh chính là bắt buộc"})
		return
	}

	primaryImageFileName, err := processImage(c, file, allowedFormats, maxImageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	movie.Image = primaryImageFileName

	// Xử lý các ảnh phụ một cách đồng thời
	moreFiles := c.Request.MultipartForm.File["moreimage[]"]
	var moreImages []string
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(moreFiles))

	for _, moreFile := range moreFiles {
		wg.Add(1)
		go func(file *multipart.FileHeader) {
			defer wg.Done()
			fileName, err := processImage(c, file, allowedFormats, maxImageSize)
			if err != nil {
				errChan <- err
				return
			}
			// Sử dụng Mutex để đảm bảo thêm tên file vào mảng moreImages một cách an toàn
			mu.Lock()
			moreImages = append(moreImages, fileName)
			mu.Unlock()
		}(moreFile)
	}

	// Chờ tất cả các goroutine kết thúc
	wg.Wait()
	close(errChan)

	// Kiểm tra nếu có lỗi từ bất kỳ goroutine nào
	if err := <-errChan; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	movie.Moreimage = moreImages

	// Lấy thời gian hiện tại cho trường CreatedAt và UpdatedAt
	movie.CreatedAt = time.Now()
	movie.UpdatedAt = time.Now()
	movie.Episode = make([]primitive.ObjectID, 0) // Luôn khởi tạo là mảng rỗng
	sub := c.PostFormArray("sub[]")
	movie.Sub = sub

	// Lấy danh sách category từ form
	categoryIDs := c.PostFormArray("category[]")
	for _, id := range categoryIDs {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid category ID"})
			return
		}
		movie.Category = append(movie.Category, oid)
	}

	// Lấy danh sách genre từ form
	genreIDs := c.PostFormArray("genre[]")
	for _, id := range genreIDs {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid genre ID"})
			return
		}
		movie.Genre = append(movie.Genre, oid)
	}

	// Lấy danh sách country từ form
	countryID := c.PostForm("country")
	oid, err := primitive.ObjectIDFromHex(countryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid country ID"})
		return
	}
	movie.Country = oid
	// Tạo context với timeout 5 giây
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Tạo ID mới trước khi tìm kiếm
	movie.ID = primitive.NewObjectID()

	err = collection.FindOne(ctx, bson.M{"_id": movie.ID}).Decode(&movie)
	if err == nil {
		// Nếu phim với ID này đã tồn tại, tạo ID mới và thử lại
		movie.ID = primitive.NewObjectID()
	}
	// Xác thực dữ liệu
	if err := movie.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Kiểm tra và thiết lập vị trí của phim
	var lastMovie models.Movie
	err = collection.FindOne(ctx, bson.M{}, options.FindOne().SetSort(bson.M{"position": -1})).Decode(&lastMovie)
	if err == mongo.ErrNoDocuments {
		movie.Position = 1
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Lỗi trong khi kiểm tra phim"})
		return
	} else {
		movie.Position = lastMovie.Position + 1
	}

	// Chèn vào cơ sở dữ liệu
	_, err = collection.InsertOne(ctx, movie)
	if err != nil {
		log.Println("Failed to insert movie:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to add movie"})
		return
	}

	dbs.DeleteCacheByKeyword(ctx, "movie")

	_, _, _, _, _, _, err = GetAllMoviesWithOptions(c, websocketServer)
	if err != nil {
		log.Println("Error fetching movies data:", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Movie added successfully", "movie": movie})
}

func GetAllMoviesWithOptions(c *gin.Context, websocketServer *websocket.WebSocketServer) ([]models.Movie, []models.Category, []models.Genre, []models.Country, []models.Episode, []models.Server, error) {
	// Các collection MongoDB
	movieCollection := models.GetMovieCollection()
	categoryCollection := models.GetCategoryCollection()
	genreCollection := models.GetGenreCollection()
	countryCollection := models.GetCountryCollection()
	episodeCollection := models.GetEpisodeCollection() // Thêm episodeCollection
	serverCollection := models.GetServerCollection()

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

	cacheKey := fmt.Sprintf("movies_%d", page)

	// // Lấy nhiều keys cùng lúc từ Redis
	// cachedData, err := dbs.RedisClient.MGet(ctx, cacheKey, "categories", "genres", "countries", "episodes", "servers").Result()
	// if err == nil {
	// 	// Kiểm tra cache không bị rỗng
	// 	if cachedData[0] != nil && cachedData[1] != nil && cachedData[2] != nil && cachedData[3] != nil && cachedData[4] != nil && cachedData[5] != nil {
	// 		var movies []models.Movie
	// 		var categories []models.Category
	// 		var genres []models.Genre
	// 		var countries []models.Country
	// 		var episodes []models.Episode
	// 		var servers []models.Server

	// 		// Giải mã cache thành các dữ liệu tương ứng
	// 		json.Unmarshal([]byte(cachedData[0].(string)), &movies)
	// 		json.Unmarshal([]byte(cachedData[1].(string)), &categories)
	// 		json.Unmarshal([]byte(cachedData[2].(string)), &genres)
	// 		json.Unmarshal([]byte(cachedData[3].(string)), &countries)
	// 		json.Unmarshal([]byte(cachedData[4].(string)), &episodes)
	// 		json.Unmarshal([]byte(cachedData[5].(string)), &servers)

	// 		// Gửi thông báo cập nhật qua WebSocket
	// 		notifyClients(websocketServer, movies)

	// 		return movies, categories, genres, countries, episodes, servers, nil
	// 	}
	// }

	// Trường hợp cache không có, lấy dữ liệu từ MongoDB
	var movies []models.Movie
	var categories []models.Category
	var genres []models.Genre
	var countries []models.Country
	var episodes []models.Episode
	var servers []models.Server

	// Pipeline cho movie, như code trước đây
	pipeline := mongo.Pipeline{
		bson.D{{"$match", bson.D{{"deleted", bson.D{{"$ne", "deleted"}}}}}},
		bson.D{{"$lookup", bson.D{
			{"from", "categories"}, {"localField", "category"}, {"foreignField", "_id"}, {"as", "categoryDetails"},
		}}},
		bson.D{{"$addFields", bson.D{
			{"categoryDetails", bson.D{
				{"$filter", bson.D{
					{"input", "$categoryDetails"},
					{"as", "category"},
					{"cond", bson.D{{"$ne", bson.A{"$$category.deleted", "deleted"}}}},
				}},
			}},
		}}},
		bson.D{{"$lookup", bson.D{
			{"from", "genres"}, {"localField", "genre"}, {"foreignField", "_id"}, {"as", "genreDetails"},
		}}},
		bson.D{{"$addFields", bson.D{
			{"genreDetails", bson.D{
				{"$filter", bson.D{
					{"input", "$genreDetails"},
					{"as", "genre"},
					{"cond", bson.D{{"$ne", bson.A{"$$genre.deleted", "deleted"}}}},
				}},
			}},
		}}},
		bson.D{{"$lookup", bson.D{
			{"from", "countries"}, {"localField", "country"}, {"foreignField", "_id"}, {"as", "countryDetails"},
		}}},
		bson.D{{"$addFields", bson.D{
			{"countryDetails", bson.D{
				{"$filter", bson.D{
					{"input", "$countryDetails"},
					{"as", "country"},
					{"cond", bson.D{{"$ne", bson.A{"$$country.deleted", "deleted"}}}},
				}},
			}},
		}}},
		bson.D{{"$sort", bson.D{{"position", 1}}}}, // Sắp xếp theo trường position tăng dần
		bson.D{{"$skip", skip}},                    // Bỏ qua số lượng bản ghi
		bson.D{{"$limit", limit}},                  // Lấy số lượng bản ghi
	}

	cursor, err := movieCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var movie models.Movie
		if err := cursor.Decode(&movie); err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}
		movies = append(movies, movie)
	}

	// Lấy tất cả categories từ MongoDB
	categoryCursor, err := categoryCollection.Find(ctx, bson.M{"deleted": bson.M{"$ne": "deleted"}})
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	defer categoryCursor.Close(ctx)

	for categoryCursor.Next(ctx) {
		var category models.Category
		if err := categoryCursor.Decode(&category); err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}
		categories = append(categories, category)
	}

	// Lấy tất cả genres từ MongoDB
	genreCursor, err := genreCollection.Find(ctx, bson.M{"deleted": bson.M{"$ne": "deleted"}})
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	defer genreCursor.Close(ctx)

	for genreCursor.Next(ctx) {
		var genre models.Genre
		if err := genreCursor.Decode(&genre); err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}
		genres = append(genres, genre)
	}

	// Lấy tất cả countries từ MongoDB
	countryCursor, err := countryCollection.Find(ctx, bson.M{"deleted": bson.M{"$ne": "deleted"}})
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	defer countryCursor.Close(ctx)

	for countryCursor.Next(ctx) {
		var country models.Country
		if err := countryCursor.Decode(&country); err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}
		countries = append(countries, country)
	}

	// Lấy tất cả episodes từ MongoDB
	episodeCursor, err := episodeCollection.Find(ctx, bson.M{"deleted": bson.M{"$ne": "deleted"}})
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	defer episodeCursor.Close(ctx)

	for episodeCursor.Next(ctx) {
		var episode models.Episode
		if err := episodeCursor.Decode(&episode); err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}
		episodes = append(episodes, episode)
	}
	// Lấy tất cả servers từ MongoDB
	serverCursor, err := serverCollection.Find(ctx, bson.M{"deleted": bson.M{"$ne": "deleted"}})
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	defer serverCursor.Close(ctx)

	for serverCursor.Next(ctx) {
		var server models.Server
		if err := serverCursor.Decode(&server); err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}
		servers = append(servers, server)
	}

	// Serialize dữ liệu
	moviesJSON, _ := json.Marshal(movies)
	categoriesJSON, _ := json.Marshal(categories)
	genresJSON, _ := json.Marshal(genres)
	countriesJSON, _ := json.Marshal(countries)
	episodesJSON, _ := json.Marshal(episodes)
	serversJSON, _ := json.Marshal(servers)

	// Lưu vào Redis với TTL
	_, err = dbs.RedisClient.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Set(ctx, cacheKey, string(moviesJSON), 30*time.Minute)
		pipe.Set(ctx, "categories", string(categoriesJSON), 30*time.Minute)
		pipe.Set(ctx, "genres", string(genresJSON), 30*time.Minute)
		pipe.Set(ctx, "countries", string(countriesJSON), 30*time.Minute)
		pipe.Set(ctx, "episodes", string(episodesJSON), 30*time.Minute)
		pipe.Set(ctx, "servers", string(serversJSON), 30*time.Minute)
		return nil
	})
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	// Gửi thông báo cập nhật qua WebSocket
	notifyClients(websocketServer, movies)

	return movies, categories, genres, countries, episodes, servers, nil
}

// Hàm gửi thông báo cập nhật tới các client qua WebSocket
func notifyClients(websocketServer *websocket.WebSocketServer, movies []models.Movie) {
	message := map[string]interface{}{
		"type":    "movie",
		"message": "A new movie or update detected!",
		"movies":  movies,
	}

	// Chuyển thông báo thành JSON
	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Println("Error encoding JSON message:", err)
		return
	}

	// Log tin nhắn trước khi gửi
	log.Println("Broadcasting message:", string(messageJSON))

	// Gửi thông điệp tới tất cả các client qua WebSocket
	websocketServer.BroadcastMessage(messageJSON)
}

// UpdateMovie cập nhật thông tin của một movie
func UpdateMovie(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	// Lấy ID của movie từ URL
	movieID := c.Param("id")
	log.Println("Movie ID from URL:", movieID) // Log ID

	// Chuyển ID thành ObjectID
	oid, err := primitive.ObjectIDFromHex(movieID)
	if err != nil {
		log.Println("Invalid movie ID:", err) // Log lỗi nếu ObjectID không hợp lệ
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "The provided movie ID is not valid",
		})
		return
	}
	log.Println("Converted movie ID to ObjectID:", oid) // Log ObjectID đã chuyển đổi

	// Tìm movie hiện tại trong database
	collection := models.GetMovieCollection()
	var existingMovie models.Movie
	err = collection.FindOne(context.TODO(), bson.M{"_id": oid}).Decode(&existingMovie)
	if err != nil {
		log.Println("Movie not found:", err) // Log lỗi nếu không tìm thấy movie
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Movie not found"})
		return
	}
	log.Println("Found existing movie:", existingMovie) // Log thông tin movie tìm thấy

	// Tạo đối tượng để cập nhật
	var movieUpdate models.Movie

	// Lấy dữ liệu từ form (không bao gồm file)
	movieUpdate.Title = c.PostForm("title")
	movieUpdate.NameEng = c.PostForm("name_eng")
	movieUpdate.Description = c.PostForm("description")
	movieUpdate.Tags = c.PostForm("tags")
	movieUpdate.Slug = c.PostForm("slug")
	movieUpdate.Duration = c.PostForm("duration")
	movieUpdate.Trailer = c.PostForm("trailer")

	log.Println("Form data collected:", movieUpdate) // Log dữ liệu form

	// Lấy và chuyển đổi dữ liệu từ các trường chọn
	status, _ := strconv.Atoi(c.PostForm("status"))
	hotmovie, _ := strconv.Atoi(c.PostForm("hotmovie"))
	maxquality, _ := strconv.Atoi(c.PostForm("maxquality"))
	season, _ := strconv.Atoi(c.PostForm("season"))
	numofep, _ := strconv.Atoi(c.PostForm("numofep"))
	year, _ := strconv.Atoi(c.PostForm("year"))

	movieUpdate.Status = status
	movieUpdate.Hotmovie = hotmovie
	movieUpdate.MaxQuality = maxquality
	movieUpdate.Season = season
	movieUpdate.Numofep = numofep
	movieUpdate.Year = year

	if len(movieUpdate.Episode) == 0 {
		movieUpdate.Episode = existingMovie.Episode
	}

	log.Println("Status-related fields set:", movieUpdate) // Log dữ liệu cập nhật liên quan tới trạng thái

	// Các bước tiếp theo (danh sách category, genre, country, và cập nhật vào DB) tương tự, cũng nên thêm log tương ứng.\
	sub := c.PostFormArray("sub[]")
	movieUpdate.Sub = sub
	// Lấy danh sách category từ form
	categoryIDs := c.PostFormArray("category[]")
	log.Printf("Category IDs: %v", categoryIDs)
	for _, id := range categoryIDs {
		categoryOID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			// Log lỗi chi tiết khi category ID không hợp lệ
			log.Printf("Error converting category ID: %s, error: %v", id, err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid category ID"})
			return
		}
		movieUpdate.Category = append(movieUpdate.Category, categoryOID)
	}

	// Lấy danh sách genre từ form
	genreIDs := c.PostFormArray("genre[]")
	log.Printf("Genre IDs: %v", genreIDs)
	for _, id := range genreIDs {
		genreOID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			// Log lỗi chi tiết khi genre ID không hợp lệ
			log.Printf("Error converting genre ID: %s, error: %v", id, err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid genre ID"})
			return
		}
		movieUpdate.Genre = append(movieUpdate.Genre, genreOID)
	}

	// Lấy country từ form
	countryID := c.PostForm("country")
	log.Printf("Country ID: %v", countryID)

	// Khai báo biến countryOID trước khi sử dụng
	var countryOID primitive.ObjectID

	countryOID, err = primitive.ObjectIDFromHex(countryID)
	if err != nil {
		// Log lỗi chi tiết khi country ID không hợp lệ
		log.Printf("Error converting country ID: %s, error: %v", countryID, err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid country ID"})
		return
	}

	movieUpdate.Country = countryOID

	// Cập nhật thời gian
	movieUpdate.UpdatedAt = time.Now()

	// Xác thực dữ liệu
	if err := movieUpdate.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Định nghĩa các định dạng ảnh được chấp nhận
	allowedFormats := map[string]bool{"image/jpeg": true, "image/png": true, "image/webp": true, "image/avif": true}
	// Xử lý ảnh chính
	file, err := c.FormFile("image")
	if err == nil {
		imageFileName, err := processImage(c, file, allowedFormats, maxImageSize)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		movieUpdate.Image = imageFileName
		log.Println("Image uploaded and set:", imageFileName)
	} else {
		movieUpdate.Image = existingMovie.Image
		log.Println("No new image, keeping old image.")
	}

	// Xử lý ảnh phụ nếu có
	moreFiles, moreFilesExists := c.Request.MultipartForm.File["moreimage[]"]
	if moreFilesExists {
		var wg sync.WaitGroup
		var mu sync.Mutex
		moreImages := existingMovie.Moreimage

		for i, moreFile := range moreFiles {
			wg.Add(1)
			go func(i int, moreFile *multipart.FileHeader) {
				defer wg.Done()

				imageFileName, err := processImage(c, moreFile, allowedFormats, maxImageSize)
				if err != nil {
					log.Printf("Error processing additional image %d: %v", i+1, err)
					return
				}

				mu.Lock()
				moreImages = append(moreImages, imageFileName)
				mu.Unlock()
				log.Printf("Additional image %d uploaded: %s", i+1, imageFileName)

			}(i, moreFile)
		}

		wg.Wait()
		movieUpdate.Moreimage = moreImages
		log.Println("All additional images processed:", moreImages)
	} else {
		movieUpdate.Moreimage = existingMovie.Moreimage
		log.Println("No new additional images, keeping old ones.")
	}

	// Cuối cùng cập nhật vào cơ sở dữ liệu và ghi log:
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updateResult, err := collection.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": movieUpdate})
	if err != nil || updateResult.MatchedCount == 0 {
		log.Println("Failed to update movie:", err) // Log khi không cập nhật được
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update movie"})
		return
	}

	dbs.DeleteCacheByKeyword(ctx, "movie")

	// Lấy dữ liệu cập nhật cho WebSocket mà không chờ
	go func() {
		_, _, _, _, _, _, err := GetAllMoviesWithOptions(c, websocketServer)
		if err != nil {
			log.Println("Error fetching movies data:", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "Movie updated successfully", "movie": movieUpdate})
}

// DeleteMovie là hàm xử lý yêu cầu xóa ảo
func DeleteMovie(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	movieCollection := models.GetMovieCollection()
	// Lấy ID từ route
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	// Tìm Movie hiện tại để lấy thông tin hình ảnh
	var existingMovie models.Movie
	err = movieCollection.FindOne(context.TODO(), bson.M{"_id": objectID}).Decode(&existingMovie)
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
		return
	}

	// Xóa ảnh chính nếu có
	if existingMovie.Image != "" {
		mainImagePath := fmt.Sprintf("views/uploads/images/%s", existingMovie.Image)
		if err := os.Remove(mainImagePath); err != nil && !os.IsNotExist(err) {
			log.Println("Failed to delete main image:", err)
		}
	}

	// Xóa các ảnh phụ nếu có
	for _, moreImage := range existingMovie.Moreimage {
		moreImagePath := fmt.Sprintf("views/uploads/images/%s", moreImage)
		if err := os.Remove(moreImagePath); err != nil && !os.IsNotExist(err) {
			log.Println("Failed to delete additional image:", err)
		}
	}

	// Tạo filter để tìm Movie theo ID
	filter := bson.M{"_id": objectID}

	// Tạo update để cập nhật trường Deleted = 1
	update := bson.M{
		"$set": bson.M{
			"deleted": "deleted",
		},
	}

	// Cập nhật trong MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = movieCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Can not delete movie!"})
		return
	}
	// Xóa cache trong Redis nếu có sử dụng
	dbs.DeleteCacheByKeyword(ctx, "movie")

	_, _, _, _, _, _, err = GetAllMoviesWithOptions(c, websocketServer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error fetching movies data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Danh mục đã được xóa"})
}

// Hàm UpdateMovieField để cập nhật trường cụ thể của Movie
func UpdateMovieField(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	// Lấy ID của Movie từ tham số URL
	idParam := c.Param("id")

	movieID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})
		return
	}

	// Nhận dữ liệu từ request body
	var requestData map[string]interface{}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Log dữ liệu request để xem cấu trúc của 'sub'
	log.Printf("Received requestData: %+v\n", requestData)

	// Kiểm tra xem request có chứa trường "field" và "value" hay không
	field, fieldOk := requestData["field"].(string)
	value, valueOk := requestData["value"]
	if !fieldOk || !valueOk {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid 'field' or 'value'"})
		return
	}

	// Log giá trị 'field' và 'value'
	log.Printf("Field: %s, Value: %+v\n", field, value)

	// Xử lý cho trường hợp sub là một mảng
	var updateData bson.M
	if field == "sub" {
		// Kiểm tra xem giá trị có phải là một mảng không
		subArray, ok := value.([]interface{})
		if !ok {
			log.Println("Invalid format for sub. Expected an array.")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid format for sub. It should be an array."})
			return
		}
		// Log giá trị của subArray
		log.Printf("SubArray: %+v\n", subArray)

		// Tạo điều kiện cập nhật cho sub (mảng)
		updateData = bson.M{
			"$set": bson.M{
				field: subArray,
			},
		}
	} else {
		// Tạo điều kiện cập nhật cho các trường khác
		updateData = bson.M{
			"$set": bson.M{
				field: value,
			},
		}
	}
	// Cập nhật trường trong MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := models.GetMovieCollection()
	filter := bson.M{"_id": movieID}
	_, err = collection.UpdateOne(ctx, filter, updateData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update movie field"})
		return
	}

	// Xóa cache trong Redis nếu có sử dụng
	dbs.DeleteCacheByKeyword(ctx, "movie")

	_, _, _, _, _, _, err = GetAllMoviesWithOptions(c, websocketServer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error fetching movies data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Movie field updated successfully",
	})
}

// Xử lý yêu cầu xóa ảnh
func DeleteMovieImage(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	// Lấy ID và tên file từ request
	movieID := c.PostForm("id")
	filename := c.PostForm("filename")

	// Chuyển movieID thành ObjectID
	oid, err := primitive.ObjectIDFromHex(movieID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid movie ID", "success": false})
		return
	}

	// Tạo context với timeout 5 giây
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Tìm và cập nhật movie trong DB
	collection := models.GetMovieCollection()
	var movie models.Movie
	err = collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&movie)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Movie not found", "success": false})
		return
	}

	// Xóa ảnh khỏi danh sách Moreimage
	newMoreImages := []string{}
	for _, img := range movie.Moreimage {
		if img != filename {
			newMoreImages = append(newMoreImages, img)
		}
	}

	// Cập nhật lại danh sách ảnh
	update := bson.M{"$set": bson.M{"moreimage": newMoreImages}}
	_, err = collection.UpdateOne(ctx, bson.M{"_id": oid}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update movie", "success": false})
		return
	}

	// Xóa file ảnh trên server
	filePath := fmt.Sprintf("views/admin/views/uploads/images/%s", filename)
	err = os.Remove(filePath)
	if err != nil {
		log.Println("Failed to delete image from server:", err)
	}

	dbs.DeleteCacheByKeyword(ctx, "movie")

	_, _, _, _, _, _, err = GetAllMoviesWithOptions(c, websocketServer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error fetching movies data"})
		return
	}

	// Trả về kết quả thành công
	c.JSON(http.StatusOK, gin.H{"message": "Image deleted successfully", "success": true})
}

func UpdateMoviePosition(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	// Nhận dữ liệu từ request body dưới dạng mảng JSON
	var movies []struct {
		ID       string `json:"ID"`
		Position int    `json:"Position"`
	}

	if err := c.ShouldBindJSON(&movies); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request format"})
		return
	}

	collection := models.GetMovieCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Duyệt qua danh sách các bộ phim và cập nhật vị trí
	for _, movie := range movies {
		oid, err := primitive.ObjectIDFromHex(movie.ID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid movie ID"})
			return
		}

		// Cập nhật vị trí mới cho bộ phim
		update := bson.M{"$set": bson.M{"position": movie.Position}}
		_, err = collection.UpdateOne(ctx, bson.M{"_id": oid}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update movie position"})
			return
		}
	}

	// Xóa cache Redis nếu có
	err := dbs.RedisClient.Del(ctx, "movies").Err()
	if err != nil {
		log.Println("Failed to clear Redis cache:", err)
	}

	_, _, _, _, _, _, err = GetAllMoviesWithOptions(c, websocketServer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error fetching movies data"})
		return
	}

	// Trả về phản hồi thành công
	c.JSON(http.StatusOK, gin.H{"message": "Movies positions updated successfully"})
}
