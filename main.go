package main

import (
	"github.com/gin-gonic/gin"

	"planetsAPI/routes"
)

func main() {
	router := gin.Default()

	routes.PlanetRoute(router)

	router.Run("localhost:8888")
}