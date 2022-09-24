package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"EWallet/internal"
	"EWallet/internal/rest"
	"EWallet/pkg/logger"
	"EWallet/pkg/repository"

	_ "github.com/jackc/pgx/v4/stdlib"
	migrate "github.com/rubenv/sql-migrate"
)

const port = 3000

var (
	pgDSN = os.Getenv("PG_DSN")
	addr  = fmt.Sprintf("localhost:%d", port)
)

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
	exch := internal.NewExchangeRate(log)
	app := internal.NewApp(log, pg, exch)
	r := rest.NewRouter(log, app)
	if err = r.Run(ctx, addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Panicf("Error starting server: %v", err)
	}
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT)
	<-sigCh
	cancel()
	pg.Close()
	log.Info("Shutting down")
}
