package pubsub

import "context"

type Handler[T any] interface {
	EventName() string
	BinderFunc() func() T
	HandlerFunc() func(ctx context.Context, event T) error
}
