package domain

import (
	"context"
	"time"
)

type DocumentCache interface {
	GetDocument(ctx context.Context, id int64) (data []byte, found bool, err error)
	SetDocument(ctx context.Context, id int64, ownerLogin string, data []byte, ttl time.Duration) error

	GetList(ctx context.Context, ownerLogin, key string) (data []byte, found bool, err error)
	SetList(ctx context.Context, ownerLogin, key string, data []byte, ttl time.Duration) error

	InvalidateOwner(ctx context.Context, ownerLogin string) error
}
