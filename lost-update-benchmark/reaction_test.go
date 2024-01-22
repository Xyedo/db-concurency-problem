package lostupdatebenchmark_test

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xyedo/db-concurency-problem/config"
	"github.com/xyedo/db-concurency-problem/db"
	"github.com/xyedo/db-concurency-problem/helper"
	lostupdatebenchmark "github.com/xyedo/db-concurency-problem/lost-update-benchmark"
	"github.com/xyedo/db-concurency-problem/repository"
)

func init() {
	config.Get("../.env")
}

type Reaction interface {
	Do(ctx context.Context, threadId, userId string) error
}

func TestReactionCounter(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name     string
		reaction Reaction
	}{
		{
			name:     "locking",
			reaction: lostupdatebenchmark.ForUpdate{},
		},
		{
			name:     "repeatable read",
			reaction: lostupdatebenchmark.RepeatableRead{},
		},
		{
			name:     "compare and set",
			reaction: lostupdatebenchmark.CompareAndSet{},
		},
	}
	for _, tt := range tests {
		userId, err := helper.CreateUser()
		require.NoError(t, err)
		threadId, err := helper.CreateThread(userId)
		require.NoError(t, err)
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			err := addConcurentReaction(ctx, tt.reaction, 100, threadId)
			if err != nil {
				log.Println(err)
			}

			c, err := db.GetConnection(ctx)
			require.NoError(t, err)
			defer c.Release()

			thread, err := repository.GetThread(ctx, c, threadId)
			require.NoError(t, err)

			assert.Equal(t, 100, thread.TotalReaction)
			fmt.Println("execution time: ", time.Since(start))
		})
	}
}

func addConcurentReaction(ctx context.Context, cb Reaction, concurentUser int, threadId string) error {

	newUserIds := make([]string, 0, concurentUser)
	for i := 0; i < concurentUser; i++ {
		userId, err := helper.CreateUser()
		if err != nil {
			return err
		}
		newUserIds = append(newUserIds, userId)
	}

	var wg sync.WaitGroup
	errs := make([]error, concurentUser)
	for i := 0; i < concurentUser; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = cb.Do(ctx, threadId, newUserIds[i])
		}(i)
	}
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}
