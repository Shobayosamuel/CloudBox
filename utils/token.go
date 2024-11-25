package utils

import (
    "os"
    "time"
    "github.com/golang-jwt/jwt/v4"
)

type TokenResponse struct {
    AccessToken  string    `json:"access_token"`
    RefreshToken string    `json:"refresh_token"`
    ExpiresAt    time.Time `json:"expires_at"`
}

func GenerateTokens(userID uint) (TokenResponse, error) {
    // Access token - short lived (15 minutes)
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(150 * time.Minute).Unix(),
        "iat":     time.Now().Unix(),
        "type":    "access",
    })

    accessTokenString, err := accessToken.SignedString([]byte(os.Getenv("SECRET")))
    if err != nil {
        return TokenResponse{}, err
    }

    // Refresh token - longer lived (7 days)
    refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
        "iat":     time.Now().Unix(),
        "type":    "refresh",
    })

    refreshTokenString, err := refreshToken.SignedString([]byte(os.Getenv("SECRET")))
    if err != nil {
        return TokenResponse{}, err
    }

    return TokenResponse{
        AccessToken:  accessTokenString,
        RefreshToken: refreshTokenString,
        ExpiresAt:    time.Now().Add(15 * time.Minute),
    }, nil
}