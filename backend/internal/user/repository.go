package user

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

// INSERT into table users new User
func (r *Repository) CreateUser(
	ctx context.Context,
	telegramID int64,
	groupID *string,
) (*User, error) {
	query := `
        INSERT INTO users (telegram_id, group_id)
        VALUES ($1, $2)
        RETURNING id, telegram_id, group_id, created_at
    `
	row := r.pool.QueryRow(
		ctx,
		query,
		telegramID,
		groupID,
	)

	var user User
	err := row.Scan(
		&user.ID,
		&user.TelegramID,
		&user.GroupID,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repository) GetUserByTelegramID(
	ctx context.Context,
	telegramID int64,
) (*User, error) {
	query := `
        SELECT id, telegram_id, group_id, created_at
		FROM users
		WHERE telegram_id = $1;
    `

	row := r.pool.QueryRow(ctx, query, telegramID)

	var user User
	err := row.Scan(
		&user.ID,
		&user.TelegramID,
		&user.GroupID,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
