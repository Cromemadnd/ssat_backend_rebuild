package handlers

import (
	"ssat_backend_rebuild/models"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	BaseHandler[models.User]
}

func (h *UserHandler) Retrieve(c *gin.Context) {
	h.BaseHandler.Retrieve(
		nil,
		nil,
	)(c)
}

func (h *UserHandler) List(c *gin.Context) {
	h.BaseHandler.List(
		[]string{"uuid", "username", "is_admin"},
		nil,
	)(c)
}

func (h *UserHandler) Update(c *gin.Context) {
	h.BaseHandler.Update(
		[]string{"username", "is_admin"},
		nil,
		nil,
	)(c)
}

func (h *UserHandler) Destroy(c *gin.Context) {
	h.BaseHandler.Destroy(
		nil,
	)(c)
}
