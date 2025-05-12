package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"ssat_backend_rebuild/models"
	"ssat_backend_rebuild/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB           *gorm.DB
	JWTSecret    string
	JWTExpires   int
	WechatAppID  string
	WechatSecret string
}

type LoginRequestBody struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	authCookie, err := c.Cookie("Authorization")
	if authCookie != "" && err == nil {
		utils.Respond(c, nil, utils.ErrAlreadyLoggedIn)
		return
	}

	var loginRequestBody LoginRequestBody
	if err := c.ShouldBindJSON(&loginRequestBody); err != nil {
		utils.Respond(c, nil, utils.ErrMissingParam)
		return
	}

	username := loginRequestBody.Username
	password := loginRequestBody.Password

	hash := md5.Sum([]byte(password))
	hashedPassword := hex.EncodeToString(hash[:])

	user := &models.User{}
	result := h.DB.First(user, "username = ? AND hashed_password = ?", username, hashedPassword)
	if result.Error != nil {
		utils.Respond(c, nil, utils.ErrIncorrectAuthInfo)
		return
	}

	claims := &jwt.RegisteredClaims{
		Issuer:    "ssat_env_monitor",
		Subject:   user.UUID.String(),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(h.JWTExpires) * time.Second)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString([]byte(h.JWTSecret))
	if err != nil {
		utils.Respond(c, nil, utils.ErrInternalServer)
		return
	}

	c.SetCookie("Authorization", tokenStr, 3600, "/", "", false, true)
	utils.Respond(c, user, utils.ErrOK)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var registerRequestBody LoginRequestBody
	if err := c.ShouldBindJSON(&registerRequestBody); err != nil {
		utils.Respond(c, nil, utils.ErrMissingParam)
		return
	}

	username := registerRequestBody.Username
	password := registerRequestBody.Password

	hash := md5.Sum([]byte(password))
	hashedPassword := hex.EncodeToString(hash[:])

	user := &models.User{
		Username:       username,
		HashedPassword: hashedPassword,
	}

	if result := h.DB.First(user, "username = ?", username); result.Error == nil {
		utils.Respond(c, nil, utils.ErrUserExists)
		return
	}

	if result := h.DB.Create(user); result.Error != nil {
		utils.Respond(c, nil, utils.ErrInternalServer)
		return
	}

	utils.Respond(c, user, utils.ErrOK)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("Authorization", "", -1, "/", "", false, true)
	utils.Respond(c, nil, utils.ErrOK)
}

func (h *AuthHandler) WechatLogin(c *gin.Context) {
	type WechatLoginRequest struct {
		Code string `json:"code" binding:"required"`
	}
	var req WechatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, utils.ErrMissingParam)
		return
	}

	// 1. 用 code 换 openid 和 session_key
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		h.WechatAppID, h.WechatSecret, req.Code,
	)

	resp, err := http.Get(url)
	if err != nil {
		utils.Respond(c, nil, utils.ErrInternalServer)
		return
	}
	defer resp.Body.Close()

	var wxResp struct {
		OpenID     string `json:"openid"`
		SessionKey string `json:"session_key"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wxResp); err != nil {
		utils.Respond(c, nil, utils.ErrInternalServer)
		return
	}
	if wxResp.ErrCode != 0 {
		utils.Respond(c, nil, utils.ErrBadRequest)
		return
	}

	// 2. 查找或创建本地用户
	user := &models.User{}
	result := h.DB.First(user, "username = ?", wxResp.OpenID)
	if result.Error != nil {
		user = &models.User{
			Username: wxResp.OpenID,
		}
		if err := h.DB.Create(user).Error; err != nil {
			utils.Respond(c, nil, utils.ErrInternalServer)
			return
		}
	}

	// 3. 生成JWT
	claims := &jwt.RegisteredClaims{
		Issuer:    "ssat_env_monitor",
		Subject:   user.UUID.String(),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(h.JWTExpires) * time.Second)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(h.JWTSecret))
	if err != nil {
		utils.Respond(c, nil, utils.ErrInternalServer)
		return
	}

	c.SetCookie("Authorization", tokenStr, 3600, "/", "", false, true)
	utils.Respond(c, user, utils.ErrOK)
}
