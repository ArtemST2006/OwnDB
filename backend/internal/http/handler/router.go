package handler

import (
	"github.com/ArtemST2006/OwnDB/backend/internal/repository"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo *repository.Repository
}

func NewHandler(repo *repository.Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	api := router.Group("/api")
	{
		api.POST("/create", h.create)
	}

	return router
}
