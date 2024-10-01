package security

import (
	"context"
	"errors"
	"github.com/allegro/bigcache/v3"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type IdentifyExtractFunc func(*gin.Context) string

type AuthorityContext struct {
	retriever           AuthorityRetriever
	cache               *bigcache.BigCache
	logger              *slog.Logger
	extractIdentityFunc IdentifyExtractFunc
}

var ErrUserNotLogin = errors.New("user not login")
var ErrNotPermitted = errors.New("operation not permitted")

func GetIdentity(c *gin.Context) string {
	return c.Request.Header.Get(HeaderAuthUser)
}

func Context(c *gin.Context) context.Context {
	ctx := c.Request.Context()
	return context.WithValue(ctx, ctxIdentity, GetIdentity(c))
}

func IdentityFromContext(c context.Context) string {
	return c.Value(ctxIdentity).(string)
}
func NewAuthorityContext(retriever AuthorityRetriever, slogger *slog.Logger, extractIdentityFunc IdentifyExtractFunc) (*AuthorityContext, func(), error) {
	cache, err := bigcache.New(context.TODO(), bigcache.DefaultConfig(1*time.Minute))
	if err != nil {
		return nil, nil, err
	}
	return &AuthorityContext{
			retriever:           retriever,
			cache:               cache,
			logger:              slogger,
			extractIdentityFunc: extractIdentityFunc,
		}, func() {
			_ = cache.Close()
		}, nil
}

type AuthorityRetriever interface {
	GetAuthorities(ctx context.Context, uid int64) ([]string, error)
}

func (ac *AuthorityContext) DecorateFunc(h gin.HandlerFunc, requires ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		idk := ac.extractIdentityFunc(c)
		if idk == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var authorities []string
		v, err := ac.cache.Get(idk)
		if err != nil {
			uid, err := strconv.ParseInt(idk, 10, 64)
			if err != nil {
				ac.logger.Error("failed to convert uid to int", "error", err)
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			authorities, err = ac.retriever.GetAuthorities(c.Request.Context(), uid)
			if err != nil {
				ac.logger.Error("failed to retriever user authorities", "error", err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			if lo.Contains(authorities, AuthorityAdmin) {
				h(c)
				return
			}
			err = ac.cache.Set(idk, []byte(strings.Join(authorities, ",")))
			if err != nil {
				ac.logger.Error("failed to set user authorities cache", "error", err)
			}
		} else {
			authorities = strings.Split(string(v), ",")
		}
		//if len(authorities) == 0 {
		authorities = append(authorities, AuthorityPublic)
		//}
		if !lo.Some(authorities, requires) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		h(c)
	}
}

const (
	AuthorityAdmin  = "admin"
	AuthorityPublic = "public"
	HeaderAuthUser  = "X-Auth-User"
)

var ctxIdentity = identityCtx{}

type identityCtx struct{}
