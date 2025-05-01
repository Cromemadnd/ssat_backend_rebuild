package handlers

import (
	"net/http"

	"ssat_backend_rebuild/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DeviceHandler struct {
	DB *gorm.DB
}

// GetDevices 获取设备列表
func (h *DeviceHandler) GetDevices(c *gin.Context) {
	var devices []models.Device

	if result := h.DB.Find(&devices); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取设备列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, devices)
}

// GetDevice 获取单个设备
func (h *DeviceHandler) GetDevice(c *gin.Context) {
	id := c.Param("id")
	var device models.Device

	if result := h.DB.First(&device, "id = ?", id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "设备不存在",
		})
		return
	}

	c.JSON(http.StatusOK, device)
}

// CreateDevice 创建设备
func (h *DeviceHandler) CreateDevice(c *gin.Context) {
	var device models.Device

	if err := c.ShouldBindJSON(&device); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数无效",
		})
		return
	}

	if result := h.DB.Create(&device); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建设备失败",
		})
		return
	}

	c.JSON(http.StatusCreated, device)
}

// UpdateDevice 更新设备
func (h *DeviceHandler) UpdateDevice(c *gin.Context) {
	id := c.Param("id")
	var device models.Device

	if result := h.DB.First(&device, "id = ?", id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "设备不存在",
		})
		return
	}

	if err := c.ShouldBindJSON(&device); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数无效",
		})
		return
	}

	h.DB.Save(&device)
	c.JSON(http.StatusOK, device)
}

// DeleteDevice 删除设备
func (h *DeviceHandler) DeleteDevice(c *gin.Context) {
	id := c.Param("id")

	if result := h.DB.Delete(&models.Device{}, "id = ?", id); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "删除设备失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "设备已删除",
	})
}
