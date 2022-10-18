package configs

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connects to database and returns mongo.Client
func ConnectDB() *mongo.Client {
	client, err := mongo.NewClient(ctx, options.Client().EnvMongoUri) //connects to mongodb
	if err != nil {
		panic(err)
	}

	return client
}