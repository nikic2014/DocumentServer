package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"TestTask/internal/domain"
	"TestTask/internal/dto"
	"TestTask/internal/service"
)

type AuthService interface {
	Register(ctx context.Context, username, password string) (*domain.User, error)
	Login(ctx context.Context, username, password string) (string, error)
	Logout(ctx context.Context, tokenString string) error
}

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) RegisterRoutes(public, protected *gin.RouterGroup) {
	public.POST("/auth/register", h.Register)
	public.POST("/auth/login", h.Login)
	protected.DELETE("/auth/logout", h.Logout)
	protected.GET("/auth/me", h.Me)
}

func extractBearerToken(c *gin.Context) (string, bool) {
	header := c.GetHeader("Authorization")
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	return parts[1], true
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}

	user, err := h.auth.Register(c.Request.Context(), req.Login, req.Password)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, domain.ErrLoginAlreadyExists) {
			status = http.StatusBadRequest
		}
		RespondError(c, status, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"username": user.Login,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}

	token, err := h.auth.Login(c.Request.Context(), req.Login, req.Password)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	tokenString, ok := extractBearerToken(c)
	if !ok {
		RespondError(c, http.StatusUnauthorized, errors.New("Отсутсвует токен авторизации"))
		return
	}

	if err := h.auth.Logout(c.Request.Context(), tokenString); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "выход успешно выполнен"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"user_id":  c.GetInt64("user_id"),
		"username": c.GetString("username"),
	})
}
