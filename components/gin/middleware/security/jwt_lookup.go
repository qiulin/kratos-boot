package security

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"log/slog"
	"time"
)

func LookupJWT(c *gin.Context) (string, bool) {
	t := c.Request.Header.Get(HeaderAuthorization)
	if len(t) <= 7 {
		t = c.Query("token")
	}
	if len(t) <= 7 {
		return "", false
	}
	return t[7:], true // remove Bearer prefix
}

func JWTLookupFunc(secretKey string, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := c.Request.Header.Get(HeaderAuthorization)
		if t == "" || len(t) <= 7 {
			c.AbortWithStatus(401)
			return
		}
		t = t[7:] // remove Bearer prefix
		token, err := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})
		if err != nil {
			logger.Error("failed to parse jwt token", "error", err)
			c.AbortWithStatus(401)
			return
		}
		expiration, err := token.Claims.GetExpirationTime()
		if err != nil {
			c.AbortWithStatus(401)
			return
		}
		if expiration.Before(time.Now()) {
			c.AbortWithStatus(401)
			return
		}
		user, err := token.Claims.GetSubject()
		if err != nil {
			logger.Error("failed to get subject from jwt token", "error", err)
			c.AbortWithStatus(401)
			return
		}
		c.Set(ContextKeyIdentity, user)

		c.Next()
	}
}

const (
	HeaderAuthorization = "Authorization"
)
