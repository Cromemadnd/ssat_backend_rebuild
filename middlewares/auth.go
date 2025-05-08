package middlewares

import (
	"errors"
	"ssat_backend_rebuild/models"
	"ssat_backend_rebuild/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"
)

type AuthMiddleware struct {
	DB        *gorm.DB
	JWTSecret string
}

var AuthCache = cache.New(5*time.Minute, 10*time.Minute)

// Authenticate 验证用户是否已登录
func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		tokenStr, err := c.Cookie("Authorization")

		if err != nil {
			utils.Respond(c, nil, utils.ErrUnauthorized)
			return
		}

		user := &models.User{}
		if user, found := AuthCache.Get(tokenStr); !found {
			// 验证token并获取用户
			claims := &jwt.RegisteredClaims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(m.JWTSecret), nil
			})

			if err != nil {
				if errors.Is(err, jwt.ErrTokenExpired) {
					c.SetCookie("Authorization", "", -1, "/", "", false, true)
					utils.Respond(c, nil, utils.ErrExpiredJWT)
					return
				}
				utils.Respond(c, nil, utils.ErrInvalidJWT)
				return
			}

			uuid, err := token.Claims.GetSubject()
			if err != nil {
				utils.Respond(c, nil, utils.ErrInvalidJWT)
				return
			}

			result := m.DB.First(user, "uuid = ?", uuid)
			if result.Error != nil {
				utils.Respond(c, nil, utils.ErrUserNotFound)
				return
			}

			// 获取 token 过期时间
			exp, err := token.Claims.GetExpirationTime()

			if err != nil {
				utils.Respond(c, nil, utils.ErrInvalidJWT)
				return
			}
			cacheDuration := time.Until(exp.Time)
			AuthCache.Set(tokenStr, user, cacheDuration)
		}

		// 将用户信息存储到上下文中
		c.Set("currentUser", user)
		c.Next()
	}
}

// AdminOnly 仅管理员可访问
func (a *AuthMiddleware) AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := c.MustGet("currentUser").(*models.User)

		if !user.IsAdmin {
			utils.Respond(c, nil, utils.ErrForbidden)
			return
		}

		c.Next()
	}
}
