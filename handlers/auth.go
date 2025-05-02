package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"ssat_backend_rebuild/models"
	"ssat_backend_rebuild/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB         *gorm.DB
	JWTSecret  string
	JWTExpires int
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
