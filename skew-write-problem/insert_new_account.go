package skewwriteproblem

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/xyedo/db-concurency-problem/db"
	"github.com/xyedo/db-concurency-problem/helper"
	"github.com/xyedo/db-concurency-problem/repository"
	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	Username    *string
	Email       *string
	PhoneNumber *string
	Password    string
}

func InsertNewAccount(ctx context.Context, txOpt pgx.TxOptions, payload Account) error {
	if payload.Email == nil && payload.Username == nil && payload.PhoneNumber == nil {
		return errors.New("unique identifier is empty")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.MinCost)
	if err != nil {
		return err
	}
	return db.AtomicWithAutoRetry(ctx, txOpt, func(tx db.Connection) error {
		if payload.Username != nil {
			err := repository.CheckAccountUsernameAvailability(ctx, tx, *payload.Username)
			if err != nil {
				return err
			}
		}

		if payload.Email != nil {
			err := repository.CheckAccountEmailAvailability(ctx, tx, *payload.Email)
			if err != nil {
				return err
			}
		}

		if payload.PhoneNumber != nil {
			err := repository.CheckAccountPhoneNumberAvailability(ctx, tx, *payload.PhoneNumber)
			if err != nil {
				return err
			}
		}

		return repository.CreateAccount(ctx, tx, repository.Account{
			Id:             helper.AccountId(),
			Username:       payload.Username,
			Email:          payload.Email,
			PhoneNumber:    payload.PhoneNumber,
			HashedPassword: string(hashedPassword),
			CreatedOn:      time.Now(),
			Version:        1,
		})

	})
}
