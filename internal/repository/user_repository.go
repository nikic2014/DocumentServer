package repository

import (
	"TestTask/internal/domain"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}

func (r UserRepository) Create(ctx context.Context, user *domain.User) error {
	const query = `INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id`

	err := r.pool.QueryRow(ctx, query, user.Login, user.PasswordHash).Scan(&user.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return domain.ErrLoginAlreadyExists
		}
		return err
	}
	return nil
}

func (r UserRepository) GetUserByLogin(ctx context.Context, login *string) (*domain.User, error) {
	const query = `SELECT id, login, password FROM users WHERE login = $1`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, login).
		Scan(&user.ID, &user.Login, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}
