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
	Deposit(ctx context.Context, request *repository.Finrequest) error
	Withdrawal(ctx context.Context, request *repository.Finrequest) error
	Transfer(ctx context.Context, request *repository.Finrequest) error
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
func (s *App) Deposit(ctx context.Context, request *repository.Finrequest) error {
	err := s.store.Deposit(ctx, request)
	if err != nil {
		return fmt.Errorf("err depositing the Wallet: %w", err)
	}
	return nil
}
func (s *App) Withdrawal(ctx context.Context, request *repository.Finrequest) error {
	wal, err := s.store.GetWallet(ctx, request.ID)
	if wal.Balance < request.Sum {
		return fmt.Errorf("err not enough money")
	}
	err = s.store.Withdrawal(ctx, request)
	if err != nil {
		return fmt.Errorf("err withdrawaling the Wallet: %w", err)
	}
	return nil
}

func (s *App) Transfer(ctx context.Context, request *repository.Finrequest) error {
	err := s.store.Transfer(ctx, request)
	if err != nil {
		return fmt.Errorf("err transfering the wallet: %w", err)
	}
	return nil
}
