// routes/country_routes.go
package routes

import (
	"fire-watch/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterCountryRoutes(router *gin.Engine) {
	countryRoutes := router.Group("/countries")
	{
		countryRoutes.POST("/addcountry", controllers.AddCountry)
		countryRoutes.GET("/getallcountries", controllers.GetAllCountries)
		countryRoutes.GET("/getcountry/:id", controllers.GetCountryByID)
		countryRoutes.PUT("/updatecountry/:id", controllers.UpdateCountry)
		countryRoutes.DELETE("/deletecountry/:id", controllers.DeleteCountry)
	}
}
