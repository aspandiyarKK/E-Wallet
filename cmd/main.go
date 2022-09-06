package main

import (
	"EWallet/internal/rest"
	"EWallet/pkg/logger"
	"EWallet/pkg/repository"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var pgDSN = os.Getenv("PG_DSN")
var addr = "localhost:8080"

func main() {
	log := logger.NewLogger()
	pg, err := repository.NewRepo(log, pgDSN)
	if err != nil {
		log.Panicf("Failed to connect to database: %v", err)
	}
	r := rest.NewRouter(log, pg)
	if err := r.Run(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Panicf("Error starting server: %v", err)
	}
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT)
	<-sigCh
	log.Info("Shutting down")
}
