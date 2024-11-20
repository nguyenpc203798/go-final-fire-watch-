package dbs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return err
}

func GetEnv(key string) string {
	return os.Getenv(key)
}
