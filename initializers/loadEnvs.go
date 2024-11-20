package initializers

import (
	"log"
	"os"
	"github.com/joho/godotenv"
)

func LoadEnvs() {
	if os.Getenv("ENV") != "production" {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
	}

}