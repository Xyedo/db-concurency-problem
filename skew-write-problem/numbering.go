package skewwriteproblem

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/xyedo/db-concurency-problem/db"
	"github.com/xyedo/db-concurency-problem/helper"
)

func InsertNewFakeTable(ctx context.Context, txOpt pgx.TxOptions) error {
	return db.AtomicWithAutoRetry(ctx, txOpt, func(tx db.Connection) error {
		var numbering string
		err := pgxscan.Get(ctx, tx, &numbering,
			`SELECT "number" FROM FAKE_TABLE ORDER BY created_on DESC LIMIT 1`,
		)
		if err != nil {
			if !pgxscan.NotFound(err) {
				return err
			}
			numbering = "ft-000"
		}
		el := strings.Split(numbering, "-")
		if len(el) < 2 {
			return errors.New("dirty data")
		}
		number, err := strconv.Atoi(el[1])
		if err != nil {
			return err
		}
		number++
		newNumber := fmt.Sprintf("ft-%0*d", 3, number)
		tag, err := tx.Exec(ctx,
			`
			INSERT INTO FAKE_TABLE 
			(
				id,
				"number",
				created_on,
				updated_on,
				version
			) VALUES ($1,$2,$3,$4, $5)`,
			helper.FakeTableId(), newNumber, time.Now(), nil, 1,
		)

		if err != nil {
			return err
		}

		if tag.RowsAffected() != 1 {
			return errors.New("sometin wen wong")
		}

		return nil

	})
}
