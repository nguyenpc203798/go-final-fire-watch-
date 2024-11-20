// routes/movie_routes.go
package routes

import (
	"fire-watch/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterMovieRoutes(router *gin.Engine) {
	movieRoutes := router.Group("/movies")
	{
		movieRoutes.POST("/addmovie", controllers.AddMovie)
		movieRoutes.GET("/getallmovies", controllers.GetAllMovies)
		movieRoutes.GET("/getmovie/:id", controllers.GetMovieByID)
		movieRoutes.PUT("/updatemovie/:id", controllers.UpdateMovie)
		movieRoutes.DELETE("/deletemovie/:id", controllers.DeleteMovie)
		movieRoutes.GET("/getmoviewithepisode/:id", controllers.GetMovieWithEpisodes)
		movieRoutes.POST("/api/movies/bulk", controllers.CreateMoviesBulk)
	}
}
