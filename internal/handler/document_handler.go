package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"TestTask/internal/domain"
	"TestTask/internal/dto"
	"TestTask/internal/service"
)

type DocumentService interface {
	Upload(ctx context.Context, in dto.UploadInput) (*domain.Document, error)
	Get(ctx context.Context, requesterLogin string, id int64) (*domain.Document, error)
	Delete(ctx context.Context, requesterLogin string, id int64) error
	List(ctx context.Context, in dto.ListInput) ([]*domain.Document, error)
}

type DocumentHandler struct {
	docs *service.DocumentService
}

func NewDocumentHandler(docs *service.DocumentService) *DocumentHandler {
	return &DocumentHandler{docs: docs}
}

func (h *DocumentHandler) RegisterRoutes(protected *gin.RouterGroup) {
	protected.POST("/docs", h.Upload)
	protected.GET("/docs", h.List)
	protected.HEAD("/docs", h.List)
	protected.GET("/docs/:id", h.Get)
	protected.HEAD("/docs/:id", h.Get)
	protected.DELETE("/docs/:id", h.Delete)
}

func (h *DocumentHandler) Upload(c *gin.Context) {
	var meta dto.UploadMeta
	if err := json.Unmarshal([]byte(c.PostForm("meta")), &meta); err != nil {
		RespondError(c, http.StatusBadRequest, "Неверный meta json")
		return
	}
	if meta.Name == "" {
		RespondError(c, http.StatusBadRequest, "meta.name обязательно для заполнения")
		return
	}

	var jsonData json.RawMessage
	if raw := c.PostForm("json"); raw != "" {
		jsonData = json.RawMessage(raw)
	}

	input := dto.UploadInput{
		OwnerLogin: c.GetString("username"),
		Name:       meta.Name,
		Mime:       meta.Mime,
		Public:     meta.Public,
		Grant:      meta.Grant,
		JSONData:   jsonData,
	}

	if meta.File {
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			RespondError(c, http.StatusBadRequest, "Поле file обязательно для заполнения, если meta.file имеет значение true.")
			return
		}
		defer file.Close()

		input.File = file
		if input.Mime == "" {
			input.Mime = header.Header.Get("Content-Type")
		}
	}

	doc, err := h.docs.Upload(c.Request.Context(), input)
	if err != nil {
		RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	RespondData(c, http.StatusOK, gin.H{
		"json": doc.JSONData,
		"file": doc.Name,
	})
}

func (h *DocumentHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "Неверный id документа")
		return
	}

	doc, err := h.docs.Get(c.Request.Context(), c.GetString("username"), id)
	if err != nil {
		h.respondDocError(c, err)
		return
	}

	if doc.FilePath != "" {
		c.Header("Content-Type", doc.Mime)
		c.File(doc.FilePath)
		return
	}

	if c.Request.Method == http.MethodHead {
		c.Status(http.StatusOK)
		return
	}
	RespondData(c, http.StatusOK, doc.JSONData)
}

func (h *DocumentHandler) List(c *gin.Context) {
	limit := 0
	if raw := c.Query("limit"); raw != "" {
		l, err := strconv.Atoi(raw)
		if err != nil || l < 0 {
			RespondError(c, http.StatusBadRequest, "Неверный лимит")
			return
		}
		limit = l
	}

	docs, err := h.docs.List(c.Request.Context(), dto.ListInput{
		RequesterLogin: c.GetString("username"),
		OwnerLogin:     c.Query("login"),
		Key:            c.Query("key"),
		Value:          c.Query("value"),
		Limit:          limit,
	})
	if err != nil {
		RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	if c.Request.Method == http.MethodHead {
		c.Status(http.StatusOK)
		return
	}

	items := make([]gin.H, 0, len(docs))
	for _, doc := range docs {
		items = append(items, gin.H{
			"id":      strconv.FormatInt(doc.ID, 10),
			"name":    doc.Name,
			"mime":    doc.Mime,
			"file":    doc.FilePath != "",
			"public":  doc.Public,
			"created": doc.CreatedAt.Format("2006-01-02 15:04:05"),
			"grant":   doc.Grant,
		})
	}

	RespondData(c, http.StatusOK, gin.H{"docs": items})
}

func (h *DocumentHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "Неверный id документа")
		return
	}

	if err := h.docs.Delete(c.Request.Context(), c.GetString("username"), id); err != nil {
		h.respondDocError(c, err)
		return
	}

	RespondResponse(c, http.StatusOK, gin.H{idParam: true})
}

func (h *DocumentHandler) respondDocError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrDocumentNotFound):
		RespondError(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrDocumentForbidden):
		RespondError(c, http.StatusForbidden, err.Error())
	default:
		RespondError(c, http.StatusInternalServerError, err.Error())
	}
}
