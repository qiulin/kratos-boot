package main

import (
	"context"
	"github.com/qiulin/kratos-boot/console"
	"log/slog"
	"time"
)

var _ console.Command = new(HelloCommand)

type HelloCommand struct {
}

func (cmd *HelloCommand) Command() string {
	return "hello"
}

func (cmd *HelloCommand) Describe() string {
	return "hello world"
}

func (cmd *HelloCommand) Run(ctx context.Context, args []string) error {
	_args := []any{}
	for _, arg := range args {
		_args = append(_args, arg)
	}
	slog.InfoContext(ctx, "hello world", _args...)
	return nil
}

func NewHelloCommand() *HelloCommand {
	return &HelloCommand{}
}

func NewCommands(hello *HelloCommand, ts *TimestampCommand) []console.Command {
	return []console.Command{hello, ts}
}

var _ console.Command = new(TimestampCommand)

func NewTimestampCommand() *TimestampCommand {
	return &TimestampCommand{}
}

type TimestampCommand struct{}

func (cli *TimestampCommand) Command() string {
	return "timestamp"
}

func (cli *TimestampCommand) Describe() string {
	return "unix timestamp"
}

func (cli *TimestampCommand) Run(ctx context.Context, args []string) error {
	var t time.Time
	var offset string
	if len(args) > 0 {
		offset = args[0]
		d, err := time.ParseDuration(offset)
		if err != nil {
			return err
		}
		t = time.Now().Add(d)
	} else {
		t = time.Now()
	}
	slog.InfoContext(ctx, "timestamp", "unix timestamp", t.Unix(), "offset", offset)
	return nil
}
