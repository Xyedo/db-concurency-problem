package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xyedo/db-concurency-problem/config"
)

var (
	pool *pgxpool.Pool
	once sync.Once
)

type Connection interface {
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults

	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func connect(ctx context.Context) {
	config := config.Get().Postgre
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable pool_max_conns=%d",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.Database,
		runtime.NumCPU()*4,
	)
	c, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalln(err)
	}

	db, err := pgxpool.NewWithConfig(ctx, c)
	if err != nil {
		log.Fatalln(err)
	}

	err = db.Ping(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	pool = db
}

func GetConnection(ctx context.Context) (*pgxpool.Conn, error) {
	once.Do(func() { connect(ctx) })
	return pool.Acquire(ctx)
}

func Atomic(ctx context.Context, txOpt pgx.TxOptions, cb func(tx Connection) error) error {
	conn, err := GetConnection(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, txOpt)
	if err != nil {
		return err
	}

	err = cb(tx)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("cannot rollback %w: %w", rbErr, err)
		}

		return err
	}

	return tx.Commit(ctx)
}

var ErrVersionMisMatch = errors.New("version mismacth, must retry")

const maxRetry = 5

func AtomicWithAutoRetry(ctx context.Context, txOpt pgx.TxOptions, cb func(tx Connection) error) error {
	return transaction(ctx, txOpt, cb, maxRetry)
}

func RetryMatchAndSet(ctx context.Context, cb func(conn Connection) error) error {
	conn, err := GetConnection(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	var errToTry error

	retryCount := 0
	for retryCount < maxRetry {
		errToTry = cb(conn)
		if errToTry != nil && !errors.Is(errToTry, ErrVersionMisMatch) {
			return errToTry
		}

		retryCount++
	}

	if retryCount == maxRetry {
		return ErrLimitRetry
	}

	return nil
}

var ErrLimitRetry = errors.New("retry limit exceeded!")

func transaction(ctx context.Context, txOpt pgx.TxOptions, cb func(tx Connection) error, retry int) error {
	if retry < 0 {
		return ErrLimitRetry
	}

	conn, err := GetConnection(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, txOpt)
	if err != nil {
		return err
	}

	err = cb(tx)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("cannot rollback %w: %w", rbErr, err)
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "40001" {
			return transaction(ctx, txOpt, cb, retry-1)

		}
		return err

	}

	err = tx.Commit(ctx)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "40001" {
			return transaction(ctx, txOpt, cb, retry-1)
		}

		return err
	}

	return nil
}
