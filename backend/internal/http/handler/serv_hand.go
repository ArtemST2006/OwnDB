package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) clear(c *gin.Context) {
	if err := h.repo.Clear(); err != nil {
		NewResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

func (h *Handler) commit(c *gin.Context) {
	if err := h.repo.Flush(); err != nil {
		NewResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}
