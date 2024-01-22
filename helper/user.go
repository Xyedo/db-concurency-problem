package helper

import (
	"context"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/go-faker/faker/v4"
	"github.com/xyedo/db-concurency-problem/db"
	"github.com/xyedo/db-concurency-problem/repository"
	"golang.org/x/crypto/bcrypt"
)

func CreateUser() (string, error) {
	conn, err := db.GetConnection(context.Background())
	if err != nil {
		return "", err
	}
	defer conn.Release()

	hashed, err := bcrypt.GenerateFromPassword([]byte(faker.Password()), bcrypt.MinCost)
	if err != nil {
		return "", err
	}

	userId := AccountId()
	err = repository.CreateAccount(context.Background(), conn, repository.Account{
		Id:             userId,
		Username:       ToPointer(faker.Username()),
		HashedPassword: string(hashed),
		CreatedOn:      time.Now(),
		Version:        1,
	})
	if err != nil {
		return "", err
	}

	return userId, nil
}

func CreateThread(userId string) (string, error) {
	conn, err := db.GetConnection(context.Background())
	if err != nil {
		return "", err
	}
	defer conn.Release()

	threadId := ThreadId()

	err = repository.CreateThread(context.Background(), conn, repository.Thread{
		Id:        threadId,
		Title:     faker.Sentence(),
		Body:      faker.Paragraph(),
		CreatedBy: userId,
		CreatedOn: time.Now(),

		Version: 1,
	})
	if err != nil {
		return "", err
	}

	return threadId, nil
}

func DeleteFakeTable(ctx context.Context) error {
	conn, err := db.GetConnection(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, `DELETE FROM FAKE_TABLE`)
	return err
}
func SelectFakeTable(ctx context.Context) ([]string, error) {
	conn, err := db.GetConnection(context.Background())
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var numbers []string

	err = pgxscan.Select(ctx, conn, &numbers, `SELECT number FROM FAKE_TABLE ORDER BY number ASC`)
	if err != nil {
		return nil, err
	}

	return numbers, nil
}
