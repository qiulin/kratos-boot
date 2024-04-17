package discovery

import (
	"github.com/go-kratos/kratos/contrib/registry/nacos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/qiulin/kratos-boot/sharedconf"
)

type Factory struct {
	c     *sharedconf.Discovery
	nacos *nacos.Registry
}

func (f *Factory) Registrar() (registry.Registrar, bool) {
	return f.nacos, f.nacos != nil
}

func (f *Factory) Discovery() (registry.Discovery, bool) {
	return f.nacos, f.nacos != nil
}

func NewFactory(c config.Config) (*Factory, error) {
	sc := &sharedconf.Discovery{}
	v := c.Value("discovery")
	if err := v.Scan(sc); err != nil {
		return &Factory{}, nil
	}
	if sc.Nacos == nil {
		return &Factory{c: sc}, nil
	}
	nc, err := NewNacosClient(sc.Nacos)
	if err != nil {
		return nil, err
	}
	nr := NewNacosRegistry(nc, sc.Nacos)
	return &Factory{c: sc, nacos: nr}, nil
}

func NewNacosRegistry(nc naming_client.INamingClient, c *sharedconf.Discovery_Nacos) *nacos.Registry {
	var opts []nacos.Option
	group := c.Group
	if group == "" {
		group = "DEFAULT"
	}
	opts = append(opts, nacos.WithGroup(group))
	if c.Prefix != "" {
		opts = append(opts, nacos.WithPrefix(c.Prefix))
	}
	if c.Cluster != "" {
		opts = append(opts, nacos.WithCluster(c.Cluster))
	}
	return nacos.New(nc, opts...)
}

func NewNacosClient(nc *sharedconf.Discovery_Nacos) (naming_client.INamingClient, error) {

	var sc []constant.ServerConfig
	for _, c := range nc.Addrs {
		sc = append(sc, *constant.NewServerConfig(c.Ip, uint64(c.Port)))
	}

	cc := constant.ClientConfig{
		NamespaceId:         nc.Namespace,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              nc.LogDir,
		CacheDir:            nc.CacheDir,
		LogLevel:            nc.LogLevel,
	}
	// a more graceful way to create naming client
	return clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
}
