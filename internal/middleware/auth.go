package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"TestTask/internal/handler"
	"TestTask/internal/service"
)

type TokenParser interface {
	ParseToken(ctx context.Context, tokenString string) (*service.Claims, error)
}

func Auth(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, handler.ErrorWrap(http.StatusUnauthorized, "Отсутсвует токен авторизации"))
			return
		}

		claims, err := auth.ParseToken(c.Request.Context(), parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, handler.ErrorWrap(http.StatusUnauthorized, err.Error()))
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
