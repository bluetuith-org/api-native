//go:build !linux

package serde

import (
	"sync"

	"github.com/ugorji/go/codec"
)

// resolver holds an encoder and decoder.
type resolver struct {
	check bool

	jsonEncoder *codec.Encoder
	jsonDecoder *codec.Decoder
	jsonHandle  codec.JsonHandle

	jsonData []byte

	jsonMu sync.Mutex
}

var gendecoder resolver

func init() {
	if !gendecoder.check {
		gendecoder.jsonHandle = codec.JsonHandle{}
		gendecoder.jsonHandle.ErrorIfNoField = true
		gendecoder.jsonHandle.ErrorIfNoArrayExpand = true
		gendecoder.jsonHandle.TypeInfos = codec.NewTypeInfos([]string{"json"})
		gendecoder.jsonEncoder = codec.NewEncoderBytes(&gendecoder.jsonData, &gendecoder.jsonHandle)
		gendecoder.jsonDecoder = codec.NewDecoderBytes(gendecoder.jsonData, &gendecoder.jsonHandle)

		gendecoder.jsonData = make([]byte, 0, 4096)
		gendecoder.check = true
	}
}

func MarshalJson[T any](v T) ([]byte, error) {
	gendecoder.jsonMu.Lock()
	defer gendecoder.jsonMu.Unlock()

	gendecoder.jsonEncoder.ResetBytes(&gendecoder.jsonData)

	return gendecoder.jsonData, gendecoder.jsonEncoder.Encode(v)
}

func UnmarshalJson[T any](data []byte, marshalTo T) error {
	gendecoder.jsonMu.Lock()
	defer gendecoder.jsonMu.Unlock()

	gendecoder.jsonDecoder.ResetBytes(data)

	return gendecoder.jsonDecoder.Decode(marshalTo)
}
