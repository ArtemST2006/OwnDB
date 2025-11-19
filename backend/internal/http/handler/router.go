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
		auth := api.Group("/auth")
		{
			auth.POST("/sign-up", h.SignUp)
			auth.POST("/sign-in", h.SignIn)
		}

		article := api.Group("/article")
		{
			article.POST("/publish", h.Publish)
			article.GET("/allartic", h.GetArticles)
			article.GET("/import", h.Import)
		}
		services := api.Group("/services")
		{
			services.PATCH("/clear", h.clear)
			services.POST("/commit", h.commit)
		}
	}

	return router
}
