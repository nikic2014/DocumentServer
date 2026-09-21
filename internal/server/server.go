package server

import (
	"TestTask/internal/handler"
	"TestTask/internal/middleware"
	"TestTask/internal/repository"
	"TestTask/internal/service"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func New(jwtSecret string, pool *pgxpool.Pool, redisClient *redis.Client, uploadDir string) (*gin.Engine, error) {
	users := repository.NewUserRepository(pool)
	tokens := repository.NewRedisTokenRepository(redisClient)
	authService := service.NewAuthService(users, tokens, jwtSecret, 15*time.Minute)
	authHandler := handler.NewAuthHandler(authService)

	fileStorage, err := repository.NewLocalFileStorage(uploadDir)
	if err != nil {
		return nil, fmt.Errorf("Создание файлового хранилища: %w", err)
	}

	documents := repository.NewDocumentRepository(pool)
	documentCache := repository.NewDocumentCache(redisClient)
	cachedDocuments := repository.NewCachedDocumentRepository(documents, documentCache, 5*time.Minute)
	documentService := service.NewDocumentService(cachedDocuments, fileStorage)
	documentHandler := handler.NewDocumentHandler(documentService)

	r := gin.Default()

	r.HandleMethodNotAllowed = true
	r.NoMethod(func(c *gin.Context) {
		handler.RespondError(c, http.StatusMethodNotAllowed, errors.New("Метод не доступен по данному пути"))
	})

	r.NoRoute(func(c *gin.Context) {
		handler.RespondError(c, http.StatusNotImplemented, errors.New("Метод не реализован"))
	})

	public := r.Group("/api")
	protected := r.Group("/api")
	protected.Use(middleware.Auth(authService))

	authHandler.RegisterRoutes(public, protected)
	documentHandler.RegisterRoutes(protected)

	return r, nil
}
