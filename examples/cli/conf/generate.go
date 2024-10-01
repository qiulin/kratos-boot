package conf

//go:generate protoc -I ../../../ -I . --go_out=paths=source_relative:. ./conf.proto
