package console

import (
	"context"
	"emperror.dev/emperror"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/google/wire"
	"github.com/qiulin/kratos-boot/boot"
	"github.com/spf13/cobra"
	"sync"
	"time"
)

type ExampleAware interface {
	Example() string
}

type LongDescribeAware interface {
	LongDescribe() string
}

type Command interface {
	Command() string
	Describe() string
	Run(ctx context.Context, args []string) error
}

type Option struct {
	Name          string
	CloseWaitTime time.Duration
}

func (o *Option) ensureDefaults() {
	if o.CloseWaitTime == 0 {
		o.CloseWaitTime = 1 * time.Second
	}
	if o.Name == "" {
		o.Name = "cli boot"
	}
}

type CommandLine struct {
	app       *kratos.App
	o         *Option
	rootCmd   *cobra.Command
	ctx       context.Context
	cancelCtx context.CancelFunc
}

func NewCommandLine(o *Option, app *kratos.App, cmds []Command, wg *sync.WaitGroup) *CommandLine {
	o.ensureDefaults()
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	rootCmd := &cobra.Command{
		Use: o.Name,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			go func() {
				emperror.Panic(app.Run())
			}()
			wg.Wait()
			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return app.Stop()
		},
	}
	for _, c := range cmds {
		cmd := &cobra.Command{
			Use:   c.Command(),
			Short: c.Describe(),
			Long:  c.Describe(),
			RunE: func(cmd *cobra.Command, args []string) error {
				err := c.Run(ctx, args)
				time.Sleep(o.CloseWaitTime)
				return err
			},
		}

		if v, ok := c.(ExampleAware); ok {
			cmd.Example = v.Example()
		}

		if v, ok := c.(LongDescribeAware); ok {
			cmd.Long = v.LongDescribe()
		}

		rootCmd.AddCommand(cmd)
	}

	return &CommandLine{
		app:       app,
		o:         o,
		rootCmd:   rootCmd,
		ctx:       ctx,
		cancelCtx: cancel,
	}
}

func (cli *CommandLine) Execute() error {
	emperror.Panic(cli.rootCmd.Execute())
	defer cli.cancelCtx()
	return nil
}

func NewWaitGroup() *sync.WaitGroup {
	return &sync.WaitGroup{}
}

func NewApp(b *boot.Bootstrap, servers []transport.Server, wg *sync.WaitGroup) *kratos.App {
	opts := []kratos.Option{
		kratos.ID(b.Option().ServiceId),
		kratos.Name(b.Option().ServiceName),
		kratos.Version(b.Option().Version),
		kratos.Metadata(b.Option().ServiceMetadata),
		kratos.Logger(b.KLogger()),
		kratos.Server(
			servers...,
		),
		kratos.AfterStart(func(ctx context.Context) error {
			wg.Done()
			return nil
		}),
	}

	return kratos.New(
		opts...,
	)
}

type WireFunc func(bootstrap *boot.Bootstrap) (*CommandLine, func(), error)

func Execute(b *boot.Bootstrap, wireFunc WireFunc) {
	cli, cleanup, err := wireFunc(b)
	emperror.Panic(err)
	defer cleanup()
	emperror.Panic(cli.Execute())
}

var ProviderSet = wire.NewSet(boot.ProviderSetBase, NewApp, NewCommandLine, NewWaitGroup)
