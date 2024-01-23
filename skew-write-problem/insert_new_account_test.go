package skewwriteproblem_test

import (
	"context"
	"sync"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/go-faker/faker/v4/pkg/options"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"github.com/xyedo/db-concurency-problem/config"
	skewwriteproblem "github.com/xyedo/db-concurency-problem/skew-write-problem"
)

func init() {
	config.Get("../.env")
}

func TestInsertNewAccount(t *testing.T) {
	tests := []struct {
		name    string
		txOpt   pgx.TxOptions
		wantErr bool
	}{
		{
			name:    "violating unique constraint when in read commited / default",
			txOpt:   pgx.TxOptions{},
			wantErr: true,
		},
		{
			name:    "violating unique constraint when in repeatable read",
			txOpt:   pgx.TxOptions{IsoLevel: pgx.RepeatableRead},
			wantErr: true,
		},
		{
			name:    "no error when in serializable",
			txOpt:   pgx.TxOptions{IsoLevel: pgx.Serializable},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userName := faker.Username(options.WithGenerateUniqueValues(true))
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

			var vErr error
			for _, err := range errs {
				if err != nil {
					vErr = err
				}
			}
			if tt.wantErr {
				require.Error(t, vErr)
				var pgErr *pgconn.PgError
				require.ErrorAs(t, vErr, &pgErr)
				require.Equal(t, "23505", pgErr.Code)
			} else {
				require.Equal(t, "username already taken", vErr.Error())
			}

		})
	}

}
