package middlewares

import (
	"CloudBox/models"
	"CloudBox/utils"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func CheckAuth() gin.HandlerFunc {
    utils.LoadEnv()
    db := utils.ConnectDB()
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
            return
        }

        bearerToken := strings.Split(authHeader, " ")
        if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
            return
        }

        token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            return []byte(os.Getenv("SECRET")), nil
        })

        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            return
        }

        if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
            // Check token type
            if claims["type"] != "access" {
                c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token type"})
                return
            }

            var user models.User
            if err := db.First(&user, claims["user_id"]).Error; err != nil {
                c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
                return
            }

            c.Set("currentUser", user)
            c.Next()
        } else {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
        }
    }
}