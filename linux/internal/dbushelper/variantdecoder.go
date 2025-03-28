//go:build linux

package dbushelper

import (
	"fmt"
	"reflect"
	"sync"

	bluetooth "github.com/bluetuith-org/api-native/api/bluetooth"
	sstore "github.com/bluetuith-org/api-native/api/helpers/sessionstore"
	"github.com/godbus/dbus/v5"
	"github.com/ugorji/go/codec"
)

// variantExt represents a go-codec extension to parse DBus variant values.
type variantExt struct{}

// resolver holds an encoder and decoder.
type resolver struct {
	check bool

	encoder *codec.Encoder
	decoder *codec.Decoder
	data    []byte

	lock sync.Mutex
}

var variantDecoder resolver

// ConvertExt converts a variant struct into an encodable value.
// Note: v is a pointer iff the registered extension type is a struct or array kind.
func (v variantExt) ConvertExt(variant interface{}) interface{} {
	return variant.(*dbus.Variant).Value()
}

// UpdateExt decodes/updates an encoded value (src) to a new variant (dst).
// Note: dst is always a pointer kind to the registered extension type.
func (v variantExt) UpdateExt(dst, src interface{}) {
	dst.(dbus.Variant).Store(src)
}

// DecodeVariantMap decodes a map of variants into the provided data.
// Note that, for types "MacAddress" and "uuid.UUID", custom TextMarshaler
// and TextUnmarshaler interfaces have been defined.
func DecodeVariantMap(
	variants map[string]dbus.Variant, data interface{},
	checkProps ...string,
) error {
	variantDecoder.lock.Lock()
	defer variantDecoder.lock.Unlock()

	if !variantDecoder.check {
		handle := codec.JsonHandle{}
		handle.SetInterfaceExt(reflect.TypeOf(dbus.Variant{}), 1, variantExt{})
		variantDecoder.encoder = codec.NewEncoderBytes(&variantDecoder.data, &handle)
		variantDecoder.decoder = codec.NewDecoderBytes(variantDecoder.data, &handle)

		variantDecoder.check = true
	}

	for key, value := range variants {
		for _, prop := range checkProps {
			if prop == key && value.Signature().Empty() {
				return fmt.Errorf("No signature found for property '%s'", prop)
			}
		}
	}

	variantDecoder.encoder.ResetBytes(&variantDecoder.data)

	if err := variantDecoder.encoder.Encode(variants); err != nil {
		return err
	}

	variantDecoder.decoder.ResetBytes(variantDecoder.data)

	return variantDecoder.decoder.Decode(data)
}

// DecodeAdapterFunc returns a function to decode and merge adapter data.
func DecodeAdapterFunc(variants map[string]dbus.Variant) sstore.MergeAdapterDataFunc {
	return func(adapter *bluetooth.AdapterData) error {
		return DecodeVariantMap(variants, adapter)
	}
}

// DecodeDeviceFunc returns a function to decode and merge device data.
func DecodeDeviceFunc(variants map[string]dbus.Variant) sstore.MergeDeviceDataFunc {
	return func(device *bluetooth.DeviceData) error {
		return DecodeVariantMap(variants, device)
	}
}
