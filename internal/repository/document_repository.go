package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"TestTask/internal/domain"
)

type DocumentRepository struct {
	pool *pgxpool.Pool
}

func NewDocumentRepository(pool *pgxpool.Pool) *DocumentRepository {
	return &DocumentRepository{pool: pool}
}

func (r *DocumentRepository) Create(ctx context.Context, doc *domain.Document) error {
	const query = `
		INSERT INTO documents (owner_login, name, mime, file_path, is_public, grant_logins, json_data)
		VALUES ($1, $2, $3, $4, $5, $6 ::jsonb, $7::jsonb)
		RETURNING id, created_at
	`

	grantJSON, err := json.Marshal(doc.Grant)
	if err != nil {
		return err
	}

	jsonData := doc.JSONData
	if len(jsonData) == 0 {
		jsonData = json.RawMessage("null")
	}

	return r.pool.QueryRow(
		ctx, query,
		doc.OwnerLogin, doc.Name, doc.Mime, doc.FilePath, doc.Public,
		string(grantJSON), string(jsonData),
	).Scan(&doc.ID, &doc.CreatedAt)
}

func (r *DocumentRepository) GetByID(ctx context.Context, id int64) (*domain.Document, error) {
	const query = `
		SELECT id, owner_login, name, mime, file_path, is_public, grant_logins, json_data, created_at
		FROM documents
		WHERE id = $1
	`

	var doc domain.Document
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&doc.ID, &doc.OwnerLogin, &doc.Name, &doc.Mime, &doc.FilePath,
		&doc.Public, &doc.Grant, &doc.JSONData, &doc.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDocumentNotFound
		}
		return nil, err
	}
	return &doc, nil
}

func (r *DocumentRepository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.Document, error) {
	query := `
		SELECT id, name, mime, file_path, is_public, grant_logins, created_at
		FROM documents
		WHERE owner_login = $1
		  AND (is_public = true OR owner_login = $2 OR grant_logins ? $2)
	`
	args := []any{filter.OwnerLogin, filter.RequesterLogin}

	if filter.FilterColumn != "" {
		args = append(args, filter.FilterValue)
		query += fmt.Sprintf(" AND %s = $%d", filter.FilterColumn, len(args))
	}

	query += " ORDER BY name, created_at"

	if filter.Limit > 0 {
		args = append(args, filter.Limit)
		query += fmt.Sprintf(" LIMIT $%d", len(args))
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*domain.Document
	for rows.Next() {
		var doc domain.Document
		if err := rows.Scan(
			&doc.ID, &doc.Name, &doc.Mime, &doc.FilePath,
			&doc.Public, &doc.Grant, &doc.CreatedAt,
		); err != nil {
			return nil, err
		}
		docs = append(docs, &doc)
	}
	return docs, rows.Err()
}

func (r *DocumentRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM documents WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrDocumentNotFound
	}
	return nil
}
