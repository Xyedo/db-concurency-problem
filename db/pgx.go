package db

import (
	"context"
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

func Transaction(ctx context.Context, txOpt pgx.TxOptions, cb func(tx Connection) error) error {
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
