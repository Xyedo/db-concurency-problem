package lostupdatebenchmark

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/xyedo/db-concurency-problem/db"
	"github.com/xyedo/db-concurency-problem/helper"
	"github.com/xyedo/db-concurency-problem/repository"
)

type ForUpdate struct{}

func (f ForUpdate) Do(ctx context.Context, threadId, userId string) error {
	return f.readModifyWriteReactionToThreadId(ctx, pgx.TxOptions{}, threadId, userId)
}

func (ForUpdate) readModifyWriteReactionToThreadId(ctx context.Context, txOpt pgx.TxOptions, threadId, userId string) error {
	return db.Atomic(ctx, txOpt, func(tx db.Connection) error {
		thread, err := repository.GetThread(ctx, tx, threadId, repository.GetThreadOption{ForUpdate: true})
		if err != nil {
			return err
		}

		_, err = repository.GetAccount(ctx, tx, userId)
		if err != nil {
			return err
		}

		err = repository.CreateReaction(ctx, tx, repository.Reaction{
			Id:        helper.ReactionId(),
			AccountId: userId,
			ThreadId:  &threadId,
			Content:   "like",
			CreatedOn: time.Now(),
			Version:   1,
		})
		if err != nil {
			return err
		}

		thread.TotalReaction++
		thread.Version++
		return repository.UpdateThread(ctx, tx, thread)
	})
}

type RepeatableRead struct{}

func (r RepeatableRead) Do(ctx context.Context, threadId, userId string) error {
	return r.readModifyWriteReactionToThreadId(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead}, threadId, userId)
}

func (RepeatableRead) readModifyWriteReactionToThreadId(ctx context.Context, txOpt pgx.TxOptions, threadId, userId string) error {
	return db.AtomicWithAutoRetry(ctx, txOpt, func(tx db.Connection) error {
		thread, err := repository.GetThread(ctx, tx, threadId)
		if err != nil {
			return err
		}

		_, err = repository.GetAccount(ctx, tx, userId)
		if err != nil {
			return err
		}

		err = repository.CreateReaction(ctx, tx, repository.Reaction{
			Id:        helper.ReactionId(),
			AccountId: userId,
			ThreadId:  &threadId,
			Content:   "like",
			CreatedOn: time.Now(),
			Version:   1,
		})
		if err != nil {
			return err
		}

		thread.TotalReaction++
		thread.Version++
		return repository.UpdateThread(ctx, tx, thread)
	})
}

type CompareAndSet struct{}

func (c CompareAndSet) Do(ctx context.Context, threadId, userId string) error {
	return c.readModifyWriteReactionToThreadId(ctx, threadId, userId)
}

func (CompareAndSet) readModifyWriteReactionToThreadId(ctx context.Context, threadId, userId string) error {
	conn, err := db.GetConnection(ctx)
	if err != nil {
		return err
	}

	_, err = repository.GetAccount(ctx, conn, userId)
	if err != nil {
		return err
	}
	conn.Release()

	return db.RetryMatchAndSet(ctx, func(conn db.Connection) error {
		thread, err := repository.GetThread(ctx, conn, threadId)
		if err != nil {
			return err
		}

		reactionId := helper.ReactionId()
		err = repository.CreateReaction(ctx, conn, repository.Reaction{
			Id:        reactionId,
			AccountId: userId,
			ThreadId:  &threadId,
			Content:   "like",
			CreatedOn: time.Now(),
			Version:   1,
		})
		if err != nil {
			//in a production environtment we need to deleteReaction in case its already created but failed to send a success response back
			// _ = repository.DeleteReaction(ctx, conn, reactionId)
			return err
		}
		oldThread := thread

		thread.TotalReaction++
		thread.Version++

		err = repository.UpdateThread(ctx, conn, thread, repository.UpdateThreadOption{
			CompareAndSet: &repository.CompareAndSetOption{
				Version: oldThread.Version,
			},
		})
		if err != nil {
			//in a production environtment we need to rollback UpdateThread to its previous State in case its already updated but failed to send a success response back
			// _ = repository.UpdateThread(ctx, conn, oldThread)
			_ = repository.DeleteReaction(ctx, conn, reactionId)

			return err
		}

		return nil
	})

}
