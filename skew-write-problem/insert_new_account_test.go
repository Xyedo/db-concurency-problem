package skewwriteproblem_test

import (
	"context"
	"sync"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/xyedo/db-concurency-problem/config"
	skewwriteproblem "github.com/xyedo/db-concurency-problem/skew-write-problem"
)

func init() {
	config.Get("../.env")
}

func TestInsertNewAccount(t *testing.T) {
	tests := []struct {
		name  string
		txOpt pgx.TxOptions
	}{
		{
			name:  "violating unique constraint when in read commited / default",
			txOpt: pgx.TxOptions{},
		},
		{
			name:  "violating unique constraint when in repeatable read",
			txOpt: pgx.TxOptions{IsoLevel: pgx.RepeatableRead},
		},
		{
			name:  "having an normal error when in serializable",
			txOpt: pgx.TxOptions{IsoLevel: pgx.Serializable},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userName := faker.Username()
			var wg sync.WaitGroup
			errs := make([]error, 2)

			wg.Add(1)
			go func() {
				defer wg.Done()
				errs[0] = skewwriteproblem.InsertNewAccount(context.Background(), tt.txOpt, skewwriteproblem.Account{
					Username: &userName,
					Password: faker.Password(),
				})
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()
				errs[1] = skewwriteproblem.InsertNewAccount(context.Background(), tt.txOpt, skewwriteproblem.Account{
					Username: &userName,
					Password: faker.Password(),
				})
			}()

			wg.Wait()

			for _, err := range errs {
				require.NoError(t, err)
			}
		})
	}

}
