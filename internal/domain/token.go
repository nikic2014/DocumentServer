package domain

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"
)

func GenerateTokenID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

type TokenStore interface {
	Store(ctx context.Context, tokenID string, expiresAt time.Time) error
	Exists(ctx context.Context, tokenID string) (bool, error)
	Delete(ctx context.Context, tokenID string) error
}
