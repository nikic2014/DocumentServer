package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"TestTask/internal/domain"
)

// CachedDocumentRepository оборачивает любой domain.DocumentRepository кэшем поверх чтения (GetByID/List)
// и выборочной инвалидацией при записи (Create/Delete).
type CachedDocumentRepository struct {
	next  domain.DocumentRepository
	cache domain.DocumentCache
	ttl   time.Duration
}

func NewCachedDocumentRepository(next domain.DocumentRepository, cache domain.DocumentCache, ttl time.Duration) *CachedDocumentRepository {
	return &CachedDocumentRepository{next: next, cache: cache, ttl: ttl}
}

func (r *CachedDocumentRepository) Create(ctx context.Context, doc *domain.Document) error {
	if err := r.next.Create(ctx, doc); err != nil {
		return err
	}
	// ошибка инвалидации не должна ронять уже созданный документ, поэтому дальше она не идет
	_ = r.cache.InvalidateOwner(ctx, doc.OwnerLogin)
	return nil
}

func (r *CachedDocumentRepository) GetByID(ctx context.Context, id int64) (*domain.Document, error) {
	if raw, ok, err := r.cache.GetDocument(ctx, id); err == nil && ok {
		var doc domain.Document
		if err := json.Unmarshal(raw, &doc); err == nil {
			return &doc, nil
		}
	}

	doc, err := r.next.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if raw, err := json.Marshal(doc); err == nil {
		_ = r.cache.SetDocument(ctx, id, doc.OwnerLogin, raw, r.ttl)
	}
	return doc, nil
}

func (r *CachedDocumentRepository) Delete(ctx context.Context, id int64) error {
	doc, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := r.next.Delete(ctx, id); err != nil {
		return err
	}
	return r.cache.InvalidateOwner(ctx, doc.OwnerLogin)
}

func (r *CachedDocumentRepository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.Document, error) {
	key := listCacheKey(filter)

	if raw, ok, err := r.cache.GetList(ctx, filter.OwnerLogin, key); err == nil && ok {
		var docs []*domain.Document
		if err := json.Unmarshal(raw, &docs); err == nil {
			return docs, nil
		}
	}

	docs, err := r.next.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	if raw, err := json.Marshal(docs); err == nil {
		_ = r.cache.SetList(ctx, filter.OwnerLogin, key, raw, r.ttl)
	}
	return docs, nil
}

func listCacheKey(filter domain.ListFilter) string {
	return fmt.Sprintf("requester=%s|col=%s|val=%v|limit=%d",
		filter.RequesterLogin, filter.FilterColumn, filter.FilterValue, filter.Limit)
}
