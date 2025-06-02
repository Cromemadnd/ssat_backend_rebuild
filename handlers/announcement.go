package handlers

import (
	"errors"
	"ssat_backend_rebuild/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AnnouncementHandler struct {
	BaseHandler[models.Announcement]
}

func (h *AnnouncementHandler) Create(c *gin.Context) {
	h.BaseHandler.Create(
		nil,
		func(c *gin.Context, query *gorm.DB, object *models.Announcement, data map[string]any) error {
			if typeFloat, ok := data["type"].(float64); ok {
				object.Type = uint8(typeFloat)
			} else {
				return errors.New("invalid type for 'type' field")
			}
			object.Title = data["title"].(string)
			object.Content = data["content"].(string)
			object.Publisher = c.MustGet("CurrentAdminUser").(*models.Admin).Username
			return nil
		},
	)(c)
}

func (h *AnnouncementHandler) List(c *gin.Context) {
	h.BaseHandler.List(
		[]string{"uuid", "created_at", "title", "publisher", "type", "modified_at"},
		nil,
	)(c)
}

func (h *AnnouncementHandler) Retrieve(c *gin.Context) {
	h.BaseHandler.Retrieve(
		nil,
		nil,
	)(c)
}

func (h *AnnouncementHandler) Update(c *gin.Context) {
	h.BaseHandler.Update(
		[]string{"title", "content", "type"},
		nil,
		nil,
	)(c)
}

func (h *AnnouncementHandler) Destroy(c *gin.Context) {
	h.BaseHandler.Destroy(
		nil,
	)(c)
}
