package gin

import (
	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"
	"github.com/things-go/gin-contrib/traceid"
	"log/slog"
)

const TraceIDHeader = "X-Trace-Id"

func NewRouter(mode string, slogger *slog.Logger) *gin.Engine {
	gin.SetMode(mode)
	r := gin.New()
	r.Use(sloggin.New(slogger))
	r.Use(gin.Recovery())

	r.Use(traceid.TraceId(traceid.WithTraceIdHeader(TraceIDHeader)))
	return r
}
