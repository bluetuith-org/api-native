//go:build !linux

package shim

import (
	"github.com/bluetuith-org/api-native/api/bluetooth"
	"github.com/bluetuith-org/api-native/api/errorkinds"
)

type network struct {
}

// Connect connects to a specific device according to the provided NetworkType
// and assigns a name to the established connection.
func (n *network) Connect(name string, nt bluetooth.NetworkType) error {
	return errorkinds.ErrNotSupported
}

// Disconnect disconnects from an established connection.
func (n *network) Disconnect() error {
	return errorkinds.ErrNotSupported
}
