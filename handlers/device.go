package handlers

import (
	"errors"
	"ssat_backend_rebuild/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeviceHandler struct {
	BaseHandler[models.Device]
}

func (h *DeviceHandler) Create(c *gin.Context) {
	h.BaseHandler.Create(
		nil,
		func(c *gin.Context, query *gorm.DB, device *models.Device, data map[string]any) error {
			device_id, ok := data["device_id"].(string)
			if !ok {
				return errors.New("device_id is required")
			}
			device.DeviceID = device_id

			secret, ok := data["secret"].(string)
			if !ok {
				return errors.New("secret is required")
			}
			device.Secret = secret

			return nil
		},
	)(c)
}

func (h *DeviceHandler) Retrieve(c *gin.Context) {
	h.BaseHandler.Retrieve(
		nil,
		nil,
	)(c)
}

func (h *DeviceHandler) List(c *gin.Context) {
	h.BaseHandler.List(
		[]string{"uuid", "device_id", "status"},
		nil,
	)(c)
}

func (h *DeviceHandler) Update(c *gin.Context) {
	h.BaseHandler.Update(
		[]string{"device_id", "status", "owner_id"},
		nil,
		nil,
	)(c)
}

func (h *DeviceHandler) Destroy(c *gin.Context) {
	h.BaseHandler.Destroy(
		nil,
	)(c)
}

func (h *DeviceHandler) Bind(c *gin.Context) {
	h.BaseHandler.Update(
		[]string{},
		nil,
		func(c *gin.Context, query *gorm.DB, device *models.Device, data map[string]any) error {
			// 绑定设备时，检查设备的拥有者是否为当前用户
			if device.OwnerID != nil && *device.OwnerID != uuid.Nil {
				return errors.New("设备已绑定")
			}
			// 获取当前用户的UUID
			currentUser := c.MustGet("currentUser").(*models.User)
			// 将当前用户的UUID绑定到设备上
			device.OwnerID = &currentUser.UUID
			device.Owner = currentUser
			return nil
		},
	)(c)
}

func (h *DeviceHandler) Unbind(c *gin.Context) {
	h.BaseHandler.Update(
		[]string{},
		nil,
		func(c *gin.Context, query *gorm.DB, device *models.Device, data map[string]any) error {
			// 解绑设备时，检查设备的拥有者是否为当前用户
			if device.OwnerID == nil || *device.OwnerID != c.MustGet("currentUser").(*models.User).UUID {
				return errors.New("设备未绑定为当前用户")
			}
			device.OwnerID = nil
			device.Owner = nil
			return nil
		},
	)(c)
}

func (h *DeviceHandler) MyDevices(c *gin.Context) {
	h.BaseHandler.List(
		[]string{"uuid", "status"},
		func(c *gin.Context, query *gorm.DB) *gorm.DB {
			return query.Where("owner_id = ?", c.MustGet("currentUser").(*models.User).UUID)
		},
	)(c)
}

func (h *DeviceHandler) RetrieveMyDevice(c *gin.Context) {
	h.BaseHandler.Retrieve(
		nil,
		func(c *gin.Context, query *gorm.DB) *gorm.DB {
			return query.Where("uuid = ? AND owner_id = ?", c.Param("uuid"), c.MustGet("currentUser").(*models.User).UUID)
		},
	)(c)
}
