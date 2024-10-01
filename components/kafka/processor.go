package kafka

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/panjf2000/ants/v2"
	"github.com/qiulin/kratos-boot/sharedconf"
	"github.com/tx7do/kratos-transport/broker"
	"github.com/tx7do/kratos-transport/transport/kafka"
	"golang.org/x/sync/errgroup"
)

var _ transport.Server = new(Processor)

type Processor struct {
	servers []*kafka.Server
}

func (s *Processor) Start(ctx context.Context) error {
	var eg *errgroup.Group
	eg, ctx = errgroup.WithContext(ctx)
	for i := range s.servers {
		srv := s.servers[i]
		eg.Go(func() error {
			return srv.Start(ctx)
		})
	}
	return eg.Wait()
}

func (s *Processor) Stop(ctx context.Context) error {
	var eg *errgroup.Group
	eg, ctx = errgroup.WithContext(ctx)
	for i := range s.servers {
		srv := s.servers[i]
		eg.Go(func() error {
			return srv.Stop(ctx)
		})
	}
	return eg.Wait()
}

func newSubscribers(servers ...*kafka.Server) *Processor {
	return &Processor{
		servers: servers,
	}
}

type PooledHandler struct {
	pool         *ants.Pool
	handler      broker.Handler
	logger       *log.Helper
	panicOnError bool
}

func NewPooledHandler(handler broker.Handler, poolSize int, logger log.Logger, panicOnError bool) *PooledHandler {
	pool, err := ants.NewPool(poolSize)
	if err != nil {
		log.Fatal(err)
	}
	return &PooledHandler{
		pool:         pool,
		handler:      handler,
		logger:       log.NewHelper(logger),
		panicOnError: panicOnError,
	}
}

func (h *PooledHandler) HandlerFunc() broker.Handler {
	return func(ctx context.Context, evt broker.Event) error {
		return h.pool.Submit(func() {
			if err := h.handler(ctx, evt); err != nil {
				if h.panicOnError {
					h.logger.Fatal("pooled handler handle message error: ", err.Error())
				} else {
					h.logger.Error("pooled handler handle message error: ", err.Error())
				}
			}
		})
	}
}

func NewProcessor(cs map[string]*sharedconf.Kafka, routes map[string]broker.Handler, logger log.Logger) *Processor {
	servers := []*kafka.Server{}
	for k, h := range routes {
		c, ok := cs[k]
		if !ok {
			log.Errorf("no config for kafka `%s`", k)
			continue
		}
		// 有消费配置才配置监听
		if c.Consumer != nil {
			s := newKafkaServer(c, h, logger)
			servers = append(servers, s)
		}
	}
	return newSubscribers(servers...)
}

func newKafkaServer(cb *sharedconf.Kafka, handler broker.Handler, logger log.Logger) *kafka.Server {
	if cb.Consumer != nil && cb.Consumer.WorkerNum > 0 {
		return subscribe(cb, NewPooledHandler(handler, int(cb.Consumer.WorkerNum), logger, false).HandlerFunc())
	} else {
		return subscribe(cb, handler)
	}
}

func subscribe(c *sharedconf.Kafka, handler broker.Handler) *kafka.Server {
	b := kafka.NewServer(
		kafka.WithAddress(c.Servers),
		kafka.WithCodec("json"),
	)
	group := c.GetConsumer().GetGroupId()
	for _, t := range c.GetTopics() {
		log.Infof("listen topic: %s", t)
		if err := b.RegisterSubscriber(context.Background(), t, group, false, handler, nil); err != nil {
			log.Fatalf("create buried point kafka broker failed: %+v", err)
		}
	}
	log.Infof("[kafka] listen addrs=%s topics=%s group=%s", c.GetServers(), c.GetTopics(), c.Consumer.GetGroupId())
	return b
}
