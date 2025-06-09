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

var (
	AuthUserCache  = cache.New(5*time.Minute, 10*time.Minute)
	AuthAdminCache = cache.New(5*time.Minute, 10*time.Minute)
)

// 验证token并获取用户/管理员信息
func (m *AuthMiddleware) validateToken(c *gin.Context, tokenStr string, model interface{}, issuer string) (utils.ErrorCode, time.Duration) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.JWTSecret), nil
	},
		jwt.WithIssuedAt(),
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(issuer),
	)

	if err != nil {
		// 判断无效原因是否是token过期
		if errors.Is(err, jwt.ErrTokenExpired) {
			return utils.ErrExpiredJWT, 0
		}
		return utils.ErrInvalidJWT, 0
	}

	uuid, err := token.Claims.GetSubject()
	if err != nil {
		return utils.ErrInvalidJWT, 0
	}

	result := m.DB.First(model, "uuid = ?", uuid)
	if result.Error != nil {
		return utils.ErrUserNotFound, 0
	}

	// 获取 token 过期时间
	exp, err := token.Claims.GetExpirationTime()
	if err != nil {
		return utils.ErrInvalidJWT, 0
	}

	return utils.ErrOK, time.Until(exp.Time)
}

// 验证用户是否已登录
func (m *AuthMiddleware) UserOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.Request.Header.Get("Authorization")
		if tokenStr == "" {
			utils.Respond(c, nil, utils.ErrUnauthorized)
			return
		}
		if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
			tokenStr = tokenStr[7:]
		}

		user := &models.User{}
		if cached, found := AuthUserCache.Get(tokenStr); found {
			user = cached.(*models.User)
		} else {
			err, cacheDuration := m.validateToken(c, tokenStr, user, "ssat_user")
			AuthUserCache.Set(tokenStr, user, cacheDuration)
			if err != utils.ErrOK {
				utils.Respond(c, nil, err)
				return
			}
		}

		c.Set("CurrentUser", user)
		c.Next()
	}
}

// 仅管理员可访问
func (m *AuthMiddleware) AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.Request.Header.Get("Authorization")
		if tokenStr == "" {
			utils.Respond(c, nil, utils.ErrUnauthorized)
			return
		}
		if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
			tokenStr = tokenStr[7:]
		}

		admin := &models.Admin{}
		if cached, found := AuthAdminCache.Get(tokenStr); found {
			admin = cached.(*models.Admin)
		} else {
			err, cacheDuration := m.validateToken(c, tokenStr, admin, "ssat_admin")
			if err != utils.ErrOK {
				utils.Respond(c, nil, err)
				return
			}
			AuthAdminCache.Set(tokenStr, admin, cacheDuration)
		}

		c.Set("CurrentAdminUser", admin)
		c.Next()
	}
}

// 用户或管理员可访问
func (m *AuthMiddleware) UserOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.Request.Header.Get("Authorization")
		if tokenStr == "" {
			utils.Respond(c, nil, utils.ErrUnauthorized)
			return
		}
		if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
			tokenStr = tokenStr[7:]
		}

		admin := &models.Admin{}
		user := &models.User{}
		if cached, found := AuthAdminCache.Get(tokenStr); found {
			admin = cached.(*models.Admin)
			c.Set("CurrentAdminUser", admin)
		} else {
			err, cacheDuration := m.validateToken(c, tokenStr, admin, "ssat_admin")
			if err != utils.ErrOK {
				if cached, found := AuthUserCache.Get(tokenStr); found {
					user = cached.(*models.User)
				} else {
					err, cacheDuration = m.validateToken(c, tokenStr, user, "ssat_user")
					if err != utils.ErrOK {
						utils.Respond(c, nil, err)
						return
					}
				}
				AuthUserCache.Set(tokenStr, user, cacheDuration)
				c.Set("CurrentUser", user)
			} else {
				AuthAdminCache.Set(tokenStr, admin, cacheDuration)
				c.Set("CurrentAdminUser", admin)
			}
		}

		c.Next()
	}
}
