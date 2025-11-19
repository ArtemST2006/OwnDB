package handler

import (
	"net/http"

	"github.com/ArtemST2006/OwnDB/backend/internal/schema"
	"github.com/gin-gonic/gin"
)

func (h *Handler) Publish(c *gin.Context) {
	var pb schema.Publish = schema.Publish{}

	if err := c.Bind(&pb); err != nil {
		NewResponseError(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.repo.Central.Publish(pb); err != nil {
		NewResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mesage": "ok",
	})
}

func (h *Handler) GetArticles(c *gin.Context) {
	var allartic schema.AllArticles = schema.AllArticles{}

	if err := h.repo.Central.GetArticles(&allartic); err != nil {
		NewResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, allartic)
}

func (h *Handler) Import(c *gin.Context) {
	f, err := h.repo.Central.Import()
	if err != nil {
		NewResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=data.xlsx")
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}
