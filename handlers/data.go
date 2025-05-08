package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"ssat_backend_rebuild/models"
	"ssat_backend_rebuild/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

type DataHandler struct {
	BaseHandler[models.Data]
}

func (h *DataHandler) List(c *gin.Context) {
	h.BaseHandler.List(
		nil,
		nil,
	)(c)
}

type DataUploadRequest struct {
	DeviceID  string           `json:"device_id" binding:"required"`
	Timestamp int64            `json:"timestamp" binding:"required"`
	Data      models.DataEntry `json:"data" binding:"required"`
	Signature string           `json:"signature" binding:"required"`
}

var DataCache = cache.New(5*time.Minute, 10*time.Minute)

func (h *DataHandler) Upload(c *gin.Context) {
	// 解析请求体
	var reqBody DataUploadRequest
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		utils.Respond(c, nil, utils.ErrMissingParam)
		return
	}

	// 验证设备ID
	device := &models.Device{}
	if err := h.DB.First(device, "uuid = ?", reqBody.DeviceID).Error; err != nil {
		utils.Respond(c, nil, utils.ErrUnknownDevice)
		return
	}

	// 校验时间戳
	if time.Since(time.UnixMilli(reqBody.Timestamp)) > time.Minute {
		utils.Respond(c, nil, utils.ErrExpiredRequest)
		return
	}

	// 验证签名
	hash := md5.Sum([]byte(fmt.Sprintf("%s:%d:%s", reqBody.DeviceID, reqBody.Timestamp, device.Secret)))
	hashedSignature := hex.EncodeToString(hash[:])
	if reqBody.Signature != hashedSignature {
		utils.Respond(c, nil, utils.ErrInvalidSignature)
		return
	}

	// 记录签名，防止重放攻击
	if _, found := DataCache.Get(reqBody.Signature); found {
		utils.Respond(c, nil, utils.ErrReplayAttack)
		return
	}
	DataCache.Set(reqBody.Signature, true, 2*time.Minute)

	// 创建数据记录
	if err := h.DB.Create(&reqBody).Error; err != nil {
		utils.Respond(c, nil, utils.ErrInternalServer)
		return
	}

	utils.Respond(c, nil, utils.ErrOK)
}
