package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Will return the uri to access mongdb database using godotenv
func EnvMongoUri() string {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found \n", err)
	}

	return os.Getenv("MONGODB_URI")
}