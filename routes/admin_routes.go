// routes/admin_routes.go
package routes

import (
	middleware "fire-watch/auth"
	controllers "fire-watch/controllers/admin"
	"fire-watch/websocket"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterAdminRoutes(router *gin.Engine, websocketServer *websocket.WebSocketServer) {
	// Nhóm các route cho admin
	adminRoutes := router.Group("/admin", middleware.AuthMiddleware())
	{
		adminRoutes.GET("/dashboard", func(c *gin.Context) {

			// Render giao diện dashboard
			c.HTML(http.StatusOK, "index.html", gin.H{
				"title":    "Admin Dashboard",
				"template": "dashboard",
			})
		})

		//quality
		//quality
		//quality
		adminRoutes.GET("/movies/:movieID/episodes/:episodeID/server/:serverID/qualities", controllers.GetQualityByMovieEpisodeServer)
		/// Route POST để thêm quality mới, truyền websocketQuality vào controller
		adminRoutes.POST("/add-quality", func(c *gin.Context) {
			controllers.AddQuality(c, websocketServer) // Truyền websocketQuality vào controller
		})
		// Route để cập nhật trường cụ thể của Movie
		adminRoutes.POST("/update-qulity-field/:id", func(c *gin.Context) {
			controllers.UpdateQualityField(c, websocketServer) // Truyền websocketServer vào controller
		})
		// /// Route POST để thêm quality mới, truyền websocketQuality vào controller
		// adminRoutes.POST("/update-quality", func(c *gin.Context) {
		// 	controllers.UpdateQuality(c, websocketServer) // Truyền websocketQuality vào controller
		// })
		/// Route DELETE để xóa quality mới, truyền websocketQuality vào controller
		adminRoutes.DELETE("/delete-quality/:id", func(c *gin.Context) {
			controllers.DeleteQuality(c, websocketServer) // Truyền websocketQuality vào controller
		})

		//episode
		//episode
		//episode
		adminRoutes.GET("/movies/:movieID/episodes", controllers.GetEpisodesByMovieID)
		/// Route POST để thêm episode mới, truyền websocketEpisode vào controller
		adminRoutes.POST("/add-episode", func(c *gin.Context) {
			controllers.AddEpisode(c, websocketServer) // Truyền websocketEpisode vào controller
		})
		/// Route POST để thêm episode mới, truyền websocketEpisode vào controller
		adminRoutes.POST("/update-episode", func(c *gin.Context) {
			controllers.UpdateEpisode(c, websocketServer) // Truyền websocketEpisode vào controller
		})
		/// Route DELETE để xóa episode mới, truyền websocketEpisode vào controller
		adminRoutes.DELETE("/delete-episode/:id", func(c *gin.Context) {
			controllers.DeleteEpisode(c, websocketServer) // Truyền websocketEpisode vào controller
		})
		//server
		//server
		//server
		adminRoutes.GET("/server", func(c *gin.Context) {
			servers, err := controllers.GetAllServers()
			if err != nil {
				c.String(http.StatusInternalServerError, "Error fetching servers")
				return
			}

			c.HTML(http.StatusOK, "index.html", gin.H{
				"title":    "Admin server List",
				"template": "server", // Đây là tên của template được định nghĩa
				"servers":  servers,
			})
		})
		adminRoutes.GET("/servers", func(c *gin.Context) {
			servers, err := controllers.GetAllServers()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching servers"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"servers": servers,
			})
		})
		/// Route POST để thêm server mới, truyền websocketServer vào controller
		adminRoutes.POST("/add-server", func(c *gin.Context) {
			controllers.AddServer(c, websocketServer) // Truyền websocketServer vào controller
		})
		adminRoutes.POST("/update-server/:id", func(c *gin.Context) {
			controllers.UpdateServer(c, websocketServer) // Truyền websocketServer vào controller
		})
		adminRoutes.DELETE("/delete-server/:id", func(c *gin.Context) {
			controllers.DeleteServer(c, websocketServer) // Truyền websocketServer vào controller
		})
		// Route để cập nhật trường cụ thể của Server
		adminRoutes.POST("/update-server-field/:id", func(c *gin.Context) {
			controllers.UpdateServerField(c, websocketServer) // Truyền websocketServer vào controller
		})

		//category
		//category
		//category
		adminRoutes.GET("/category", func(c *gin.Context) {
			categories, err := controllers.GetAllCategories()
			if err != nil {
				c.String(http.StatusInternalServerError, "Error fetching categories")
				return
			}

			c.HTML(http.StatusOK, "index.html", gin.H{
				"title":      "Admin category List",
				"template":   "category", // Đây là tên của template được định nghĩa
				"categories": categories,
			})
		})
		adminRoutes.GET("/categories", func(c *gin.Context) {
			categories, err := controllers.GetAllCategories()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching categories"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"categories": categories,
			})
		})
		/// Route POST để thêm category mới, truyền websocketServer vào controller
		adminRoutes.POST("/add-category", func(c *gin.Context) {
			controllers.AddCategory(c, websocketServer) // Truyền websocketServer vào controller
		})
		adminRoutes.POST("/update-category/:id", func(c *gin.Context) {
			controllers.UpdateCategory(c, websocketServer) // Truyền websocketServer vào controller
		})
		adminRoutes.DELETE("/delete-category/:id", func(c *gin.Context) {
			controllers.DeleteCategory(c, websocketServer) // Truyền websocketServer vào controller
		})
		// Route để cập nhật trường cụ thể của Category
		adminRoutes.POST("/update-category-field/:id", func(c *gin.Context) {
			controllers.UpdateCategoryField(c, websocketServer) // Truyền websocketServer vào controller
		})

		//Genre
		//Genre
		//Genre
		adminRoutes.GET("/genre", func(c *gin.Context) {
			genres, err := controllers.GetAllGenres()
			if err != nil {
				c.String(http.StatusInternalServerError, "Error fetching genres")
				return
			}

			c.HTML(http.StatusOK, "index.html", gin.H{
				"title":    "Admin genre List",
				"template": "genre", // Đây là tên của template được định nghĩa
				"genres":   genres,
			})
		})
		adminRoutes.GET("/genres", func(c *gin.Context) {
			genres, err := controllers.GetAllGenres()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching genres"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"genres": genres,
			})
		})
		/// Route POST để thêm genre mới, truyền websocketServer vào controller
		adminRoutes.POST("/add-genre", func(c *gin.Context) {
			controllers.AddGenre(c, websocketServer) // Truyền websocketServer vào controller
		})
		adminRoutes.POST("/update-genre/:id", func(c *gin.Context) {
			controllers.UpdateGenre(c, websocketServer) // Truyền websocketServer vào controller
		})
		adminRoutes.DELETE("/delete-genre/:id", func(c *gin.Context) {
			controllers.DeleteGenre(c, websocketServer) // Truyền websocketServer vào controller
		})
		// Route để cập nhật trường cụ thể của Genre
		adminRoutes.POST("/update-genre-field/:id", func(c *gin.Context) {
			controllers.UpdateGenreField(c, websocketServer) // Truyền websocketServer vào controller
		})

		//country
		//country
		//country
		adminRoutes.GET("/country", func(c *gin.Context) {
			countries, err := controllers.GetAllCountries()
			if err != nil {
				c.String(http.StatusInternalServerError, "Error fetching countries")
				return
			}

			c.HTML(http.StatusOK, "index.html", gin.H{
				"title":     "Admin country List",
				"template":  "country", // Đây là tên của template được định nghĩa
				"countries": countries,
			})
		})
		adminRoutes.GET("/countries", func(c *gin.Context) {
			countries, err := controllers.GetAllCountries()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching countries"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"countries": countries,
			})
		})
		/// Route POST để thêm country mới, truyền websocketServer vào controller
		adminRoutes.POST("/add-country", func(c *gin.Context) {
			controllers.AddCountry(c, websocketServer) // Truyền websocketServer vào controller
		})
		adminRoutes.POST("/update-country/:id", func(c *gin.Context) {
			controllers.UpdateCountry(c, websocketServer) // Truyền websocketServer vào controller
		})
		adminRoutes.DELETE("/delete-country/:id", func(c *gin.Context) {
			controllers.DeleteCountry(c, websocketServer) // Truyền websocketServer vào controller
		})
		// Route để cập nhật trường cụ thể của Country
		adminRoutes.POST("/update-country-field/:id", func(c *gin.Context) {
			controllers.UpdateCountryField(c, websocketServer) // Truyền websocketServer vào controller
		})

		//movie
		//movie
		//movie
		adminRoutes.GET("/movie", func(c *gin.Context) {
			movies, categories, genres, countries, episodes, servers, err := controllers.GetAllMoviesWithOptions(c, websocketServer)
			if err != nil {
				c.String(http.StatusInternalServerError, "Error fetching movies with options")
				return
			}

			// Truyền dữ liệu vào template HTML
			c.HTML(http.StatusOK, "index.html", gin.H{
				"title":      "Admin Movie List",
				"template":   "movie",    // Đây là tên của template được định nghĩa
				"movies":     movies,     // Danh sách phim
				"categories": categories, // Danh sách thể loại
				"genres":     genres,     // Danh sách thể loại phim
				"countries":  countries,  // Danh sách quốc gia
				"episodes":   episodes,   // Danh sách quốc gia
				"servers":    servers,    // Danh sách quốc gia
			})
		})
		adminRoutes.GET("/movies", func(c *gin.Context) {
			movies, categories, genres, countries, episodes, servers, err := controllers.GetAllMoviesWithOptions(c, websocketServer)
			if err != nil {
				c.String(http.StatusInternalServerError, "Error fetching movies with options")
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"movies":     movies,     // Danh sách phim
				"categories": categories, // Danh sách thể loại
				"genres":     genres,     // Danh sách thể loại phim
				"countries":  countries,  // Danh sách quốc gia
				"episodes":   episodes,   // Danh sách quốc gia
				"servers":    servers,
			})
		})
		/// Route POST để thêm country mới, truyền websocketServer vào controller
		adminRoutes.POST("/add-movie", func(c *gin.Context) {
			controllers.AddMovie(c, websocketServer) // Truyền websocketServer vào controller
		})
		adminRoutes.POST("/update-movie/:id", func(c *gin.Context) {
			controllers.UpdateMovie(c, websocketServer) // Truyền websocketServer vào controller
		})
		adminRoutes.POST("/movie-update-position", func(c *gin.Context) {
			controllers.UpdateMoviePosition(c, websocketServer) // Truyền websocketServer vào controller
		})
		adminRoutes.DELETE("/delete-movie/:id", func(c *gin.Context) {
			controllers.DeleteMovie(c, websocketServer) // Truyền websocketServer vào controller
		})
		// Route để cập nhật trường cụ thể của Movie
		adminRoutes.POST("/update-movie-field/:id", func(c *gin.Context) {
			controllers.UpdateMovieField(c, websocketServer) // Truyền websocketServer vào controller
		})
		// Route để cập nhật trường cụ thể của Movie
		adminRoutes.POST("/delete-movie-image", func(c *gin.Context) {
			controllers.DeleteMovieImage(c, websocketServer) // Truyền websocketServer vào controller
		})
	}
}
