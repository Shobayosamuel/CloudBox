package main

import (
	"log"
	"CloudBox/initializers"
	"CloudBox/models"
	"CloudBox/utils"
)

func init() {
	initializers.LoadEnvs()
	initializers.ConnectDB()

}

func main() {
    db := utils.ConnectDB()
    err := db.AutoMigrate(&models.User{})
    if err != nil {
        log.Fatal(err)
    }
}