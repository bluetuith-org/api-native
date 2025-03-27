//go:build !linux

package commands

import (
	"strconv"
	"time"

	"github.com/bluetuith-org/api-native/api/appfeatures"
	"github.com/bluetuith-org/api-native/api/bluetooth"
	"github.com/bluetuith-org/api-native/api/errorkinds"
	"github.com/bluetuith-org/api-native/shim/internal/serde"
	"github.com/google/uuid"
)

// Session commands.
func GetFeatureFlags() *Command[appfeatures.Features] {
	return &Command[appfeatures.Features]{cmd: "rpc feature-flags"}
}
func GetAdapters() *Command[[]bluetooth.AdapterData] {
	return &Command[[]bluetooth.AdapterData]{cmd: "adapter list"}
}
func AuthenticationReply(id int, input string) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "rpc auth"}).WithArguments(func(am ArgumentMap) {
		am[OperationIdArgument] = strconv.FormatInt(int64(id), 10)
		am[ResponseArgument] = input
	})
}

// Adapter commands.
func AdapterProperties(Address bluetooth.MacAddress) *Command[bluetooth.AdapterData] {
	return (&Command[bluetooth.AdapterData]{cmd: "adapter properties"}).WithArgument(AddressArgument, Address.String())
}
func GetPairedDevices(Address bluetooth.MacAddress) *Command[[]bluetooth.DeviceData] {
	return (&Command[[]bluetooth.DeviceData]{cmd: "adapter get-paired-devices"}).WithArgument(AddressArgument, Address.String())
}
func SetPairableState(Address bluetooth.MacAddress, State bool) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "adapter set-pairable-state"}).WithArguments(func(am ArgumentMap) {
		am[AddressArgument] = Address.String()
		am[StateArgument] = StateArgumentValue(State)
	})
}
func SetDiscoverableState(Address bluetooth.MacAddress, State bool) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "adapter set-discoverable-state"}).WithArguments(func(am ArgumentMap) {
		am[AddressArgument] = Address.String()
		am[StateArgument] = StateArgumentValue(State)
	})
}
func SetPoweredState(Address bluetooth.MacAddress, State bool) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "adapter set-powered-state"}).WithArguments(func(am ArgumentMap) {
		am[AddressArgument] = Address.String()
		am[StateArgument] = StateArgumentValue(State)
	})
}
func StartDiscovery(Address bluetooth.MacAddress) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "adapter discovery start"}).WithArgument(AddressArgument, Address.String())
}
func StopDiscovery(Address bluetooth.MacAddress) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "adapter discovery stop"}).WithArgument(AddressArgument, Address.String())
}

// Device commands.
func DeviceProperties(Address bluetooth.MacAddress) *Command[bluetooth.DeviceData] {
	return (&Command[bluetooth.DeviceData]{cmd: "device properties"}).WithArgument(AddressArgument, Address.String())
}
func Pair(Address bluetooth.MacAddress) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "device pair"}).WithArgument(AddressArgument, Address.String())
}
func CancelPairing(Address bluetooth.MacAddress) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "device pair cancel"}).WithArgument(AddressArgument, Address.String())
}
func Connect(Address bluetooth.MacAddress) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "device connect"}).WithArgument(AddressArgument, Address.String())
}
func Disconnect(Address bluetooth.MacAddress) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "device disconnect"}).WithArgument(AddressArgument, Address.String())
}
func ConnectProfile(Address bluetooth.MacAddress, Profile uuid.UUID) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "device connect profile"}).WithArguments(func(am ArgumentMap) {
		am[AddressArgument] = Address.String()
		am[ProfileArgument] = Profile.String()
	})
}
func DisconnectProfile(Address bluetooth.MacAddress, Profile uuid.UUID) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "device disconnect profile"}).WithArguments(func(am ArgumentMap) {
		am[AddressArgument] = Address.String()
		am[ProfileArgument] = Profile.String()
	})
}
func Remove(Address bluetooth.MacAddress) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "device remove"}).WithArgument(AddressArgument, Address.String())
}

// Obex commands.
func CreateSession(Address bluetooth.MacAddress) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "device opp start-session"}).WithArgument(AddressArgument, Address.String())
}
func RemoveSession() *Command[NoResult] {
	return (&Command[NoResult]{cmd: "device opp stop-session"})
}
func SendFile(File string) *Command[bluetooth.FileTransferData] {
	return (&Command[bluetooth.FileTransferData]{cmd: "device opp send-file"}).WithArgument(FileArgument, File)
}
func CancelTransfer(Address bluetooth.MacAddress) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "device opp cancel-transfer"}).WithArgument(AddressArgument, Address.String())
}
func SuspendTransfer(Address bluetooth.MacAddress) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "device opp suspend-transfer"}).WithArgument(AddressArgument, Address.String())
}
func ResumeTransfer(Address bluetooth.MacAddress) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "device opp resume-transfer"}).WithArgument(AddressArgument, Address.String())
}

func (c *Command[T]) ExecuteWith(fn ExecuteFunc, timeoutSeconds ...int) (T, error) {
	var result T

	var timeout = time.Duration(10)
	if timeoutSeconds != nil {
		timeout = time.Duration(timeoutSeconds[0])
	}

	responseChan, commandErr := fn(c.Slice())
	if commandErr != nil {
		return result, commandErr
	}

	commandErr = errorkinds.ErrSessionStop

	select {
	case response, ok := <-responseChan:
		if !ok {
			break
		}

		if response.Status == "error" {
			return result, response.Error
		}

		if response.Status == "ok" {
			reply := make(map[string]T, 1)
			if err := serde.UnmarshalJson(response.Data, &reply); err != nil {
				return result, err
			}

			for _, mv := range reply {
				result = mv
			}

			commandErr = nil
		}

	case <-time.After(timeout * time.Second):
		commandErr = errorkinds.ErrMethodTimeout
	}

	return result, commandErr
}
