package bluetooth

import (
	"context"

	ac "github.com/bluetuith-org/api-native/api/appcapability"
	"github.com/google/uuid"
)

// Session describes a Bluetooth application session.
type Session interface {
	// Start attempts to initialize a session with the system's Bluetooth daemon or service.
	// Upon complete initialization, it returns the session descriptor, and capabilities of
	// the application.
	Start(authHandler SessionAuthHandler, path ...string) (ac.Collection, error)

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
}

// SessionAuthHandler describes an authentication interface for authorizing session functions.
type SessionAuthHandler interface {
	ReceiveFileAuthHandler

	DisplayPinCode(ctx context.Context, address MacAddress, pincode string) error
	DisplayPasskey(ctx context.Context, address MacAddress, passkey uint32, entered uint16) error
	ConfirmPasskey(ctx context.Context, address MacAddress, passkey uint32) error
	AuthorizePairing(ctx context.Context, address MacAddress) error
	AuthorizeService(ctx context.Context, address MacAddress, uuid uuid.UUID) error
}
