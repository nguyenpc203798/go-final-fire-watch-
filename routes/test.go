package routes

import (
	"fire-watch/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	// Định nghĩa các route cho sách
	router.GET("/test/movies", controllers.GetAllMovies)
	router.GET("/test/movies/:id", controllers.GetMovieByID)
	router.POST("/test/movies", controllers.AddMovie)
	router.PUT("/test/movies/:id", controllers.UpdateMovie)
	router.DELETE("/test/movies/:id", controllers.DeleteMovie)

	return router
}
