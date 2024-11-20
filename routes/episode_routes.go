// routes/episode_routes.go
package routes

import (
	"fire-watch/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterEpisodeRoutes(router *gin.Engine) {
	episodeRoutes := router.Group("/episodes")
	{
		episodeRoutes.POST("/addepisode", controllers.AddEpisode)
		episodeRoutes.GET("/getallepisodes", controllers.GetAllEpisodes)
		episodeRoutes.GET("/getepisode/:id", controllers.GetEpisodeByID)
		episodeRoutes.PUT("/updateepisode/:id", controllers.UpdateEpisode)
		episodeRoutes.DELETE("/deleteepisode/:id", controllers.DeleteEpisode)
	}
}
