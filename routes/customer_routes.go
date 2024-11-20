// routes/cutomer_routes.go
package routes

import (
	controllers "fire-watch/controllers/customer"
	"fire-watch/websocket"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterCustomerRoutes(router *gin.Engine, websocketServer *websocket.WebSocketServer) {
	router.GET("/home", func(c *gin.Context) {
		// Gọi hàm lấy danh sách phim
		movies, err := controllers.GetAllMovies(c)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error fetching movies with options")
			return
		}

		// Gọi hàm lấy danh mục kèm danh sách phim
		categorieswithmovie, err := controllers.GetCategoriesWithMovies(c)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error fetching categories with movies")
			return
		}

		// Gọi hàm lấy thông tin người dùng từ Redis
		user, err := controllers.GetUserFromRedis(c)
		if err != nil {
			// Nếu không lấy được user, đặt giá trị mặc định
			user = map[string]interface{}{
				"username": "Guest",
				"role":     "visitor",
			}
		}

		// Render giao diện dashboard và truyền dữ liệu cần thiết
		c.HTML(http.StatusOK, "customer.html", gin.H{
			"title":               "Customer Home",
			"template":            "home",
			"movies":              movies,              // Danh sách phim
			"categorieswithmovie": categorieswithmovie, // Danh sách danh mục kèm phim
			"user":                user,                // Danh sách danh mục kèm phim
		})
	})

	router.GET("/search", func(c *gin.Context) {
		movies, err := controllers.SearchMovies(c)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error fetching movie: %v", err))
			return
		}

		c.HTML(http.StatusOK, "customer.html", gin.H{
			"title":    "search",
			"template": "search",
			"movies":   movies, // Danh sách phim
			// Danh sách danh mục kèm phim
		})
	})

	router.GET("/movie/:id", func(c *gin.Context) {
		movie, err := controllers.GetMoviesDetail(c)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error fetching movie: %v", err))
			return
		}

		// Render HTML với dữ liệu movie
		c.HTML(http.StatusOK, "movie-detail.html", gin.H{
			"title": "Movie Detail",
			"movie": movie,
		})
	})
	router.GET("/movies/:id", func(c *gin.Context) {
		movie, err := controllers.GetMoviesDetail(c)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error fetching movie: %v", err))
			return
		}

		// Render HTML với dữ liệu movie
		c.JSON(http.StatusOK, gin.H{
			"movie": movie, // Danh sách phim
		})
	})

	router.GET("/movies", func(c *gin.Context) {
		movies, err := controllers.GetAllMovies(c)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error fetching movies with options")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"movies": movies, // Danh sách phim
		})
	})

	router.GET("/categories-movies", func(c *gin.Context) {
		categorieswithmovie, err := controllers.GetCategoriesWithMovies(c)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error fetching categorieswithmovie with options")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"categorieswithmovie": categorieswithmovie, // Danh sách phim
		})
	})
}
