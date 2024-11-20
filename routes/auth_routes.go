// routes/auth_routes.go
package routes

import (
	"fire-watch/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(router *gin.Engine) {
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/login", controllers.LoginUser)
		// Route GET /auth/login trả về trang đăng nhập
		authRoutes.GET("/login", func(c *gin.Context) {
			// Render trang sign-in.html
			c.HTML(200, "sign-in.html", gin.H{
				"title": "Sign In",
			})
		})
		authRoutes.POST("/register", controllers.RegisterUser)
		authRoutes.GET("/register", func(c *gin.Context) {
			// Render trang sign-in.html
			c.HTML(200, "sign-up.html", gin.H{
				"title": "Sign Up",
			})
		})
		authRoutes.GET("/getallusers", controllers.GetAllUsers)
		authRoutes.GET("/getuser/:id", controllers.GetUserByID)
		authRoutes.PUT("/updateuser/:id", controllers.UpdateUser)
		authRoutes.DELETE("/deleteuser/:id", controllers.DeleteUser)
	}
}
