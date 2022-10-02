package rest

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

const (
	TokenExpireDuration = time.Hour * 1000
	sessionKey          = "session"
)

type MyClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type UserSession struct {
	Username string
}

func (r *Router) GenToken(username string) (string, error) {
	c := MyClaims{
		username,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(TokenExpireDuration).Unix(),
			Issuer:    "e-wallet",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString(r.secret)
}

func (r *Router) ParseToken(tokenString string) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		return r.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token: %w", err)
}

func (r *Router) jwtAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}
		parts := strings.Split(authHeader, " ")
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}
		claims, err := r.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
			c.Abort()
			return
		}
		u := UserSession{
			Username: claims.Username,
		}
		c.Set(sessionKey, &u)
		c.Next()
	}
}

func (r *Router) GetUserSession(c *gin.Context) *UserSession {
	u, ok := c.Get(sessionKey)
	if !ok {
		r.log.Errorf("session not found in context")
		return &UserSession{}
	}
	us, ok := u.(*UserSession)
	if !ok {
		r.log.Errorf("invalid session type")
		return &UserSession{}
	}
	return us
}
