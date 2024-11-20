package utils

import (
    "log"
    "github.com/joho/godotenv"
    "os"
)

func LoadEnv() {
    if os.Getenv("ENV") != "production" {
        err := godotenv.Load()
        if err != nil {
            log.Fatalf("Error loading .env file: %v", err)
        }
    }
}