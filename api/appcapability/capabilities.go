package appcapability

import (
	"fmt"
	"strings"
)

// AppCapabilities describes the capabilities of an application.
type AppCapabilities uint8

// The different kinds of individual capabilities.
const (
	CapabilityNone       AppCapabilities = 0 // The zero value for this type.
	CapabilityConnection                 = 1 << iota
	CapabilityPairing
	CapabilitySendFile
	CapabilityReceiveFile
	CapabilityNetwork
)

// Error describes an error which occurred while attempting
// to enable support for the specified capability.
type Error struct {
	Capabilities AppCapabilities
	Err          error
}

// Errors holds a list of capability based errors.
type Errors struct {
	errors map[AppCapabilities]Error
}

// Collection holds all supported capabilities and capability related errors.
type Collection struct {
	Supported AppCapabilities
	Errors    Errors
}

// capabilityMap holds a list of descriptions for each capability.
var capabilityMap = map[AppCapabilities]string{
	CapabilityConnection:  "Bluetooth Connection",
	CapabilityPairing:     "Bluetooth Pairing",
	CapabilitySendFile:    "OBEX Send Files",
	CapabilityReceiveFile: "OBEX Receive Files",
	CapabilityNetwork:     "PANU/DUN Network Connection",
}

// NewCollection returns a new Collection (of capabilities).
func NewCollection(capabilities AppCapabilities, errors Errors) Collection {
	return Collection{
		Supported: capabilities,
		Errors:    errors,
	}
}

// NewError returns a capability-based Error.
func NewError(c AppCapabilities, err error) *Error {
	return &Error{
		Capabilities: c,
		Err:          err,
	}
}

// NilCollection returns an empty collection of capabilities.
func NilCollection() Collection {
	return Collection{}
}

// String converts a set of capabilities to a comma-separated string of
// their respective descriptions.
func (c AppCapabilities) String() string {
	s := make([]string, 0, len(capabilityMap))

	for capability, title := range capabilityMap {
		if c&capability != 0 {
			s = append(s, title)
		}
	}

	return strings.Join(s, ", ")
}

// Slice returns a slice of individual app capabilities.
func (c AppCapabilities) Slice() []AppCapabilities {
	s := make([]AppCapabilities, 0, len(capabilityMap))

	for capability := range capabilityMap {
		if c&capability != 0 {
			s = append(s, capability)
		}
	}

	return s
}

// Append appends a single capability error to the capability error list.
func (c *Errors) Append(e *Error) {
	if c.errors == nil {
		c.errors = make(map[AppCapabilities]Error)
	}

	c.errors[e.Capabilities] = *e
}

// Exists checks and returns all capability based errors.
func (c *Errors) Exists() (map[AppCapabilities]Error, bool) {
	return c.errors, c.errors != nil
}

// Error returns a text representation of the capability error.
func (c *Error) Error() string {
	return fmt.Sprintf(
		"Capabilities '%s' cannot be activated: %s",
		c.Capabilities, c.Err,
	)
}
