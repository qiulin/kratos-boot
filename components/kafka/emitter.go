package kafka

import (
	"context"
	"errors"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/qiulin/kratos-boot/sharedconf"
	"github.com/tx7do/kratos-transport/broker"
	"github.com/tx7do/kratos-transport/broker/kafka"
	"golang.org/x/sync/errgroup"
	"log/slog"
)

var _ transport.Server = new(Emitter)

type Emitter struct {
	publishers map[string]broker.Broker
	logger     *slog.Logger
}

func (p *Emitter) Start(ctx context.Context) error {
	var eg *errgroup.Group
	eg, ctx = errgroup.WithContext(ctx)
	for i := range p.publishers {
		pub := p.publishers[i]
		eg.Go(func() error {
			if err := pub.Init(); err != nil {
				return err
			}
			return pub.Connect()
		})
	}
	return eg.Wait()
}

func (p *Emitter) Stop(ctx context.Context) error {

	var eg *errgroup.Group
	eg, ctx = errgroup.WithContext(ctx)
	for i := range p.publishers {
		pub := p.publishers[i]
		eg.Go(func() error {
			return pub.Disconnect()
		})
	}
	return eg.Wait()
}

var ErrTopicNotFound = errors.New("topic not found")

func (p *Emitter) Emit(ctx context.Context, eventName string, event any) error {
	pub, ok := p.publishers[eventName]
	if !ok {
		return ErrTopicNotFound
	}
	return pub.Publish(ctx, eventName, event)
}

func (p *Emitter) Publisher(name string) broker.Broker {
	return p.publishers[name]
}

func NewPublishers(cs map[string]*sharedconf.Kafka, logger *slog.Logger) *Emitter {
	publishers := map[string]broker.Broker{}
	for k, c := range cs {
		if c.Producer == nil {
			continue
		}
		b := newPublisher(c)
		publishers[k] = b
	}
	return &Emitter{publishers: publishers, logger: logger}
}

func newPublisher(c *sharedconf.Kafka) broker.Broker {
	codec := "json"
	if c.Producer != nil && c.Producer.Codec != "" {
		codec = c.Producer.Codec
	}
	b := kafka.NewBroker(
		broker.WithAddress(c.Servers...),
		broker.WithCodec(codec),
	)
	return b
}
