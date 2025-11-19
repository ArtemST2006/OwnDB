package handler

import "github.com/gin-gonic/gin"

type errorResponse struct {
	Message string `json:"message"`
}

func NewResponseError(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, errorResponse{message})
}
