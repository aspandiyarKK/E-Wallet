package rest

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func authHandler(c *gin.Context) {
	var user UserInfo
	err := c.BindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if user.Username == "aspan" && user.Password == "12345" {
		tokenString, _ := GenToken(user.Username)
		c.JSON(http.StatusOK, tokenString)
		return
	} else {
		c.JSON(http.StatusUnauthorized, "authentication failed")
		return
	}
}
