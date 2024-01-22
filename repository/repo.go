package repository

import (
	"context"
	"errors"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/xyedo/db-concurency-problem/db"
)

type Account struct {
	Id             string     `db:"id"`
	Username       *string    `db:"username"`
	Email          *string    `db:"email"`
	PhoneNumber    *string    `db:"phone_number"`
	HashedPassword string     `db:"hashed_password"`
	IsDeleted      bool       `db:"is_deleted"`
	CreatedOn      time.Time  `db:"created_on"`
	UpdatedOn      *time.Time `db:"updated_on"`
	Version        int        `db:"version"`
}

func CreateAccount(ctx context.Context, conn db.Connection, payload Account) error {
	tag, err := conn.Exec(ctx, `INSERT INTO ACCOUNT (
		id, 
		username, 
		phone_number, 
		email, 
		hashed_password, 
		is_deleted, 
		created_on, 
		updated_on,
		version
		) VALUES 
		($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		payload.Id,
		payload.Username,
		payload.Email,
		payload.PhoneNumber,
		payload.HashedPassword,
		payload.IsDeleted,
		payload.CreatedOn,
		payload.UpdatedOn,
		payload.Version,
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() != 1 {
		return errors.New("nothing was inserted, something went wrong")
	}

	return nil
}

func CheckAccountUsernameAvailability(ctx context.Context, conn db.Connection, username string) error {
	usernameCount := 0
	err := conn.QueryRow(ctx,
		`
		SELECT 
			count(1)
		FROM ACCOUNT 
		where username = $1`,
		username,
	).Scan(&usernameCount)
	if err != nil {
		return err
	}

	if usernameCount > 0 {
		return errors.New("username already taken")
	}

	return nil
}
func CheckAccountEmailAvailability(ctx context.Context, conn db.Connection, email string) error {
	emailCount := 0
	err := conn.QueryRow(ctx,
		`
		SELECT 
			count(1)
		FROM ACCOUNT 
		where email = $1`,
		email,
	).Scan(&emailCount)
	if err != nil {
		return err
	}

	if emailCount > 0 {
		return errors.New("username already taken")
	}

	return nil
}
func CheckAccountPhoneNumberAvailability(ctx context.Context, conn db.Connection, phoneNumber string) error {
	phoneNumberCount := 0
	err := conn.QueryRow(ctx,
		`
		SELECT 
			count(1)
		FROM ACCOUNT 
		where phone_number = $1`,
		phoneNumber,
	).Scan(&phoneNumberCount)
	if err != nil {
		return err
	}

	if phoneNumberCount > 0 {
		return errors.New("username already taken")
	}

	return nil
}
func GetAccount(ctx context.Context, conn db.Connection, id string) (Account, error) {
	var account Account
	err := pgxscan.Get(ctx, conn, &account, `SELECT
		id, 
		username, 
		phone_number, 
		email, 
		hashed_password, 
		is_deleted, 
		created_on, 
		updated_on,
		version
		FROM ACCOUNT
		WHERE id = $1`, id,
	)
	if err != nil {
		return Account{}, err
	}

	return account, nil
}

func UpdateAccount(ctx context.Context, conn db.Connection, payload Account) error {
	tag, err := conn.Exec(ctx, `
	UPDATE ACCOUNT SET
		username = $2, 
		phone_number = $3, 
		email = $4, 
		hashed_password = $5, 
		is_deleted = $6, 
		updated_on = $7,
		version = version + 1
	WHERE id = $1
	
	`,
		payload.Id,
		payload.Username,
		payload.Email,
		payload.PhoneNumber,
		payload.HashedPassword,
		payload.IsDeleted,
		payload.UpdatedOn,
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() != 1 {
		return errors.New("nothing was inserted, something went wrong")
	}

	return nil
}

type Thread struct {
	Id            string     `db:"id"`
	Title         string     `db:"title"`
	Body          string     `db:"body"`
	TotalComment  int        `db:"total_comment"`
	TotalReaction int        `db:"total_reaction"`
	CreatedBy     string     `db:"created_by"`
	CreatedOn     time.Time  `db:"created_on"`
	UpdatedBy     *string    `db:"updated_by"`
	UpdatedOn     *time.Time `db:"updated_on"`
	IsDeleted     bool       `db:"is_deleted"`
	Version       int        `db:"version"`
}

func CreateThread(ctx context.Context, conn db.Connection, payload Thread) error {
	tag, err := conn.Exec(ctx, `INSERT INTO THREAD (
		id, 
		title,
		body,
		created_by,
		created_on,
		updated_by,
		updated_on,
		is_deleted,
		version
		) VALUES 
		($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		payload.Id,
		payload.Title,
		payload.Body,
		payload.CreatedBy,
		payload.CreatedOn,
		payload.UpdatedBy,
		payload.UpdatedOn,
		payload.IsDeleted,
		payload.Version,
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() != 1 {
		return errors.New("nothing was inserted, something went wrong")
	}

	return nil
}

type GetThreadOption struct {
	ForUpdate bool
}

func GetThread(ctx context.Context, conn db.Connection, id string, opts ...GetThreadOption) (Thread, error) {
	const getThread = `
	SELECT
		id, 
		title,
		body,
		total_comment,
		total_reaction,
		created_by,
		created_on,
		updated_by,
		updated_on,
		is_deleted,
		version
	FROM THREAD
	WHERE id = $1`

	query := getThread
	if len(opts) > 0 && opts[0].ForUpdate {
		query += "\n FOR UPDATE"
	}
	var thread Thread
	err := pgxscan.Get(ctx, conn,
		&thread,
		query,
		id,
	)
	if err != nil {
		return Thread{}, err
	}

	return thread, nil
}

type CompareAndSetOption struct {
	Version int
}
type UpdateThreadOption struct {
	CompareAndSet *CompareAndSetOption
}

func UpdateThread(ctx context.Context, conn db.Connection, payload Thread, opts ...UpdateThreadOption) error {
	const updateThread = `
	UPDATE THREAD SET
		title = $2,
		body = $3,
		total_comment = $4,
		total_reaction = $5,
		updated_by = $6,
		updated_on = $7,
		is_deleted = $8,
		version = version +1
	WHERE id = $1`

	query := updateThread
	args := []any{
		payload.Id,
		payload.Title,
		payload.Body,
		payload.TotalComment,
		payload.TotalReaction,
		payload.UpdatedBy,
		payload.UpdatedOn,
		payload.IsDeleted,
	}

	if len(opts) > 0 && opts[0].CompareAndSet != nil && opts[0].CompareAndSet.Version != 0 {
		query += " AND\nversion = $9"
		args = append(args, opts[0].CompareAndSet.Version)
	}
	tag, err := conn.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	if tag.RowsAffected() != 1 {
		if len(opts) > 0 && opts[0].CompareAndSet != nil {
			return db.ErrVersionMisMatch
		}

		return errors.New("nothing was updated, something went wrong")
	}

	return nil
}

type Comment struct {
	Id            string     `db:"id"`
	ThreadId      string     `db:"thread_id"`
	UserId        string     `db:"user_id"`
	ReplyTo       *string    `db:"reply_to"`
	TotalReply    int        `db:"total_reply"`
	TotalReaction int        `db:"total_reaction"`
	Content       string     `db:"content"`
	CreatedOn     time.Time  `db:"created_on"`
	UpdatedOn     *time.Time `db:"updated_on"`
	IsDeleted     bool       `db:"is_deleted"`
	Version       int        `db:"version"`
}

func CreateComment(ctx context.Context, conn db.Connection, payload Comment) error {
	tag, err := conn.Exec(ctx, `INSERT INTO COMMENT (
		id,
		thread_id,
		user_id,
		reply_to,
		content,
		created_on,
		updated_on,
		is_deleted,
		version
		) VALUES 
		($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		payload.Id,
		payload.ThreadId,
		payload.UserId,
		payload.ReplyTo,
		payload.Content,
		payload.CreatedOn,
		payload.UpdatedOn,
		payload.IsDeleted,
		payload.Version,
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() != 1 {
		return errors.New("nothing was inserted, something went wrong")
	}

	return nil
}

func GetComment(ctx context.Context, conn db.Connection, id string) (Comment, error) {
	var comment Comment
	err := pgxscan.Get(ctx, conn, &comment,
		`
		SELECT
			id,
			thread_id,
			user_id,
			reply_to,
			content,
			total_reply,
			total_reaction,
			created_on,
			updated_on,
			is_deleted,
			version
		FROM COMMENT
		WHERE id = $1`,
		id,
	)
	if err != nil {
		return Comment{}, err
	}

	return comment, nil
}

func UpdateComment(ctx context.Context, conn db.Connection, payload Comment) error {
	tag, err := conn.Exec(ctx, `
	UPDATE COMMENT SET
		content = $2,
		total_reply = $3,
		total_reaction = $4,
		updated_on = $5,
		is_deleted = $6,
		version = version  +1
	WHERE id = $1`,
		payload.Id,
		payload.Content,
		payload.UpdatedOn,
		payload.IsDeleted,
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() != 1 {
		return errors.New("nothing was inserted, something went wrong")
	}

	return nil
}

type Reaction struct {
	Id        string     `db:"id"`
	AccountId string     `db:"account_id"`
	ThreadId  *string    `db:"thread_id"`
	CommentId *string    `db:"comment_id"`
	Content   string     `db:"content"`
	CreatedOn time.Time  `db:"created_on"`
	UpdatedOn *time.Time `db:"updated_on"`
	Version   int        `db:"version"`
}

func CreateReaction(ctx context.Context, conn db.Connection, payload Reaction) error {
	tag, err := conn.Exec(ctx, `INSERT INTO REACTION (
		id,
		account_id,
		thread_id,
		comment_id,
		content,
		created_on,
		updated_on,
		version
		) VALUES 
		($1,$2,$3,$4,$5,$6,$7,$8)`,
		payload.Id,
		payload.AccountId,
		payload.ThreadId,
		payload.CommentId,
		payload.Content,
		payload.CreatedOn,
		payload.UpdatedOn,
		payload.Version,
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() != 1 {
		return errors.New("nothing was inserted, something went wrong")
	}

	return nil
}

func UpdateReaction(ctx context.Context, conn db.Connection, payload Reaction) error {
	tag, err := conn.Exec(ctx, `
	UPDATE REACTION SET
		content = $2,
		updated_on = $3,
		version = $4
	WHERE id = $1`,
		payload.Id,
		payload.Content,
		payload.UpdatedOn,
		payload.Version,
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() != 1 {
		return errors.New("nothing was inserted, something went wrong")
	}

	return nil
}

func DeleteReaction(ctx context.Context, conn db.Connection, id string) error {
	tag, err := conn.Exec(ctx, `DELETE FROM REACTION WHERE id = $1`, id)
	if err != nil {
		return err
	}

	if tag.RowsAffected() != 1 {
		return errors.New("nothing was inserted, something went wrong")
	}
	return nil
}
