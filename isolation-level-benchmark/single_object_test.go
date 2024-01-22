package isolationlevelbenchmark

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/xyedo/db-concurency-problem/config"
	"github.com/xyedo/db-concurency-problem/db"
	"github.com/xyedo/db-concurency-problem/helper"
	"github.com/xyedo/db-concurency-problem/repository"
)

func init() {
	config.Get("../.env")
}

func ReadModifyWriteUser(ctx context.Context, txOpt pgx.TxOptions, userId string) error {
	return db.Atomic(ctx, txOpt, func(tx db.Connection) error {
		account, err := repository.GetAccount(ctx, tx, userId)
		if err != nil {
			return err
		}

		account.Username = helper.ToPointer(faker.Username())
		account.Email = helper.ToPointer(faker.Email())
		return repository.UpdateAccount(ctx, tx, account)
	})
}

func TestReadModifyWriteUser(t *testing.T) {
	userId, err := helper.CreateUser()
	require.NoError(t, err)

	err = ReadModifyWriteUser(context.Background(), pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, userId)
	require.NoError(t, err)
}

func BenchmarkSingleObjectReadCommitted(b *testing.B) {
	userId, err := helper.CreateUser()
	if err != nil {
		panic(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = ReadModifyWriteUser(context.Background(), pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, userId)
	}
}

func BenchmarkSingleObjectRepeatableRead(b *testing.B) {
	userId, err := helper.CreateUser()
	if err != nil {
		panic(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = ReadModifyWriteUser(context.Background(), pgx.TxOptions{IsoLevel: pgx.RepeatableRead}, userId)
	}
}

func BenchmarkSingleObjectSerializable(b *testing.B) {
	userId, err := helper.CreateUser()
	if err != nil {
		panic(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = ReadModifyWriteUser(context.Background(), pgx.TxOptions{IsoLevel: pgx.Serializable}, userId)
	}
}
