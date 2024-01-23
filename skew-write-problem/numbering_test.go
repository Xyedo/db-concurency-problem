package skewwriteproblem_test

import (
	"context"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xyedo/db-concurency-problem/helper"
	skewwriteproblem "github.com/xyedo/db-concurency-problem/skew-write-problem"
)

func TestInsertNewFakeTable(t *testing.T) {
	tests := []struct {
		name  string
		txOpt pgx.TxOptions
	}{
		{
			name:  "no error but duplicate numbering on default / read commited",
			txOpt: pgx.TxOptions{},
		},
		{
			name:  "mo error but duplicate numbering in repeatable read",
			txOpt: pgx.TxOptions{IsoLevel: pgx.RepeatableRead},
		},
		{
			name:  "no error and sequential numbering when in serializable",
			txOpt: pgx.TxOptions{IsoLevel: pgx.Serializable},
		},
	}
	const maxIteration = 10
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wg sync.WaitGroup
			errs := make([]error, maxIteration)

			wg.Add(maxIteration)
			for i := 0; i < maxIteration; i++ {
				go func(i int) {
					defer wg.Done()
					errs[i] = skewwriteproblem.InsertNewFakeTable(context.Background(), tt.txOpt)
				}(i)
			}

			wg.Wait()

			for _, err := range errs {
				assert.NoError(t, err)
			}
			s, err := helper.SelectFakeTable(context.Background())
			assert.NoError(t, err)

			assert.Equal(t, []string{"ft-001", "ft-002", "ft-003", "ft-004", "ft-005", "ft-006", "ft-007", "ft-008", "ft-009"}, s)

			err = helper.DeleteFakeTable(context.Background())
			require.NoError(t, err)
		})
	}

}
