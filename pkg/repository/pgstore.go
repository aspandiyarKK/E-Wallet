package repository

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"strconv"
	"time"

	"EWallet/pkg/models"

	"EWallet/pkg/metrics"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"

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
type FinRequest struct {
	Sum          float64 `json:"sum"`
	WalletTarget int     `json:"walletTarget"`
	UUID         string  `json:"uuid"`
}
type Transaction struct {
	Id        int       `json:"transaction_id" db:"id"`
	UUID      string    `json:"uuid" db:"uuid"`
	FromId    int       `json:"from_id" db:"from_id"`
	ToId      *int      `json:"to_id" db:"to_id"`
	Sum       float64   `json:"sum" db:"sum"`
	Operation string    `json:"operation" db:"operation"`
	Date      time.Time `json:"date" db:"date"`
}
type PG struct {
	log *logrus.Entry
	db  *sqlx.DB
	dsn string
}

var (
	ErrInsufficientFunds    = fmt.Errorf("err insuficient funds")
	ErrWalletNotFound       = fmt.Errorf("err wallet not found")
	ErrWalletTargetNotFound = fmt.Errorf("err wallet target  not found")
	ErrDuplicateKey         = fmt.Errorf("err duplicate key")
	ErrTransactionNotFound  = fmt.Errorf("err transaction not found")
)

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
	started := time.Now()
	defer func() {
		metrics.MetricDBRequestsDuration.WithLabelValues("CreateWallet").Observe(time.Since(started).Seconds())
	}()
	query := `INSERT INTO wallet (owner, balance, updated_at) VALUES ($1,$2,$3) RETURNING id`
	var id int
	row := pg.db.QueryRowContext(ctx, query, wallet.Owner, wallet.Balance, time.Now())
	if err := row.Scan(&id); err != nil {
		metrics.MetricErrCount.WithLabelValues("CreateWallet").Inc()
		return 0, fmt.Errorf("err creating wallet: %w", err)
	}
	return id, nil
}

func (pg *PG) GetWallet(ctx context.Context, id int) (Wallet, error) {
	started := time.Now()
	defer func() {
		metrics.MetricDBRequestsDuration.WithLabelValues("GetWallet").Observe(time.Since(started).Seconds())
	}()
	query := `SELECT owner, balance, created_at, updated_at FROM Wallet WHERE id = $1`
	var wallet Wallet
	if err := pg.db.GetContext(ctx, &wallet, query, id); err != nil {
		metrics.MetricErrCount.WithLabelValues("GetWallet").Inc()
		if errors.Is(err, sql.ErrNoRows) {
			return Wallet{}, ErrWalletNotFound
		}
		return Wallet{}, fmt.Errorf("err getting wallet : %w", err)
	}
	return wallet, nil
}

func (pg *PG) UpdateWallet(ctx context.Context, id int, wallet Wallet) (Wallet, error) {
	started := time.Now()
	defer func() {
		metrics.MetricDBRequestsDuration.WithLabelValues("UpdateWallet").Observe(time.Since(started).Seconds())
	}()
	query := `UPDATE wallet SET owner = $1, balance = $2,updated_at = $3 WHERE id = $4 RETURNING owner, balance, created_at, updated_at`
	row := pg.db.QueryRowxContext(ctx, query, wallet.Owner, wallet.Balance, time.Now(), id)
	err := row.StructScan(&wallet)
	if err != nil {
		metrics.MetricErrCount.WithLabelValues("UpdateWallet").Inc()
		if errors.Is(err, sql.ErrNoRows) {
			return Wallet{}, ErrWalletNotFound
		}
		return Wallet{}, fmt.Errorf("err updating the Wallet: %w", err)
	}
	return wallet, nil
}

func (pg *PG) DeleteWallet(ctx context.Context, id int) error {
	started := time.Now()
	defer func() {
		metrics.MetricDBRequestsDuration.WithLabelValues("DeleteWallet").Observe(time.Since(started).Seconds())
	}()
	query := `DELETE FROM wallet WHERE id = $1`
	res, err := pg.db.ExecContext(ctx, query, id)
	cnt, _ := res.RowsAffected()
	if cnt == 0 {
		return ErrWalletNotFound
	}
	if err != nil {
		metrics.MetricErrCount.WithLabelValues("DeleteWallet").Inc()
		return fmt.Errorf("err deleting wallet : %w", err)
	}
	return nil
}

