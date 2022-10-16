package internal

import (
	"EWallet/pkg/models"
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
	Deposit(ctx context.Context, id int, request *repository.FinRequest) error
	Withdrawal(ctx context.Context, id int, request *repository.FinRequest) error
	Transfer(ctx context.Context, id int, request *repository.FinRequest) error
	GetTransactions(ctx context.Context, id int, params *models.TransactionQueryParams) ([]repository.Transaction, error)
}
type Exchange interface {
	GetRate(ctx context.Context, currency string, amount float64) (float64, error)
}

type App struct {
	log      *logrus.Entry
	store    Storage
	exchange Exchange
}

func NewApp(log *logrus.Logger, store Storage, exchange Exchange) *App {
	return &App{
		log:      log.WithField("component", "service"),
		store:    store,
		exchange: exchange,
	}
}

func (s *App) CreateWallet(ctx context.Context, wallet repository.Wallet) (int, error) {
	id, err := s.store.CreateWallet(ctx, wallet)
	if err != nil {
		return 0, fmt.Errorf("err inserting last_visit: %w", err)
	}
	return id, nil
}

func (s *App) GetWallet(ctx context.Context, id int, currency string) (repository.Wallet, error) {
	wal, err := s.store.GetWallet(ctx, id)
	if err != nil {
		return repository.Wallet{}, fmt.Errorf("err getting wallet : %w", err)
	}
	if currency != "" {
		wal.Balance, err = s.exchange.GetRate(ctx, currency, wal.Balance)
		if err != nil {
			return repository.Wallet{}, fmt.Errorf("err converting currency : %w", err)
		}
	}
	return wal, nil
}

func (s *App) GetRate(ctx context.Context, currency string) (float64, error) {
	return s.exchange.GetRate(ctx, currency, 1)
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

func (s *App) Deposit(ctx context.Context, id int, request *repository.FinRequest) error {
	err := s.store.Deposit(ctx, id, request)
	if err != nil {
		return fmt.Errorf("err depositing the Wallet: %w", err)
	}
	return nil
}

func (s *App) Withdrawal(ctx context.Context, id int, request *repository.FinRequest) error {
	if err := s.store.Withdrawal(ctx, id, request); err != nil {
		return fmt.Errorf("err withdrawing from the wallet: %w", err)
	}
	return nil
}

func (s *App) Transfer(ctx context.Context, id int, request *repository.FinRequest) error {
	err := s.store.Transfer(ctx, id, request)
	if err != nil {
		return fmt.Errorf("err transferring the wallet: %w", err)
	}
	return nil
}

func (s *App) GetTransactions(ctx context.Context, id int, params *models.TransactionQueryParams) ([]repository.Transaction, error) {
	trans, err := s.store.GetTransactions(ctx, id, params)
	if err != nil {
		return nil, fmt.Errorf("err getting the transactions: %w", err)
	}
	return trans, nil
}
