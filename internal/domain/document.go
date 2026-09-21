package domain

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"time"
)

type Document struct {
	ID         int64
	OwnerLogin string
	Name       string
	Mime       string
	FilePath   string
	Public     bool
	Grant      []string
	JSONData   json.RawMessage
	CreatedAt  time.Time
}

type ListFilter struct {
	OwnerLogin     string
	RequesterLogin string
	FilterColumn   string
	FilterValue    any
	Limit          int
}

func (doc Document) CanAccess(requesterLogin string) bool {
	if doc.Public {
		return true
	}
	if doc.OwnerLogin == requesterLogin {
		return true
	}
	return slices.Contains(doc.Grant, requesterLogin)
}

var (
	ErrDocumentNotFound  = errors.New("Документ не найден")
	ErrDocumentForbidden = errors.New("У вас недостаточно прав для просмотра документа")
)

type DocumentRepository interface {
	Create(ctx context.Context, doc *Document) error
	GetByID(ctx context.Context, id int64) (*Document, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, filter ListFilter) ([]*Document, error)
}
