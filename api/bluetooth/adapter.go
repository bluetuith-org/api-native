package bluetooth

import "github.com/google/uuid"

// Adapter describes a function call interface to invoke adapter related functions.
type Adapter interface {
	// StartDiscovery will put the adapter into "discovering" mode, which means
	// the bluetooth device will be able to discover other bluetooth devices
	// that are in pairing mode.
	StartDiscovery() error

	// StopDiscovery will stop the  "discovering" mode, which means the bluetooth device will
	// no longer be able to discover other bluetooth devices that are in pairing mode.
	StopDiscovery() error

	// SetPoweredState sets the powered state of the adapter.
	SetPoweredState(enable bool) error

	// SetDiscoverableState sets the discoverable state of the adapter.
	SetDiscoverableState(enable bool) error

	// SetPairableState sets the pairable state of the adapter.
	SetPairableState(enable bool) error

	// Properties returns all the properties of the adapter.
	Properties() (AdapterData, error)

	// Devices returns all the devices associated with the adapter
	Devices() ([]DeviceData, error)
}

// AdapterData holds the static bluetooth adapter information installed for a system.
type AdapterData struct {
	// Name holds the system-assigned name of the adapter.
	Name string

	// Alias holds the optional or user-assigned name for the adapter.
	// Usually valid for Linux systems, may be empty or equate to "Name"
	// for other systems.
	Alias string

	// UniqueName holds a unique name for the adapter.
	// For example, on Linux it can be "hci0".
	// For other systems, it can equate to (Adapter).Name.
	UniqueName string

	// UUIDs holds all the supported profile uuids.
	UUIDs uuid.UUIDs

	AdapterEventData
}

// AdapterEventData holds the dynamic (variable) bluetooth adapter information.
// This is primarily used to send adapter event related data.
type AdapterEventData struct {
	// Address holds the bluetooth mac address of the adapter.
	Address MacAddress

	// Discoverable indicates whether the adapter is discoverable by other devices.
	Discoverable bool

	// Pairable indicates whether the adapter is pairable with other devices.
	Pairable bool

	// Powered indicates whether the adapter is powered on or off.
	Powered bool

	// Discovering indicates whether the adapter is discovering devices.
	Discovering bool
}
