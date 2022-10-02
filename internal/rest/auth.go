package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r *Router) authHandler(c *gin.Context) {
	var user UserInfo
	err := c.BindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if user.Username == "aspan" && user.Password == "12345" {
		tokenString, _ := r.GenToken(user.Username)
		c.JSON(http.StatusOK, tokenString)
		return
	} else {
		c.JSON(http.StatusUnauthorized, "authentication failed")
		return
	}
}
