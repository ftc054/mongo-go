package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	err := godotenv.Load("../.devcontainer/.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func GetMongoDB_URL() string {
	return os.Getenv("MONGODB_URL")
}
