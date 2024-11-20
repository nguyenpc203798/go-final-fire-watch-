package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fire-watch/dbs"
	"fire-watch/models"
	"fire-watch/routes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	// Đảm bảo đã cài đặt thư viện này
	// Đảm bảo đã cài đặt thư viện này
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Khởi tạo movieCollection

// Mock Redis client
// setupMockRedis khởi tạo Redis client cho mục đích test
func TestCreateMovie(t *testing.T) {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize test database
	dbs.Connect()
	dbs.InitializeRedis()
	models.InitializeMovieCollection()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	routes.RegisterMovieRoutes(router) // Register routes

	// Mock movie data
	movie := models.Movie{
		Title:       "Test Movie",
		NameEng:     "Test Movie ENG",
		Description: "This is a test movie description.",
		Tags:        "test, movie",
		Status:      1,
		Image:       "test_image.jpg",
		Moreimage:   []string{"img1.jpg", "img2.jpg"},
		Slug:        "test-movie",
		Category:    []primitive.ObjectID{primitive.NewObjectID()},
		Genre:       []primitive.ObjectID{primitive.NewObjectID()},
		Country:     primitive.NewObjectID(),
		Episode:     []primitive.ObjectID{primitive.NewObjectID(), primitive.NewObjectID()},
		Hotmovie:    1,
		MaxQuality:  1080,
		Sub:         []string{"English", "Vietnamese"},
		Trailer:     "https://testtrailer.com/trailer.mp4",
		Year:        2024,
		Season:      1,
		Duration:    "2h30m",
		Rating:      8,
		Comment:     "This is a great movie.",
		Numofep:     12,
		Views:       12345,
		Position:    1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Deleted:     "",
	}

	// Serialize movie to JSON
	body, _ := json.Marshal(movie)

	// Create request
	req, _ := http.NewRequest("POST", "/movies/addmovie", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check HTTP response
	assert.Equal(t, http.StatusOK, rr.Code, "Expected status 200, got %d", rr.Code)

	// Parse response body
	var createdMovie models.Movie
	err = json.Unmarshal(rr.Body.Bytes(), &createdMovie)
	assert.NoError(t, err, "Error unmarshalling response")

	// Validate data
	assert.Equal(t, movie.Title, createdMovie.Title)
}

func TestGetAllMovies(t *testing.T) {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize test database
	dbs.Connect()
	dbs.InitializeRedis()
	models.InitializeMovieCollection()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	routes.RegisterMovieRoutes(router)

	// Insert mock movies into the database
	movieCollection := models.GetMovieCollection()
	movie1 := models.Movie{
		Title:     "Movie 1",
		Status:    1,
		Slug:      "movie-1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	movie2 := models.Movie{
		Title:     "Movie 2",
		Status:    1,
		Slug:      "movie-2",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, _ = movieCollection.InsertOne(context.TODO(), movie1)
	_, _ = movieCollection.InsertOne(context.TODO(), movie2)

	// Create request
	req, _ := http.NewRequest("GET", "/movies/getallmovies", nil)

	// Record response
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check HTTP response
	assert.Equal(t, http.StatusOK, rr.Code, "Expected status 200, got %d", rr.Code)

	// Parse response body
	var movies []models.Movie
	err = json.Unmarshal(rr.Body.Bytes(), &movies)
	assert.NoError(t, err, "Error unmarshalling response")
	assert.GreaterOrEqual(t, len(movies), 2, "Expected at least 2 movies")
}

func TestGetMovieByID(t *testing.T) {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize test database
	dbs.Connect()
	dbs.InitializeRedis()
	models.InitializeMovieCollection()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	routes.RegisterMovieRoutes(router)

	// Insert a mock movie into the database
	movieCollection := models.GetMovieCollection()
	mockMovie := models.Movie{
		ID:        primitive.NewObjectID(),
		Title:     "Test Movie",
		Status:    1,
		Slug:      "test-movie",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, _ = movieCollection.InsertOne(context.TODO(), mockMovie)

	// Create request
	req, _ := http.NewRequest("GET", "/movies/getmovie/"+mockMovie.ID.Hex(), nil)

	// Record response
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check HTTP response
	assert.Equal(t, http.StatusOK, rr.Code, "Expected status 200, got %d", rr.Code)

	// Parse response body
	var fetchedMovie models.Movie
	err = json.Unmarshal(rr.Body.Bytes(), &fetchedMovie)
	assert.NoError(t, err, "Error unmarshalling response")
	assert.Equal(t, mockMovie.Title, fetchedMovie.Title, "Expected movie titles to match")
}

func TestUpdateMovie(t *testing.T) {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize test database
	dbs.Connect()
	dbs.InitializeRedis()
	models.InitializeMovieCollection()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	routes.RegisterMovieRoutes(router)

	// Insert a mock movie into the database
	movieCollection := models.GetMovieCollection()
	mockMovie := models.Movie{
		ID:        primitive.NewObjectID(), // Khởi tạo ObjectID
		Title:     "Test Movie",
		Status:    1,
		Slug:      "test-movie",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, _ = movieCollection.InsertOne(context.TODO(), mockMovie)

	// Update movie data (bao gồm ID)
	updatedMovie := models.Movie{
		ID:     mockMovie.ID,         // Gán ID của mockMovie
		Title:  "Updated Test Movie", // Cập nhật title
		Status: 2,                    // Thay đổi status
	}
	body, _ := json.Marshal(updatedMovie) // Serialize thành JSON

	// Create PUT request với đúng ID trong URL
	req, _ := http.NewRequest("PUT", "/movies/updatemovie/"+mockMovie.ID.Hex(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Kiểm tra phản hồi HTTP
	assert.Equal(t, http.StatusOK, rr.Code, "Expected status 200, got %d", rr.Code)

	// Kiểm tra xem dữ liệu có được cập nhật trong DB hay không
	var result models.Movie
	_ = movieCollection.FindOne(context.TODO(), bson.M{"_id": mockMovie.ID}).Decode(&result)

	// So sánh kết quả
	assert.Equal(t, "Updated Test Movie", result.Title, "Expected movie title to be updated")
	assert.Equal(t, 2, result.Status, "Expected movie status to be updated")
}

func TestDeleteMovie(t *testing.T) {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize test database
	dbs.Connect()
	dbs.InitializeRedis()
	models.InitializeMovieCollection()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	routes.RegisterMovieRoutes(router)

	// Insert a mock movie into the database
	movieCollection := models.GetMovieCollection()
	mockMovie := models.Movie{
		ID:        primitive.NewObjectID(), // Tạo ID ngẫu nhiên cho mock movie
		Title:     "Test Movie2",
		Status:    1,
		Slug:      "test-movie2",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err = movieCollection.InsertOne(context.TODO(), mockMovie)
	if err != nil {
		t.Fatalf("Failed to insert mock movie: %v", err)
	}

	// Create DELETE request với ID chính xác
	req, err := http.NewRequest("DELETE", "/movies/deletemovie/"+mockMovie.ID.Hex(), nil)
	if err != nil {
		t.Fatalf("Failed to create DELETE request: %v", err)
	}

	// Record response
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Kiểm tra mã phản hồi HTTP
	assert.Equal(t, http.StatusOK, rr.Code, "Expected status 200, got %d", rr.Code)

	// Xác minh phim đã bị xóa khỏi cơ sở dữ liệu
	count, err := movieCollection.CountDocuments(context.TODO(), bson.M{"_id": mockMovie.ID})
	if err != nil {
		t.Fatalf("Failed to count documents: %v", err)
	}
	assert.Equal(t, int64(0), count, "Expected no documents with the given ID")
}
