package handler

import "github.com/gin-gonic/gin"

func RespondError(c *gin.Context, status int, err error) {
	c.JSON(status, gin.H{"error": err.Error()})
}
