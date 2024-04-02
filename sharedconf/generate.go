package sharedconf

//go:generate protoc -I ../protobuf -I . --go_out=paths=source_relative:. ./sharedconf.proto
