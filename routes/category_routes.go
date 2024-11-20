// routes/category_routes.go
package routes

import (
	"fire-watch/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterCategoryRoutes(router *gin.Engine) {
	categoryRoutes := router.Group("/categories")
	{
		categoryRoutes.POST("/addcategory", controllers.AddCategory)
		categoryRoutes.GET("/getallcategories", controllers.GetAllCategories)
		categoryRoutes.GET("/getcategory/:id", controllers.GetCategoryByID)
		categoryRoutes.PUT("/updatecategory/:id", controllers.UpdateCategory)
		categoryRoutes.DELETE("/deletecategory/:id", controllers.DeleteCategory)
	}
}
