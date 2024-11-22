package main

import (
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"
    "CloudBox/controllers"
    "CloudBox/middlewares"
    "CloudBox/utils"
    "time"
)

func main() {
    // Load env variables
    utils.LoadEnv()

    r := gin.Default()

    // CORS
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"https://cloudbox-seven.vercel.app", "http://localhost:3000"},
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Refresh-Token"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))

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
