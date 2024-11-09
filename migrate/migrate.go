package main

import (
	"CloudBox/initializers"
	"CloudBox/models"
)

func init() {
	initializers.LoadEnvs()
	initializers.ConnectDB()

}

func main() {

     initializers.DB.AutoMigrate(&models.User{})
}