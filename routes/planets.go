package routes

import (
	"github.com/gin-gonic/gin"

	"planetsAPI/controllers"
)

func PlanetRoute(router *gin.Engine) {
	router.GET("/", controllers.GetCollections())
	router.GET("/:collection_id/all", controllers.GetAllDocuments())
	router.GET("/:collection_id/", controllers.SearchesCollection())

	router.POST(":collection_id/add/", controllers.InsertOnePlanet())

	router.DELETE("/:collection_id/delete/", controllers.DeletePlanets())

	router.PATCH("/:collection_id/update/", controllers.UpdatePlanets())
}