// routes/role_routes.go
package routes

import (
	"fire-watch/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterRoleRoutes(router *gin.Engine) {
	roleRoutes := router.Group("/roles")
	{
		roleRoutes.POST("/addrole", controllers.AddRole)
		roleRoutes.GET("/getallroles", controllers.GetAllRoles)
		roleRoutes.GET("/getrole/:id", controllers.GetRoleByID)
		roleRoutes.PUT("/updaterole/:id", controllers.UpdateRole)
		roleRoutes.DELETE("/deleterole/:id", controllers.DeleteRole)
	}
}
