package pubsub

import (
	"context"
	"github.com/go-kratos/kratos/v2/transport"
)

type Emitter interface {
	transport.Server
	Emit(ctx context.Context, eventName string, event any) error
}
