package main

import (
	"fire-watch/controllers"
	"fire-watch/dbs"
	"fire-watch/models"
	"fire-watch/routes"
	"fire-watch/websocket"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	router := gin.Default()

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	router.Use(gin.RecoveryWithWriter(log.Writer()))
	// Khởi tạo WebSocket server
	websocketServer := websocket.NewWebSocketServer()

	go websocketServer.Run() // Chạy WebSocket server trong goroutine

	// Load template từ thư mục views/admin/layouts và views/admin/pages

	router.LoadHTMLGlob("views/**/**/*.html") // Chỉ load các file .html

	// Khai báo đường dẫn tĩnh cho thư mục uploads
	router.Static("/uploads", "./views/uploads")

	// Khai báo đường dẫn tĩnh cho assets (CSS, JS, hình ảnh)
	router.Static("/admin/assets", "./views/admin/assets") // Phục vụ các file tĩnh từ /admin/assets
	// Khai báo đường dẫn tĩnh cho assets (CSS, JS, hình ảnh)
	router.Static("/customer/assets", "./views/customer/assets") // Phục vụ các file tĩnh từ /customer/assets

	// Kết nối với MongoDB
	dbs.Connect()

	// Khởi tạo kết nối Redis
	dbs.InitializeRedis()

	// Khởi tạo collections cho các bảng cần thiết
	models.InitializeMovieCollection()
	models.InitializeCategoryCollection()
	models.InitializeServerCollection()
	models.InitializeCountryCollection() // Khởi tạo collection cho countries
	models.InitializeEpisodeCollection()
	models.InitializeQualityCollection()
	models.InitializeUserCollection() // Khởi tạo collection cho users
	controllers.InitializeEpisodeCollection()
	models.InitializeGenreCollection()     // Khởi tạo collection cho genres
	controllers.InitializeroleCollection() // Khởi tạo collection cho roles

	// Đăng ký WebSocket route
	router.GET("/ws", func(c *gin.Context) {
		websocketServer.HandleConnections(c.Writer, c.Request)
	})

	// Đăng ký các routes
	routes.RegisterMovieRoutes(router)
	routes.RegisterCategoryRoutes(router)
	routes.RegisterCountryRoutes(router)
	routes.RegisterGenreRoutes(router)
	routes.RegisterRoleRoutes(router)
	routes.RegisterAuthRoutes(router)
	routes.RegisterEpisodeRoutes(router)

	// Đăng ký các route và truyền websocketServer vào
	routes.RegisterAdminRoutes(router, websocketServer)
	routes.RegisterCustomerRoutes(router, websocketServer)

	// Khởi động server trên port 8080
	fmt.Println("Server is running on port 8080")
	router.Run(":8080") // Khởi động Gin server
}
