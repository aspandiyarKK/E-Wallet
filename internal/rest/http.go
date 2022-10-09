package rest

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"EWallet/pkg/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Router struct {
	log    *logrus.Entry
	router *gin.Engine
	app    App
	secret []byte
}

type App interface {
	GetWallet(ctx context.Context, id int, currency string) (repository.Wallet, error)
	UpdateWallet(ctx context.Context, id int, wallet repository.Wallet) (repository.Wallet, error)
	DeleteWallet(ctx context.Context, id int) error
	CreateWallet(ctx context.Context, wallet repository.Wallet) (int, error)
	Deposit(ctx context.Context, id int, request *repository.FinRequest) error
	Withdrawal(ctx context.Context, id int, request *repository.FinRequest) error
	Transfer(ctx context.Context, id int, request *repository.FinRequest) error
	GetTransactions(ctx context.Context, id int, order string) (*[]repository.Transaction, error)
}

func NewRouter(log *logrus.Logger, app App, secret string) *Router {
	r := &Router{
		log:    log.WithField("component", "router"),
		router: gin.Default(),
		app:    app,
		secret: []byte(secret),
	}
	r.router.GET("/metrics", prometheusHandler())
	r.router.POST("/auth", r.authHandler)
	g := r.router.Group("/api/v1").Use(r.jwtAuth())
	g.GET("/wallet/:id", r.getWallet)
	g.GET("/transaction/:id", r.transaction)
	g.POST("/wallet", r.addWallet)
	g.DELETE("/wallet/:id", r.deleteWallet)
	g.PUT("/wallet/:id", r.updateWallet)
	g.PUT("/wallet/:id/deposit", r.deposit)
	g.PUT("/wallet/:id/withdraw", r.withdrawal)
	g.PUT("/wallet/:id/transfer", r.transfer)
	return r
}

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func (r *Router) Run(_ context.Context, addr string) error {
	return r.router.Run(addr)
}

func (r *Router) addWallet(c *gin.Context) {
	var input repository.Wallet
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	id, err := r.app.CreateWallet(c, input)
	if err != nil {
		r.log.Errorf("failed to store date: %v", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (r *Router) getWallet(c *gin.Context) {
	val := c.Param("id")
	currency := c.Query("currency")
	id, err := strconv.Atoi(val)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	w, err := r.app.GetWallet(c, id, currency)
	switch {
	case err == nil:
	case errors.Is(err, repository.ErrWalletNotFound):
		c.JSON(http.StatusNotFound, err)
		return
	default:
		r.log.Errorf("failed to get Wallet: %v", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, w)
}

func (r *Router) deleteWallet(c *gin.Context) {
	val := c.Param("id")
	id, err := strconv.Atoi(val)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	err = r.app.DeleteWallet(c, id)
	switch {
	case err == nil:
	case errors.Is(err, repository.ErrWalletNotFound):
		c.JSON(http.StatusNotFound, err)
		return
	default:
		r.log.Errorf("failed to delete wallet %v: ", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (r *Router) updateWallet(c *gin.Context) {
	val := c.Param("id")
	id, err := strconv.Atoi(val)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	var wallet repository.Wallet
	if err = c.BindJSON(&wallet); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	wallet, err = r.app.UpdateWallet(c, id, wallet)
	switch {
	case err == nil:
	case errors.Is(err, repository.ErrWalletNotFound):
		c.JSON(http.StatusNotFound, err)
		return
	default:
		r.log.Errorf("failed to update wallet: %v", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, wallet)
}

func (r *Router) deposit(c *gin.Context) {
	val := c.Param("id")
	id, err := strconv.Atoi(val)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	var input repository.FinRequest
	err = c.BindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if !isValidUUID(input.UUID) {
		c.JSON(http.StatusBadRequest, "incorrect format of uuid")
		return
	}
	err = r.app.Deposit(c, id, &input)
	switch {
	case err == nil:
	case errors.Is(err, repository.ErrDuplicateKey):
		c.JSON(http.StatusConflict, err)
		return
	case errors.Is(err, repository.ErrWalletNotFound):
		c.JSON(http.StatusNotFound, err)
		return
	default:
		r.log.Errorf("failed to deposit wallet: %v", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, "Ok")
}

func (r *Router) withdrawal(c *gin.Context) {
	val := c.Param("id")
	id, err := strconv.Atoi(val)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	var input repository.FinRequest
	err = c.BindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if !isValidUUID(input.UUID) {
		c.JSON(http.StatusBadRequest, "incorrect format of uuid")
		return
	}
	err = r.app.Withdrawal(c, id, &input)
	switch {
	case err == nil:
	case errors.Is(err, repository.ErrDuplicateKey):
		c.JSON(http.StatusConflict, err)
		return
	case errors.Is(err, repository.ErrInsufficientFunds):
		c.JSON(http.StatusBadRequest, err)
		return
	case errors.Is(err, repository.ErrWalletNotFound):
		c.JSON(http.StatusNotFound, repository.ErrWalletNotFound)
		return
	default:
		r.log.Errorf("failed to withdraw from wallet: %v", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, "Ok")
}

func (r *Router) transfer(c *gin.Context) {
	val := c.Param("id")
	id, err := strconv.Atoi(val)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	var input repository.FinRequest
	err = c.BindJSON(&input)
	if err != nil || input.Sum <= 0 {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if !isValidUUID(input.UUID) {
		c.JSON(http.StatusBadRequest, "incorrect format of uuid")
		return
	}
	if err = r.app.Transfer(c, id, &input); err != nil {
		switch {
		case err == nil:
		case errors.Is(err, repository.ErrDuplicateKey):
			c.JSON(http.StatusConflict, err)
			return
		case errors.Is(err, repository.ErrInsufficientFunds):
			c.JSON(http.StatusBadRequest, err)
			return
		case errors.Is(err, repository.ErrWalletNotFound):
			c.JSON(http.StatusNotFound, repository.ErrWalletNotFound)
			return
		default:
			r.log.Errorf("failed to transfer money: %v", err)
			c.JSON(http.StatusInternalServerError, err)
			return
		}
	}
	c.JSON(http.StatusOK, "Success transferring")
}

func (r *Router) transaction(c *gin.Context) {
	val := c.Param("id")
	order := c.Query("order")
	id, err := strconv.Atoi(val)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	trans, err := r.app.GetTransactions(c, id, order)
	switch {
	case err == nil:
	case errors.Is(err, repository.ErrTransactionNotFound):
		c.JSON(http.StatusNotFound, err)
		return
	default:
		r.log.Errorf("failed to get Transactions: %v", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, trans)
}
