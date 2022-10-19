package controllers

import (
	"context"
	"log"
	"net/http"
	"planetsAPI/configs"
	"time"

	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/bson/primitive"
	// "go.mongodb.org/mongo-driver/mongo"
)

//Returns the databases
func getDatabase() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

		databases, err := configs.CLIENT.ListDatabaseNames(ctx, bson.D{})
		if err != nil {
			log.Fatal(err)
			c.JSON(http.StatusInternalServerError, err.Error())
		}

		c.IndentedJSON(http.StatusOK, databases)
	}
}