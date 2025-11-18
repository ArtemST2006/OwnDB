package handler

import (
	"net/http"

	"github.com/ArtemST2006/OwnDB/backend/internal/schema"
	"github.com/gin-gonic/gin"
)

func (h *Handler) create(c *gin.Context) { // создвние очередной базы данных

	var cr schema.Create = schema.Create{}
	if err := c.Bind(&cr); err != nil {
		NewResponseError(c, http.StatusBadRequest, err.Error())
	}

	if err := h.repo.Create(cr.Name, cr.Timestamp); err != nil {
		NewResponseError(c, http.StatusInternalServerError, err.Error())
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "created",
	})
}
