package handlers

import (
	"ssat_backend_rebuild/models"
	"ssat_backend_rebuild/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	DB *gorm.DB
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	var users []models.User

	if result := h.DB.Find(&users); result.Error != nil {
		utils.Respond(c, nil, utils.ErrNotFound)
		return
	}

	utils.Respond(c, users, utils.ErrOK)
}
