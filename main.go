package main

import (
    "github.com/gin-gonic/gin"
    "CloudBox/controllers"
    "CloudBox/middlewares"
    "CloudBox/utils"
)

func main() {
    // Load env variables
    utils.LoadEnv()

    r := gin.Default()

    // Public routes
    auth := r.Group("/auth")
    {
        auth.POST("/register", controllers.CreateUser)
        auth.POST("/login", controllers.Login)
        auth.POST("/refresh", controllers.RefreshToken)
    }

    // Protected routes
    protected := r.Group("/api")
    protected.Use(middlewares.CheckAuth())
    {
        protected.GET("/profile", controllers.GetUserProfile)

        
        protected.POST("/files/upload", controllers.UploadFile)
        protected.GET("/files/list", controllers.ListFiles)
        protected.GET("/files/download/:id", controllers.DownloadFile)
    }

    r.Run()
}