//go:build windows

package platform

import (
	"github.com/bluetuith-org/api-native/api/bluetooth"
	"github.com/bluetuith-org/api-native/shim"
)

// Session returns a platform-specific session handler.
func Session() (bluetooth.Session, PlatformInfo) {
	return &shim.ShimSession{}, NewPlatformInfo(MicrosoftBluetoothStack)
}