func (pg *PG) Deposit(ctx context.Context, id int, request *FinRequest) error {
	started := time.Now()
	defer func() {
		metrics.MetricDBRequestsDuration.WithLabelValues("Deposit").Observe(time.Since(started).Seconds())
	}()
	tx, err := pg.db.BeginTxx(ctx, nil)
	if err != nil {
		metrics.MetricErrCount.WithLabelValues("Deposit").Inc()
		return fmt.Errorf("err starting transaction: %w", err)
	}
	defer func() {
		if err = tx.Rollback(); err != nil {
			pg.log.Error("err rolling back transaction")
		}
	}()
	query := `INSERT INTO transaction (uuid,from_id,operation,sum) VALUES ($1,$2,$3,$4)`
	_, err = tx.ExecContext(ctx, query, request.UUID, id, "deposit", request.Sum)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgerrcode.UniqueViolation {
			return ErrDuplicateKey
		}
	}

	query = `UPDATE wallet SET balance = balance + $1 WHERE id = $2`
	res, err := tx.ExecContext(ctx, query, request.Sum, id)
	cnt, _ := res.RowsAffected()
	if cnt == 0 {
		return ErrWalletNotFound
	}
	if err != nil {
		metrics.MetricErrCount.WithLabelValues("Deposit").Inc()
		return fmt.Errorf("err depositing the Wallet: %w", err)
	}
	if err = tx.Commit(); err != nil {
		metrics.MetricErrCount.WithLabelValues("Deposit").Inc()
		return fmt.Errorf("err committing transaction: %w", err)
	}
	return nil
}

func (pg *PG) Withdrawal(ctx context.Context, id int, request *FinRequest) error {
	started := time.Now()
	defer func() {
		metrics.MetricDBRequestsDuration.WithLabelValues("Withdrawal").Observe(time.Since(started).Seconds())
	}()
	tx, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		metrics.MetricErrCount.WithLabelValues("Withdrawal").Inc()
		return fmt.Errorf("err trasnfering the wallet: %w", err)
	}
	defer func() {
		if err = tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			metrics.MetricErrCount.WithLabelValues("Withdrawal").Inc()
			pg.log.Error("err rolling back transfer transaction")
		}
	}()
	query := `INSERT INTO transaction (uuid,from_id,operation,sum) VALUES ($1,$2,$3,$4)`
	_, err = tx.ExecContext(ctx, query, request.UUID, id, "withdraw", request.Sum)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgerrcode.UniqueViolation {
			return ErrDuplicateKey
		}
	}
	if err = pg.checkBalance(ctx, tx, id, request.Sum); err != nil {
		return err
	}
	query = `UPDATE wallet SET balance = balance - $1 WHERE id = $2 `
	res, err := tx.ExecContext(ctx, query, request.Sum, id)
	cnt, _ := res.RowsAffected()
	if cnt == 0 {
		return ErrWalletNotFound
	}
	if err != nil {
		metrics.MetricErrCount.WithLabelValues("Withdrawal").Inc()
		return fmt.Errorf("err withdrawing the Wallet: %w", err)
	}
	if err = tx.Commit(); err != nil {
		metrics.MetricErrCount.WithLabelValues("Withdrawal").Inc()
		return fmt.Errorf("err committing the transaction: %w", err)
	}
	return nil
}

