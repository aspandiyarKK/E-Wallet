package internal

import (
	"context"
	"fmt"

	"EWallet/pkg/repository"

	"github.com/sirupsen/logrus"
)

type Storage interface {
	GetWallet(ctx context.Context, id int) (repository.Wallet, error)
	UpdateWallet(ctx context.Context, id int, wallet repository.Wallet) (repository.Wallet, error)
	DeleteWallet(ctx context.Context, id int) error
	CreateWallet(ctx context.Context, wallet repository.Wallet) (int, error)
}

type App struct {
	log   *logrus.Entry
	store Storage
}

func NewApp(log *logrus.Logger, store Storage) *App {
	return &App{
		log:   log.WithField("component", "service"),
		store: store,
	}
}

func (s *App) CreateWallet(ctx context.Context, wallet repository.Wallet) (int, error) {
	id, err := s.store.CreateWallet(ctx, wallet)
	if err != nil {
		return 0, fmt.Errorf("err inserting last_visit: %w", err)
	}
	return id, nil
}

func (s *App) GetWallet(ctx context.Context, id int) (repository.Wallet, error) {
	wal, err := s.store.GetWallet(ctx, id)
	if err != nil {
		return repository.Wallet{}, fmt.Errorf("err getting wallet : %w", err)
	}
	return wal, nil
}

func (s *App) DeleteWallet(ctx context.Context, id int) error {
	err := s.store.DeleteWallet(ctx, id)
	if err != nil {
		return fmt.Errorf("err deleting wallet : %w", err)
	}
	return nil
}

func (s *App) UpdateWallet(ctx context.Context, id int, wallet repository.Wallet) (repository.Wallet, error) {
	wal, err := s.store.UpdateWallet(ctx, id, wallet)
	if err != nil {
		return repository.Wallet{}, fmt.Errorf("err updating the Wallet: %w", err)
	}
	return wal, nil
}
