//go:build !linux

package serde

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/ugorji/go/codec"
)

// resolver holds an encoder and decoder.
type resolver struct {
	check bool

	jsonEncoder *codec.Encoder
	jsonDecoder *codec.Decoder
	jsonHandle  codec.JsonHandle

	simpleEncoder *codec.Encoder
	simpleDecoder *codec.Decoder
	simpleHandle  codec.SimpleHandle

	jsonData   []byte
	simpleData []byte

	jsonMu   sync.Mutex
	simpleMu sync.Mutex
}

var gendecoder resolver

func init() {
	if !gendecoder.check {
		gendecoder.jsonHandle = codec.JsonHandle{}
		//gendecoder.jsonHandle.ErrorIfNoField = true
		//gendecoder.jsonHandle.ErrorIfNoArrayExpand = true
		gendecoder.jsonHandle.TypeInfos = codec.NewTypeInfos([]string{"json"})
		gendecoder.jsonEncoder = codec.NewEncoderBytes(&gendecoder.jsonData, &gendecoder.jsonHandle)
		gendecoder.jsonDecoder = codec.NewDecoderBytes(gendecoder.jsonData, &gendecoder.jsonHandle)

		gendecoder.simpleHandle = codec.SimpleHandle{}
		gendecoder.simpleHandle.ErrorIfNoField = true
		gendecoder.simpleHandle.ErrorIfNoArrayExpand = true
		gendecoder.simpleEncoder = codec.NewEncoderBytes(&gendecoder.simpleData, &gendecoder.simpleHandle)
		gendecoder.simpleDecoder = codec.NewDecoderBytes(gendecoder.simpleData, &gendecoder.simpleHandle)

		gendecoder.jsonData = make([]byte, 0, 4096)
		gendecoder.simpleData = make([]byte, 0, 128)
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

func UnmarshalJsonMap[M any](variants map[string]any, marshalTo M) error {
	gendecoder.jsonMu.Lock()
	defer gendecoder.jsonMu.Unlock()

	gendecoder.jsonEncoder.ResetBytes(&gendecoder.jsonData)

	if err := gendecoder.jsonEncoder.Encode(variants); err != nil {
		return err
	}

	gendecoder.jsonDecoder.ResetBytes(gendecoder.jsonData)

	return gendecoder.jsonDecoder.Decode(marshalTo)
}

func UnmarshalSingleValue[T any](variants map[string]any, marshalTo T) error {
	if isJsonMarshallable[T]() {
		return UnmarshalJsonMap(variants, marshalTo)
	}
	if len(variants) != 1 {
		return fmt.Errorf("the provided map does not represent a single value")
	}

	gendecoder.simpleMu.Lock()
	defer gendecoder.simpleMu.Unlock()

	gendecoder.simpleEncoder.ResetBytes(&gendecoder.simpleData)

	var value any
	for _, v := range variants {
		value = v
	}

	if err := gendecoder.simpleEncoder.Encode(value); err != nil {
		return err
	}

	gendecoder.simpleDecoder.ResetBytes(gendecoder.simpleData)

	return gendecoder.simpleDecoder.Decode(marshalTo)
}

func isJsonMarshallable[T any]() bool {
	t := reflect.TypeFor[T]()

	switch t.Kind() {
	case reflect.Pointer:
		fallthrough

	case reflect.Slice:
		t = t.Elem()
	}

	return t.Kind() == reflect.Struct || t.Kind() == reflect.Map || t.Kind() == reflect.Slice
}
