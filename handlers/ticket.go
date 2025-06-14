package handlers

import (
	"ssat_backend_rebuild/models"
	"ssat_backend_rebuild/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TicketHandler struct {
	BaseHandler[models.Ticket]
}

func (h *TicketHandler) Create(c *gin.Context) {
	h.BaseHandler.Create(
		nil,
		func(c *gin.Context, query *gorm.DB, object *models.Ticket, data map[string]any) error {
			object.Title = data["title"].(string)
			if object.Title == "" {
				return utils.ErrMissingParam
			}

			object.Content = data["content"].(string)
			if object.Content == "" {
				return utils.ErrMissingParam
			}

			if typeStr, ok := data["feedback_type"].(string); ok {
				if t, err := strconv.ParseUint(typeStr, 10, 8); err == nil {
					object.Type = uint8(t)
				} else {
					return utils.ErrBadRequest
				}
			} else {
				object.Type = 0 // 默认类型
			}

			if deviceUUIDStr, ok := data["device_uuid"].(string); ok && deviceUUIDStr != "" {
				if deviceUUID, err := uuid.Parse(deviceUUIDStr); err == nil {
					object.DeviceUUID = &deviceUUID
				} else {
					return utils.ErrBadRequest
				}
			} else {
				object.DeviceUUID = nil // 如果没有提供设备UUID，则设置为nil
			}

			object.User = c.MustGet("CurrentUser").(*models.User)
			object.Status = 0 // 默认状态为未处理
			return nil
		},
	)(c)
}

func (h *TicketHandler) List(c *gin.Context) {
	h.BaseHandler.List(
		[]string{"uuid", "created_at", "title", "user_uuid", "type", "status"},
		func(c *gin.Context, query *gorm.DB) *gorm.DB {
			// 按Params过滤
			if status := c.Query("status"); status != "" {
				if s, err := strconv.ParseUint(status, 10, 8); err == nil {
					query = query.Where("status = ?", s)
				}
			}

			if typeStr := c.Query("type"); typeStr != "" {
				if t, err := strconv.ParseUint(typeStr, 10, 8); err == nil {
					query = query.Where("type = ?", uint8(t))
				}
			}

			if userUUID := c.Query("user_uuid"); userUUID != "" {
				query = query.Where("user_uuid = ?", userUUID)
			}

			return query
		},
	)(c)
}

func (h *TicketHandler) ListMyTickets(c *gin.Context) {
	h.BaseHandler.List(
		[]string{"uuid", "created_at", "title", "user_uuid", "type", "status"},
		func(c *gin.Context, query *gorm.DB) *gorm.DB {
			user := c.MustGet("CurrentUser").(*models.User)
			query = query.Where("user_uuid = ?", user.UUID)

			// 按Params过滤
			if status := c.Query("status"); status != "" {
				if s, err := strconv.ParseUint(status, 10, 8); err == nil {
					query = query.Where("status = ?", s)
				}
			}

			if typeStr := c.Query("type"); typeStr != "" {
				if t, err := strconv.ParseUint(typeStr, 10, 8); err == nil {
					query = query.Where("type = ?", uint8(t))
				}
			}

			return query
		},
	)(c)
}

func (h *TicketHandler) Retrieve(c *gin.Context) {
	h.BaseHandler.Retrieve(
		nil,
		func(c *gin.Context, query *gorm.DB) *gorm.DB {
			return query.Preload("ChatHistory").Where("uuid = ?", c.Param("uuid"))
		},
	)(c)
}

func (h *TicketHandler) RetrieveMyTicket(c *gin.Context) {
	h.BaseHandler.Retrieve(
		nil,
		func(c *gin.Context, query *gorm.DB) *gorm.DB {
			user := c.MustGet("CurrentUser").(*models.User)
			return query.Preload("ChatHistory").Where("user_uuid = ? AND uuid = ?", user.UUID, c.Param("uuid"))
		},
	)(c)
}

func (h *TicketHandler) Reply(c *gin.Context) {
	h.BaseHandler.Update(
		[]string{},
		nil,
		func(c *gin.Context, query *gorm.DB, object *models.Ticket, data map[string]any) error {
			if data["content"] == nil {
				return utils.ErrMissingParam
			}

			if object.Status == 3 {
				return utils.ErrClosedTicket
			}

			chat := models.TicketChat{
				Ticket:  object,
				Type:    1, // 管理员消息
				Subject: c.MustGet("CurrentAdminUser").(*models.Admin).Username,
				Content: data["content"].(string),
			}
			if err := h.DB.Create(&chat).Error; err != nil {
				return err
			}
			object.Status = 1 // 更新状态为等待用户回应
			object.ChatHistory = append(object.ChatHistory, chat)
			return query.Save(object).Error
		},
	)(c)
}

func (h *TicketHandler) Supply(c *gin.Context) {
	h.BaseHandler.Update(
		[]string{},
		nil,
		func(c *gin.Context, query *gorm.DB, object *models.Ticket, data map[string]any) error {
			if data["content"] == nil {
				return utils.ErrMissingParam
			}

			if object.Status == 3 {
				return utils.ErrClosedTicket
			}

			chat := models.TicketChat{
				Ticket:  object,
				Type:    0, // 用户消息
				Content: data["content"].(string),
			}
			if err := h.DB.Create(&chat).Error; err != nil {
				return err
			}
			object.Status = 2 // 更新状态为等待管理员回应
			object.ChatHistory = append(object.ChatHistory, chat)
			return query.Save(object).Error
		},
	)(c)
}

func (h *TicketHandler) Close(c *gin.Context) {
	h.BaseHandler.Update(
		nil,
		nil,
		func(c *gin.Context, query *gorm.DB, object *models.Ticket, data map[string]any) error {
			if object.Status == 3 {
				return utils.ErrClosedTicket
			}

			object.Status = 3 // 更新状态为已关闭
			return query.Save(object).Error
		},
	)(c)
}
