package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"TestTask/internal/domain"
	"TestTask/internal/dto"
)

type DocumentService struct {
	repo    domain.DocumentRepository
	storage domain.FileStorage
}

func NewDocumentService(repo domain.DocumentRepository, storage domain.FileStorage) *DocumentService {
	return &DocumentService{repo: repo, storage: storage}
}

func (s *DocumentService) Upload(ctx context.Context, in dto.UploadInput) (*domain.Document, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, errors.New("Имя документа не может быть пустым")
	}

	doc := &domain.Document{
		OwnerLogin: in.OwnerLogin,
		Name:       name,
		Mime:       in.Mime,
		Public:     in.Public,
		Grant:      in.Grant,
		JSONData:   in.JSONData,
	}

	if in.File != nil {
		path, err := s.storage.Save(ctx, name, in.File)
		if err != nil {
			return nil, err
		}
		doc.FilePath = path
	}

	if err := s.repo.Create(ctx, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (s *DocumentService) Get(ctx context.Context, requesterLogin string, id int64) (*domain.Document, error) {
	doc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !doc.CanAccess(requesterLogin) {
		return nil, domain.ErrDocumentForbidden
	}
	return doc, nil
}

// Grant не учитывается, тк он даёт право читать, а не удалять чужой документ
func (s *DocumentService) Delete(ctx context.Context, requesterLogin string, id int64) error {
	doc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if doc.OwnerLogin != requesterLogin {
		return domain.ErrDocumentForbidden
	}

	// Если удаление файла упадёт, лучше остаться с "висячей" записью в БД, чем с осиротевшим файлом на диске без метаданных,
	if doc.FilePath != "" {
		if err := s.storage.Delete(ctx, doc.FilePath); err != nil {
			return err
		}
	}

	return s.repo.Delete(ctx, id)
}

func (s *DocumentService) List(ctx context.Context, in dto.ListInput) ([]*domain.Document, error) {
	ownerLogin := in.OwnerLogin
	if ownerLogin == "" {
		ownerLogin = in.RequesterLogin
	}

	filter := domain.ListFilter{
		OwnerLogin:     ownerLogin,
		RequesterLogin: in.RequesterLogin,
		Limit:          in.Limit,
	}

	var listFilterColumns = map[string]string{
		"name":   "name",
		"mime":   "mime",
		"public": "is_public",
	}

	if in.Key != "" {
		column, ok := listFilterColumns[in.Key]
		if !ok {
			return nil, fmt.Errorf("Неподдерживаемый фильтр %q", in.Key)
		}
		filter.FilterColumn = column

		// приводим тип явно, иначе pgx откажется сравнивать bool со string
		if column == "is_public" {
			b, err := strconv.ParseBool(in.Value)
			if err != nil {
				return nil, fmt.Errorf("Значение ключа %q должно быть true/false", in.Key)
			}
			filter.FilterValue = b
		} else {
			filter.FilterValue = in.Value
		}
	}

	return s.repo.List(ctx, filter)
}
