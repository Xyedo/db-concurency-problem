package isolationlevelbenchmark

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/xyedo/db-concurency-problem/db"
	"github.com/xyedo/db-concurency-problem/helper"
	"github.com/xyedo/db-concurency-problem/repository"
)

func ReadModifyWriteComment(ctx context.Context, txOpt pgx.TxOptions, userId, threadId string) error {
	return db.Atomic(ctx, txOpt, func(tx db.Connection) error {
		_, err := repository.GetAccount(ctx, tx, userId)
		if err != nil {
			return err
		}

		thread, err := repository.GetThread(ctx, tx, threadId)
		if err != nil {
			return err
		}

		commentId := helper.CommentId()
		err = repository.CreateComment(ctx, tx, repository.Comment{
			Id:        commentId,
			ThreadId:  threadId,
			UserId:    userId,
			Content:   faker.Sentence(),
			CreatedOn: time.Now(),
			Version:   1,
		})
		if err != nil {
			return err
		}

		thread.TotalComment++
		return repository.UpdateThread(ctx, tx, thread)
	})
}

func TestReadModifyWriteComment(t *testing.T) {
	userId, err := helper.CreateUser()
	require.NoError(t, err)

	threadId, err := helper.CreateThread(userId)
	require.NoError(t, err)

	err = ReadModifyWriteComment(context.Background(), pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, userId, threadId)
	require.NoError(t, err)
}

func BenchmarkMultipleObjectObjectReadCommitted(b *testing.B) {
	userId, err := helper.CreateUser()
	if err != nil {
		panic(err)
	}

	threadId, err := helper.CreateThread(userId)
	if err != nil {
		panic(err)
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = ReadModifyWriteComment(context.Background(), pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, userId, threadId)
	}
}

func BenchmarkMultipleObjectObjectRepeatableRead(b *testing.B) {
	userId, err := helper.CreateUser()
	if err != nil {
		panic(err)
	}

	threadId, err := helper.CreateThread(userId)
	if err != nil {
		panic(err)
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = ReadModifyWriteComment(context.Background(), pgx.TxOptions{IsoLevel: pgx.RepeatableRead}, userId, threadId)
	}
}

func BenchmarkMultipleObjectObjectSerializable(b *testing.B) {
	userId, err := helper.CreateUser()
	if err != nil {
		panic(err)
	}

	threadId, err := helper.CreateThread(userId)
	if err != nil {
		panic(err)
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = ReadModifyWriteComment(context.Background(), pgx.TxOptions{IsoLevel: pgx.Serializable}, userId, threadId)
	}
}
