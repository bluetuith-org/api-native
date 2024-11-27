//go:build linux

package dbushelper

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/ugorji/go/codec"
)

// VariantExt represents a go-codec extension to parse DBus variant values.
type VariantExt struct{}

// Resolver holds an encoder and decoder.
type Resolver struct {
	check bool

	encoder *codec.Encoder
	decoder *codec.Decoder
	data    []byte

	lock sync.Mutex
}

var resolver Resolver

// ConvertExt converts a variant struct into an encodable value.
// Note: v is a pointer iff the registered extension type is a struct or array kind.
func (v VariantExt) ConvertExt(variant interface{}) interface{} {
	return variant.(*dbus.Variant).Value()
}

// UpdateExt decodes/updates an encoded value (src) to a new variant (dst).
// Note: dst is always a pointer kind to the registered extension type.
func (v VariantExt) UpdateExt(dst, src interface{}) {
	dst.(dbus.Variant).Store(src)
}

// DecodeVariantMap decodes a map of variants into the provided data.
// Note that, for types "MacAddress" and "uuid.UUID", custom TextMarshaler
// and TextUnmarshaler interfaces have been defined.
func DecodeVariantMap(
	variants map[string]dbus.Variant, data interface{},
	checkProps ...string,
) error {
	resolver.lock.Lock()
	defer resolver.lock.Unlock()

	if !resolver.check {
		handle := codec.JsonHandle{}
		handle.SetInterfaceExt(reflect.TypeOf(dbus.Variant{}), 1, VariantExt{})
		resolver.encoder = codec.NewEncoderBytes(&resolver.data, &handle)
		resolver.decoder = codec.NewDecoderBytes(resolver.data, &handle)

		resolver.check = true
	}

	for key, value := range variants {
		for _, prop := range checkProps {
			if prop == key && value.Signature().Empty() {
				return fmt.Errorf("No signature found for property '%s'", prop)
			}
		}
	}

	resolver.encoder.ResetBytes(&resolver.data)

	if err := resolver.encoder.Encode(variants); err != nil {
		return err
	}

	resolver.decoder.ResetBytes(resolver.data)

	return resolver.decoder.Decode(data)
}
