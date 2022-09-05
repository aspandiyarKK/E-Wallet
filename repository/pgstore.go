package repository

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"time"
)

type Wallet struct {
	Owner     string    `json:"owner"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type PG struct {
	log *logrus.Entry
	db  *sqlx.DB
	dsn string
}

func NewRepo(log *logrus.Logger, dsn string) (*PG, error) {
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("err connecting to PG : %w", err)
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("err pinging pg after initing connection: %w", err)
	}
	pg := &PG{
		log: log.WithField("component", "pgstore"),
		db:  db,
		dsn: dsn,
	}
	return pg, nil

}
func (pg *PG) CreateWallet(wallet Wallet) error {
	query := `INSERT INTO Wallet(id,owner,balance,createad_at,updated_at) VALUES (0,$1,$2,$3,$4)`
	_, err := pg.db.Exec(query, wallet.Owner, wallet.Balance, wallet.CreatedAt, wallet.UpdatedAt)
	if err != nil {
		return fmt.Errorf("err inserting last_visit: %w", err)
	}
	return nil

}

func (pg *PG) GetWallet() (Wallet, error) {
	query := `SELECT Owner,Balance,CreatedAt,UpdatedAt FROM Wallet WHERE id = 0`
	var wallet Wallet
	if err := pg.db.Get(&wallet, query); err != nil {
		return Wallet{}, fmt.Errorf("err getting wallet : %w", err)
	}
	return wallet, nil

}

func (pg *PG) UpdateWallet(wallet Wallet) error {
	query := `UPDATE users SET owner = $1, balance = $2,updated_at = $3, WHERE id = 0`
	_, err := pg.db.Exec(query, wallet.Owner, wallet.Balance, wallet.UpdatedAt)
	if err != nil {
		return fmt.Errorf("err updating the Wallet: %w", err)
	}
	return nil

}

func (pg *PG) DeleteWallet() error {
	query := `DELETE FROM Wallet WHERE id = 0`
	_, err := pg.db.Exec(query)

	if err != nil {
		fmt.Errorf("err deleting wallet : %w", err)
	}
	return nil
}
