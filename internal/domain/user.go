package domain

import (
	"context"
	"errors"
)

type User struct {
	ID           int64
	Login        string
	PasswordHash string
}

var (
	ErrLoginAlreadyExists = errors.New("Логин уже используется")
	ErrUserNotFound       = errors.New("Пользователь не найден")
	ErrInvalidPassword    = errors.New("Некорректный пароль")
)

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetUserByLogin(ctx context.Context, login *string) (*User, error)
}
