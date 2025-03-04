package platform

import "runtime"

type BluetoothStack string

const (
	BluezStack              BluetoothStack = "BlueZ (DBus)"
	MicrosoftBluetoothStack BluetoothStack = "Microsoft"
)

// PlatformInfo describes platform-specific information.
type PlatformInfo struct {
	OS    string         `json:"os,omitempty"`
	Stack BluetoothStack `json:"bluetooth_stack,omitempty"`
}

// NewPlatformInfo returns a new PlatformInfo.
func NewPlatformInfo(stack BluetoothStack) PlatformInfo {
	return PlatformInfo{
		OS:    runtime.GOOS + " (" + runtime.GOARCH + ")",
		Stack: stack,
	}
}

// String converts a BluetoothStack to a string.
func (b BluetoothStack) String() string {
	return string(b)
}
