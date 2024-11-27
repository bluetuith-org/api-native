package config

import "time"

const (
	// The default timeout duration for authentication requests.
	DefaultAuthTimeout = 10 * time.Second
)

// Configuration describes a general configuration.
type Configuration struct {
	// ExecutablePath holds the path to the executable.
	// Specific to Windows, MacOS and FreeBSD shims.
	ExecutablePath string

	// AuthTimeout holds the timeout for authentication requests.
	AuthTimeout time.Duration
}

// New returns a new configuration with the default authentication timeout.
func New() Configuration {
	return Configuration{
		AuthTimeout: DefaultAuthTimeout,
	}
}
