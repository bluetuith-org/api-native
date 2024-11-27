package bluetooth

// Device describes a function call interface to invoke device related functions.
type Device interface {
	// Pair will attempt to pair a bluetooth device that is in pairing mode.
	Pair() error

	// CancelPairing will cancel a pairing attempt.
	CancelPairing() error

	// Connect will attempt to connect an already paired bluetooth device
	// to an adapter.
	Connect() error

	// Disconnect will disconnect the bluetooth device from the adapter.
	Disconnect() error

	// ConnectProfile will attempt to connect an already paired bluetooth device
	// to an adapter, using a specific Bluetooth profile UUID .
	ConnectProfile(profileUUID string) error

	// DisconnectProfile will attempt to disconnect an already paired bluetooth device
	// to an adapter, using a specific Bluetooth profile UUID .
	DisconnectProfile(profileUUID string) error

	// Remove removes a device from its associated adapter.
	Remove() error

	// Properties returns all the properties of the device.
	Properties() (DeviceData, error)
}

// DeviceData holds the static bluetooth device information installed for a system.
type DeviceData struct {
	// Name holds the name of the device.
	Name string

	// Class holds the device type class specifier.
	Class uint32

	// Type holds the type name of the device.
	// For example, type of the device can be "Phone", "Headset" etc.
	Type string

	// Alias holds the optional or user-assigned name for the adapter.
	// Usually valid for Linux systems, may be empty or equate to "Name"
	// for other systems.
	Alias string

	// LegacyPairing indicates whether the device only supports the pre-2.1 pairing mechanism.
	// This property is useful during device discovery to anticipate whether
	// legacy or simple pairing will occur if pairing is initiated.
	LegacyPairing bool

	DeviceEventData
}

// DeviceEventData holds the dynamic (variable) bluetooth device information.
// This is primarily used to send device event related data.
type DeviceEventData struct {
	// Address holds the bluetooth mac address of the device.
	Address MacAddress

	// AssociatedAdapter holds the bluetooth mac address of the adapter
	// the device is associated with.
	AssociatedAdapter MacAddress

	// Paired indicates if the device is paired.
	Paired bool

	// Connected indicates if the device is connected.
	Connected bool

	// Trusted indicates if the device is marked as trusted.
	// Valid only on Linux systems, will equate to "true"
	// on other systems if the device is paired.
	Trusted bool

	// Blocked indicates if the device is marked as blocked.
	// Valid only on Linux systems, will equate to "false"
	// on other systems.
	Blocked bool

	// Bonded indicates if the device is bonded.
	Bonded bool

	// RSSI indicates the signal strength of the device.
	RSSI int16

	// BatteryPercentage holds the battery percentage of the device.
	// Valid only on Linux systems, may not hold valid information
	// on other systems.
	BatteryPercentage int

	// UUIDs holds the device-supported Bluetooth profile UUIDs.
	UUIDs []string
}

// DeviceTypeFromClass parses the device class and returns its type.
//
//gocyclo:ignore
func DeviceTypeFromClass(class uint32) string {
	/*
		Adapted from:
		https://gitlab.freedesktop.org/upower/upower/-/blob/master/src/linux/up-device-bluez.c#L64
	*/
	switch (class & 0x1f00) >> 8 {
	case 0x01:
		return "Computer"

	case 0x02:
		switch (class & 0xfc) >> 2 {
		case 0x01, 0x02, 0x03, 0x05:
			return "Phone"

		case 0x04:
			return "Modem"
		}

	case 0x03:
		return "Network"

	case 0x04:
		switch (class & 0xfc) >> 2 {
		case 0x01, 0x02:
			return "Headset"

		case 0x05:
			return "Speakers"

		case 0x06:
			return "Headphones"

		case 0x0b, 0x0c, 0x0d:
			return "Video"

		default:
			return "Audio device"
		}

	case 0x05:
		switch (class & 0xc0) >> 6 {
		case 0x00:
			switch (class & 0x1e) >> 2 {
			case 0x01, 0x02:
				return "Gaming input"

			case 0x03:
				return "Remote control"
			}

		case 0x01:
			return "Keyboard"

		case 0x02:
			switch (class & 0x1e) >> 2 {
			case 0x05:
				return "Tablet"

			default:
				return "Mouse"
			}
		}

	case 0x06:
		if (class & 0x80) > 0 {
			return "Printer"
		}

		if (class & 0x40) > 0 {
			return "Scanner"
		}

		if (class & 0x20) > 0 {
			return "Camera"
		}

		if (class & 0x10) > 0 {
			return "Monitor"
		}

	case 0x07:
		return "Wearable"

	case 0x08:
		return "Toy"
	}

	return "Unknown"
}
