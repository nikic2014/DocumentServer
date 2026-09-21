package handler

import "github.com/gin-gonic/gin"

type ErrorInfo struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

type WrapResponse struct {
	Error    *ErrorInfo `json:"error,omitempty"`
	Response any        `json:"response,omitempty"`
	Data     any        `json:"data,omitempty"`
}

func RespondError(c *gin.Context, status int, text string) {
	c.JSON(status, ErrorWrap(status, text))
}

func ErrorWrap(status int, text string) WrapResponse {
	return WrapResponse{Error: &ErrorInfo{Code: status, Text: text}}
}

func RespondResponse(c *gin.Context, status int, response any) {
	c.JSON(status, WrapResponse{Response: response})
}

func RespondData(c *gin.Context, status int, data any) {
	c.JSON(status, WrapResponse{Data: data})
}
