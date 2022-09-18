package repository

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	migrate "github.com/rubenv/sql-migrate"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

//go:embed migrations
var migrations embed.FS

type Wallet struct {
	Owner     string    `json:"owner" db:"owner"`
	Balance   float64   `json:"balance" db:"balance"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
type Finrequest struct {
	ID           int     `json:"id"`
	Sum          float64 `json:"sum"`
	WalletSource int     `json:"walletSource"`
	WalletTarget int     `json:"walletTarget"`
}
type PG struct {
	log *logrus.Entry
	db  *sqlx.DB
	dsn string
}

func NewRepo(ctx context.Context, log *logrus.Logger, dsn string) (*PG, error) {
	db, err := sqlx.ConnectContext(ctx, "pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("err connecting to PG : %w", err)
	}
	err = db.PingContext(ctx)
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

func (pg *PG) Migrate(direction migrate.MigrationDirection) error {
	conn, err := sql.Open("pgx", pg.dsn)
	if err != nil {
		return err
	}
	defer func() {
		if err = conn.Close(); err != nil {
			pg.log.Error("err closing migration connection")
		}
	}()
	assetDir := func() func(string) ([]string, error) {
		return func(path string) ([]string, error) {
			dirEntry, er := migrations.ReadDir(path)
			if er != nil {
				return nil, er
			}
			entries := make([]string, 0)
			for _, e := range dirEntry {
				entries = append(entries, e.Name())
			}

			return entries, nil
		}
	}()
	asset := migrate.AssetMigrationSource{
		Asset:    migrations.ReadFile,
		AssetDir: assetDir,
		Dir:      "migrations",
	}
	_, err = migrate.Exec(conn, "postgres", asset, direction)
	return err
}

func (pg *PG) Close() {
	if err := pg.db.Close(); err != nil {
		pg.log.Errorf("err closing pg connection: %v", err)
	}
}

func (pg *PG) CreateWallet(ctx context.Context, wallet Wallet) (int, error) {
	query := `INSERT INTO wallet (owner, balance, updated_at) VALUES ($1,$2,$3) RETURNING id`
	var id int
	row := pg.db.QueryRowContext(ctx, query, wallet.Owner, wallet.Balance, time.Now())
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("err creating wallet: %w", err)
	}
	return id, nil

}

func (pg *PG) GetWallet(ctx context.Context, id int) (Wallet, error) {
	query := `SELECT owner, balance, created_at, updated_at FROM Wallet WHERE id = $1`
	var wallet Wallet
	if err := pg.db.GetContext(ctx, &wallet, query, id); err != nil {
		return Wallet{}, fmt.Errorf("err getting wallet : %w", err)
	}
	return wallet, nil
}

func (pg *PG) UpdateWallet(ctx context.Context, id int, wallet Wallet) (Wallet, error) {
	query := `UPDATE wallet SET owner = $1, balance = $2,updated_at = $3 WHERE id = $4 RETURNING owner, balance, created_at, updated_at`
	row := pg.db.QueryRowxContext(ctx, query, wallet.Owner, wallet.Balance, time.Now(), id)
	err := row.StructScan(&wallet)
	if err != nil {
		return Wallet{}, fmt.Errorf("err updating the Wallet: %w", err)
	}
	return wallet, nil
}

func (pg *PG) DeleteWallet(ctx context.Context, id int) error {
	query := `DELETE FROM wallet WHERE id = $1`
	_, err := pg.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("err deleting wallet : %w", err)
	}
	return nil
}

func (pg *PG) Deposit(ctx context.Context, request *Finrequest) error {
	query := `UPDATE wallet SET balance = balance + $1 WHERE id = $2`
	_, err := pg.db.ExecContext(ctx, query, request.Sum, request.ID)
	if err != nil {
		return fmt.Errorf("err depositing the Wallet: %w", err)
	}
	return nil
}
func (pg *PG) Withdrawal(ctx context.Context, request *Finrequest) error {
	wallet, err := pg.GetWallet(ctx, request.ID)
	if wallet.Balance < request.Sum {
		return fmt.Errorf("err not enough money: %w", err)
	}
	query := `UPDATE wallet SET balance = balance - $1 WHERE id = $2`
	_, err = pg.db.ExecContext(ctx, query, request.Sum, request.ID)

	if err != nil {
		return fmt.Errorf("err withdrawaling the Wallet: %w", err)
	}
	return nil
}

func (pg *PG) Transfer(ctx context.Context, request *Finrequest) error {
	tx, err := pg.db.BeginTx(ctx, nil)
	defer tx.Rollback()
	if err != nil {
		return fmt.Errorf("err trasnfering the wallet: %w", err)
	}
	query := `UPDATE wallet SET balance = balance - $1 WHERE id = $2 RETURNING balance`
	if _, err = tx.ExecContext(ctx, query, request.Sum, request.WalletSource); err != nil {
		return fmt.Errorf("err withdrawaling the Wallet: %w", err)
	}
	query = `UPDATE wallet SET balance = balance + $1 WHERE id = $2 RETURNING balance`
	if _, err = tx.ExecContext(ctx, query, request.Sum, request.WalletTarget); err != nil {
		return fmt.Errorf("err depositing the Wallet: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("err trasnfering the wallet: %w", err)
	}
	return nil
}
