package rest

import (
	"EWallet/pkg/repository"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type Router struct {
	log    *logrus.Entry
	router *gin.Engine
	app    App
}

type App interface {
	GetWallet(ctx context.Context, id int) (repository.Wallet, error)
	UpdateWallet(ctx context.Context, id int, wallet repository.Wallet) (repository.Wallet, error)
	DeleteWallet(ctx context.Context, id int) error
	CreateWallet(ctx context.Context, wallet repository.Wallet) (int, error)
}

func NewRouter(log *logrus.Logger, app App) *Router {
	r := &Router{
		log:    log.WithField("component", "router"),
		router: gin.Default(),
		app:    app,
	}
	g := r.router.Group("/api/v1")
	g.GET("/wallet/:id", r.getWallet)
	g.POST("/wallet", r.addWallet)
	g.DELETE("/wallet/:id", r.deleteWallet)
	g.PUT("/wallet/:id", r.updateWallet)
	return r
}

func (r *Router) Run(_ context.Context, addr string) error {
	return r.router.Run(addr)
}

func (r *Router) addWallet(c *gin.Context) {
	var input repository.Wallet
	if err := c.BindJSON(&input); err != nil {
		r.log.Errorf("invalid input: %v", err)
		return
	}
	id, err := r.app.CreateWallet(c, input)
	if err != nil {
		r.log.Errorf("failed to store date: %v", err)
		c.JSON(http.StatusInternalServerError, nil)
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (r *Router) getWallet(c *gin.Context) {
	val := c.Param("id")
	id, err := strconv.Atoi(val)
	if err != nil {
		c.JSON(http.StatusBadRequest, nil)
		return
	}
	w, err := r.app.GetWallet(c, id)
	if err != nil {
		r.log.Errorf("failed to get Wallet: %v", err)
	}
	c.JSON(http.StatusOK, w)
}

func (r *Router) deleteWallet(c *gin.Context) {
	val := c.Param("id")
	id, err := strconv.Atoi(val)
	if err != nil {
		c.JSON(http.StatusBadRequest, nil)
		return
	}
	err = r.app.DeleteWallet(c, id)
	if err != nil {
		r.log.Errorf("failed to delete wallet %v: ", err)
	}
	c.JSON(http.StatusNoContent, nil)
}

func (r *Router) updateWallet(c *gin.Context) {
	val := c.Param("id")
	id, err := strconv.Atoi(val)
	if err != nil {
		c.JSON(http.StatusBadRequest, nil)
		return
	}
	var wallet repository.Wallet
	if err = c.BindJSON(&wallet); err != nil {
		c.JSON(http.StatusBadRequest, nil)
		return
	}
	if wallet, err = r.app.UpdateWallet(c, id, wallet); err != nil {
		r.log.Errorf("failed to update wallet: %v", err)
		return
	}
	c.JSON(http.StatusOK, wallet)
}
