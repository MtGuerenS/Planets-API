package configs

import (
	"context"
	"log"
	"time"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connects to database and returns mongo.Client
func ConnectDB() *mongo.Client {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	client, err := mongo.NewClient(options.Client().ApplyURI(EnvMongoUri())) //connects to mongodb
	if err != nil {
		log.Fatal("Unable to create mongodb client \n", err)
	}

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal("Unable to connect to mongodb client \n", err)
	}

	fmt.Println("Connected to the database")

	return client
}

var CLIENT *mongo.Client = ConnectDB()

func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	return client.Database("sample_guides").Collection(collectionName)
}