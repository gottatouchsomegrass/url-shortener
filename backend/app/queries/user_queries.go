package queries

import (
	"context"

	"github.com/gottatouchsomegrass/url/app/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserQuery struct {
	DB *pgxpool.Pool
}

// CreateUser inserts a new user into the database.
func (q *UserQuery) CreateUser(
	ctx context.Context,
	user *models.User,
) error {
	query := `
		INSERT INTO users (
			email,
			password_hash
		)
		VALUES ($1, $2)
		RETURNING id, created_at
	`

	return q.DB.QueryRow(
		ctx,
		query,
		user.Email,
		user.PasswordHash,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)
}

// GetUserByEmail retrieves a user by their email address.
func (q *UserQuery) GetUserByEmail(
	ctx context.Context,
	email string,
) (*models.User, error) {
	query := `
		SELECT
			id,
			email,
			password_hash,
			created_at
		FROM users
		WHERE email = $1
	`

	var user models.User

	err := q.DB.QueryRow(
		ctx,
		query,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByID retrieves a user by their ID.
func (q *UserQuery) GetUserByID(
	ctx context.Context,
	userID int64,
) (*models.User, error) {
	query := `
		SELECT
			id,
			email,
			created_at
		FROM users
		WHERE id = $1
	`

	var user models.User

	err := q.DB.QueryRow(
		ctx,
		query,
		userID,
	).Scan(
		&user.ID,
		&user.Email,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// InvalidateToken invalidates a user's token.
// (implement later for refersh tokens)
// func (q *UserQuery) InvalidateToken(
// 	ctx context.Context,
// 	userID int64,
// ) error {
// 	query := `
// 		DELETE FROM tokens
// 		WHERE user_id = $1
// 	`

// 	_, err := q.DB.Exec(
// 		ctx,
// 		query,
// 		userID,
// 	)

// 	return err
// }
