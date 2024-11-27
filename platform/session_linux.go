//go:build linux

package platform

import (
	"github.com/bluetuith-org/api-native/api/bluetooth"
	"github.com/bluetuith-org/api-native/linux"
)

// Session returns a platform-specific session handler.
func Session() (bluetooth.Session, PlatformInfo) {
	return &linux.BluezSession{}, NewPlatformInfo(BluezStack)
}
