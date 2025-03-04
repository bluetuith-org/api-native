package config

import (
	"runtime"
	"time"
)

const (
	// The default timeout duration for authentication requests.
	DefaultAuthTimeout = 10 * time.Second
)

// Configuration describes a general configuration.
type Configuration struct {
	// ShimPath holds the path to the shim executable.
	// Specific to Windows, MacOS and FreeBSD shims.
	ShimPath string

	// SocketPath holds the path to the socket used to interface with the shim.
	SocketPath string

	// AuthTimeout holds the timeout for authentication requests.
	AuthTimeout time.Duration
}

// New returns a new configuration with the default authentication timeout.
func New() Configuration {
	var shimpath = "bluetuith-shim"
	if runtime.GOOS == "windows" {
		shimpath += ".exe"
	}

	return Configuration{
		AuthTimeout: DefaultAuthTimeout,
		ShimPath:    shimpath,
	}
}
