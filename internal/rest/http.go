package rest

import (
	"EWallet/pkg/repository"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

type Router struct {
	log    *logrus.Entry
	router *gin.Engine
	app    App
}
type App interface {
	GetWallet() (repository.Wallet, error)
	UpdateWallet(wallet repository.Wallet) error
	DeleteWallet() error
	CreateWallet(wallet repository.Wallet) error
}

func NewRouter(log *logrus.Logger, app App) *Router {
	r := &Router{
		log:    log.WithField("component", "router"),
		router: gin.Default(),
	}
	r.router.GET("/getWallet", r.getWallet)
	r.router.POST("/addWallet", r.addWallet)
	r.router.DELETE("/deleteWallet", r.deleteWallet)
	r.router.PUT("/updateWallet")
	return r
}
func (r *Router) Run(addr string) error {
	return r.router.Run(addr)
}
func (r *Router) addWallet(c *gin.Context) {
	var input repository.Wallet
	if err := c.BindJSON(&input); err != nil {
		r.log.Errorf("invalid input %w", err)
		return
	}
	if err := r.app.CreateWallet(input); err != nil {
		r.log.Errorf("failed to store date: %v", err)
	}
}
func (r *Router) getWallet(c *gin.Context) {
	w, err := r.app.GetWallet()
	if err != nil {
		r.log.Errorf("failed to get Wallet")
	}
	c.JSON(http.StatusOK, w)

}

func (r *Router) deleteWallet(c *gin.Context) {
	err := r.app.DeleteWallet()
	if err != nil {
		r.log.Errorf("failed to delete wallet %w : ", err)
	}
	c.JSON(http.StatusOK, nil)
}

func (r *Router) updateWallet(c *gin.Context) {
	var input repository.Wallet
	if err := c.BindJSON(&input); err != nil {
		r.log.Errorf("invalid input %w ", err)
		return
	}
	if err := r.app.UpdateWallet(input); err != nil {
		r.log.Errorf("failed to update wallet: %v", err)
	}

	c.JSON(http.StatusOK, nil)

}