func (pg *PG) Transfer(ctx context.Context, id int, request *FinRequest) error {
	started := time.Now()
	defer func() {
		metrics.MetricDBRequestsDuration.WithLabelValues("Transfer").Observe(time.Since(started).Seconds())
	}()
	tx, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		metrics.MetricErrCount.WithLabelValues("Transfer").Inc()
		return fmt.Errorf("err trasnfering the wallet: %w", err)
	}
	defer func() {
		if err = tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			metrics.MetricErrCount.WithLabelValues("Transfer").Inc()
			pg.log.Error("err rolling back transfer transaction")
		}
	}()

	query := `INSERT INTO transaction (uuid,from_id,to_id,operation,sum) VALUES ($1,$2,$3,$4,$5)`
	_, err = tx.ExecContext(ctx, query, request.UUID, id, request.WalletTarget, "transfer", request.Sum)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgerrcode.UniqueViolation {
			return ErrDuplicateKey
		}
	}
	if err = pg.checkBalance(ctx, tx, id, request.Sum); err != nil {
		return err
	}
	query = `UPDATE wallet SET balance = balance - $1 WHERE id = $2 RETURNING balance`
	if res, err := tx.ExecContext(ctx, query, request.Sum, id); err != nil {
		metrics.MetricErrCount.WithLabelValues("Transfer").Inc()
		cnt, _ := res.RowsAffected()
		if cnt == 0 {
			return ErrWalletNotFound
		}
		return fmt.Errorf("err withdrawaling the Wallet: %w", err)
	}
	query = `UPDATE wallet SET balance = balance + $1 WHERE id = $2 RETURNING balance`
	if res, err := tx.ExecContext(ctx, query, request.Sum, request.WalletTarget); err != nil {
		cnt, _ := res.RowsAffected()
		if cnt == 0 {
			return ErrWalletTargetNotFound
		}
		metrics.MetricErrCount.WithLabelValues("Transfer").Inc()
		return fmt.Errorf("err depositing the Wallet: %w", err)
	}
	if err = tx.Commit(); err != nil {
		metrics.MetricErrCount.WithLabelValues("Transfer").Inc()
		return fmt.Errorf("err trasnfering the wallet: %w", err)
	}
	return nil
}

type querier interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func (pg *PG) checkBalance(ctx context.Context, querier querier, id int, sum float64) error {
	started := time.Now()
	defer func() {
		metrics.MetricDBRequestsDuration.WithLabelValues("checkBalance").Observe(time.Since(started).Seconds())
	}()
	var balance float64
	query := `SELECT balance FROM wallet WHERE id = $1 FOR UPDATE`
	row := querier.QueryRowContext(ctx, query, id)
	if err := row.Scan(&balance); err != nil {
		metrics.MetricErrCount.WithLabelValues("checkBalance").Inc()
		if errors.Is(err, sql.ErrNoRows) {
			return ErrWalletNotFound
		}
		return fmt.Errorf("err checking balance: %w", err)
	}
	if balance < sum {
		return ErrInsufficientFunds
	}
	return nil
}

func (pg *PG) CheckBalance(ctx context.Context, id int) (float64, error) {
	started := time.Now()
	defer func() {
		metrics.MetricDBRequestsDuration.WithLabelValues("CheckBalance").Observe(time.Since(started).Seconds())
	}()
	var balance float64
	query := `SELECT balance FROM wallet WHERE id = $1`
	row := pg.db.QueryRowContext(ctx, query, id)
	if err := row.Scan(&balance); err != nil {
		metrics.MetricErrCount.WithLabelValues("CheckBalance").Inc()
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrWalletNotFound
		}
		return 0, fmt.Errorf("err checking balance: %w", err)
	}
	return balance, nil
}

func (pg *PG) GetTransactions(ctx context.Context, id int, params *models.TransactionQueryParams) ([]Transaction, error) {
	var ans []Transaction
	query := `
SELECT id,
       uuid,
       from_id,
       to_id,
       operation,
       sum,
       date 
FROM transaction
WHERE from_id=$1 OR to_id=$1 `
	if params != nil {
		switch params.Sort {
		case "sum":
			query += " ORDER BY SUM"
		default:
			query += " ORDER BY DATE"
		}
		switch params.Desc {
		case false:
			query += " ASC"
		default:
			query += " DESC"
		}
		if params.Limit != 0 {
			query += " LIMIT " + strconv.Itoa(params.Limit)
		} else {
			query += " LIMIT 100"
		}
		if params.Offset != 0 {
			query += " OFFSET " + strconv.Itoa(params.Offset)
		}
	} else {
		query += "ORDER BY date DESC"
	}
	_, err := pg.GetWallet(ctx, id)
	if errors.Is(err, ErrWalletNotFound) {
		return nil, ErrWalletNotFound
	}
	if err := pg.db.SelectContext(ctx, &ans, query, id); err != nil {
		return nil, fmt.Errorf("err getting transaction : %w", err)
	}
	return ans, nil
}
