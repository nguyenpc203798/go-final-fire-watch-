// routes/genre_routes.go
package routes

import (
	"fire-watch/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterGenreRoutes(router *gin.Engine) {
	genreRoutes := router.Group("/genres")
	{
		genreRoutes.POST("/addgenre", controllers.AddGenre)
		genreRoutes.GET("/getallgenres", controllers.GetAllGenres)
		genreRoutes.GET("/getgenre/:id", controllers.GetGenreByID)
		genreRoutes.PUT("/updategenre/:id", controllers.UpdateGenre)
		genreRoutes.DELETE("/deletegenre/:id", controllers.DeleteGenre)
	}
}
