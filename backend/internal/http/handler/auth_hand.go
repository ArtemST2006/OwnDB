package handler

import (
	"net/http"

	"github.com/ArtemST2006/OwnDB/backend/internal/schema"
	"github.com/gin-gonic/gin"
)

func (h *Handler) SignUp(c *gin.Context) {
	sign := schema.SignUp{}
	if err := c.Bind(&sign); err != nil {
		NewResponseError(c, http.StatusBadRequest, err.Error())
		return
	}

	_, err := h.repo.Authorization.AddUser(sign)
	if err != nil {
		NewResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

func (h *Handler) SignIn(c *gin.Context) {
	sign := schema.SignIn{}
	if err := c.Bind(&sign); err != nil {
		NewResponseError(c, http.StatusBadRequest, err.Error())
		return
	}

	id, access, err := h.repo.Authorization.Login(sign)
	if err != nil {
		NewResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"id":      id,
		"access":  access,
	})
}
