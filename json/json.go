package jsonx

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

func init() {
	extra.RegisterFuzzyDecoders()
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func MustMarshal(v interface{}) []byte {
	bs, _ := Marshal(v)
	return bs
}

func MustMarshalString(v interface{}) string {
	bs := MustMarshal(v)
	return string(bs)
}
