package handler

import (
	"net/http"

	"github.com/ArtemST2006/OwnDB/backend/internal/schema"
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

func (h *Handler) createBackup(c *gin.Context) {
	if err := h.repo.CreateBackup(); err != nil {
		NewResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

func (h *Handler) importBackup(c *gin.Context) {
	if err := h.repo.ImportBackup(); err != nil {
		NewResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

func (h *Handler) delete(c *gin.Context) {
	del := schema.Delete{}
	if err := c.Bind(&del); err != nil {
		NewResponseError(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.repo.Delete(del); err != nil {
		NewResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}
