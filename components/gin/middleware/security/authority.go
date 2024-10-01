package security

import (
	"context"
	"errors"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"log/slog"
	"net/http"
)

type IdentifyExtractFunc[T comparable] func(*gin.Context) (T, bool)

type AuthorityDecorator[T comparable] struct {
	retriever           AuthorityRetriever[T]
	cache               *cache.LoadableCache[[]string]
	logger              *slog.Logger
	extractIdentityFunc IdentifyExtractFunc[T]
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

func newAuthorityCache[T comparable](retriever AuthorityRetriever[T]) (*cache.LoadableCache[[]string], error) {
	return nil, errors.New("TODO")
}

func NewAuthorityDecorator[T comparable](retriever AuthorityRetriever[T], slogger *slog.Logger, extractIdentityFunc IdentifyExtractFunc[T]) (*AuthorityDecorator[T], func(), error) {
	ca, err := newAuthorityCache(retriever)
	if err != nil {
		return nil, nil, err
	}
	return &AuthorityDecorator[T]{
			retriever:           retriever,
			cache:               ca,
			logger:              slogger,
			extractIdentityFunc: extractIdentityFunc,
		}, func() {
			_ = ca.Close()
		}, nil
}

type AuthorityRetriever[T comparable] interface {
	GetAuthorities(ctx context.Context, uid T) ([]string, error)
}

func (ac *AuthorityDecorator[T]) DecorateFunc(h gin.HandlerFunc, requires ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		identity, ok := ac.extractIdentityFunc(c)
		if !ok && !lo.Contains(requires, AuthorityAnonymous) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		authorities, err := ac.cache.Get(c.Request.Context(), identity)
		if err != nil {
			ac.logger.Error("failed to set user authorities cache", "error", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		authorities = append(authorities, AuthorityAnonymous)
		if !lo.Contains(authorities, AuthorityRoot) && !lo.Some(authorities, requires) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		h(c)
	}
}

const (
	AuthorityRoot      = "root"
	AuthorityAnonymous = "anonymous"
	HeaderAuthUser     = "X-Auth-User"
	ContextKeyIdentity = "_identity"
)

var ctxIdentity = identityCtx{}

type identityCtx struct{}
