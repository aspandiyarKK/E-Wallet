package rest

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

const TokenExpireDuration = time.Hour

var secret = []byte("Goal:Senior in 2 years")

type MyClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func ParseToken(tokenString string) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token: %w", err)
}

func jwtAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, "Unauthorized")
			return
		}
		parts := strings.Split(authHeader, " ")
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, "Unauthorized")
			return
		}
		id, err := ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, "Unauthorized")
			return
		}
		c.Set("id", id)
		c.Next()
	}
}
