package bluetooth

import (
	ac "github.com/bluetuith-org/api-native/api/appfeatures"
	"github.com/bluetuith-org/api-native/api/config"
)

// Session describes a Bluetooth application session.
type Session interface {
	// Start attempts to initialize a session with the system's Bluetooth daemon or service.
	// Upon complete initialization, it returns the session descriptor, and capabilities of
	// the application.
	Start(authHandler SessionAuthorizer, cfg config.Configuration) (ac.FeatureSet, error)

	// Stop attempts to stop a session with the system's Bluetooth daemon or service.
	Stop() error

	// Adapters returns a list of known adapters.
	Adapters() []AdapterData

	// Adapter returns a function call interface to invoke adapter related functions.
	Adapter(adapterAddress MacAddress) Adapter

	// Device returns a function call interface to invoke device related functions.
	Device(deviceAddress MacAddress) Device

	// Obex returns a function call interface to invoke obex related functions.
	Obex(deviceAddress MacAddress) Obex

	// Network returns a function call interface to invoke network related functions.
	Network(deviceAddress MacAddress) Network

	// MediaPlayer returns a function call interface to invoke media player/control
	// related functions on a device.
	MediaPlayer(deviceAddress MacAddress) MediaPlayer
}
