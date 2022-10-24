package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"EWallet/pkg/exchange"

	"EWallet/internal"
	"EWallet/internal/rest"
	"EWallet/pkg/logger"
	"EWallet/pkg/repository"

	_ "github.com/jackc/pgx/v4/stdlib"
	migrate "github.com/rubenv/sql-migrate"
)

const port = 3000

var (
	pgDSN  = os.Getenv("PG_DSN")
	addr   = fmt.Sprintf("localhost:%d", port)
	xrHost = os.Getenv("XR_HOST")
	apiKey = os.Getenv("API_KEY")
	secret = os.Getenv("SECRET_JWT")
)

// @title         E-wallet API
// @version         1.0
// @description     This is a api server for E-wallet application.

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @host      localhost:3000
// @BasePath  /api/v1

func main() {
	log := logger.NewLogger()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pg, err := repository.NewRepo(ctx, log, pgDSN)
	if err != nil {
		log.Panicf("Failed to connect to database: %v", err)
	}

	if err = pg.Migrate(migrate.Up); err != nil {
		log.Panicf("err migrating pg: %v", err)
	}
	exch := exchange.NewExchangeRate(log, xrHost, apiKey)
	app := internal.NewApp(log, pg, exch)
	r := rest.NewRouter(log, app, secret)
	go func() {
		if err = r.Run(ctx, addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Panicf("Error starting server: %v", err)
		}
	}()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT)
	<-sigCh
	cancel()
	pg.Close()
	log.Info("Shutting down")
}
