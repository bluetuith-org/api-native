//go:build !linux

package commands

import (
	"fmt"

	"github.com/bluetuith-org/api-native/api/appfeatures"
	"github.com/bluetuith-org/api-native/api/bluetooth"
	"github.com/bluetuith-org/api-native/shim/internal/serde"
	"github.com/google/uuid"
)

// Session commands.
func StartRpcServer(Socket string) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "rpc start-session"}).WithArgument(SocketArgument, Socket)
}
func StopRpcServer() *Command[NoResult] {
	return &Command[NoResult]{cmd: "rpc stop-session"}
}
func GetFeatureFlags() *Command[appfeatures.Features] {
	return &Command[appfeatures.Features]{cmd: "rpc feature-flags"}
}
func GetAdapters() *Command[[]bluetooth.AdapterData] {
	return &Command[[]bluetooth.AdapterData]{cmd: "adapter list"}
}

// Adapter commands.
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
	return (&Command[NoResult]{cmd: "adapter start-discovery"}).WithArgument(AddressArgument, Address.String())
}
func StopDiscovery(Address bluetooth.MacAddress) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "adapter stop-discovery"}).WithArgument(AddressArgument, Address.String())
}

// Device commands.
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

func (c *Command[T]) ExecuteWith(fn ExecuteFunc) (T, CommandMetadata, error) {
	var result T
	var metadata CommandMetadata

	replyChan, err := fn(c.Slice()...)
	if err != nil {
		return result, metadata, err
	}

	for rawReply := range replyChan {
		reply := CommandReply[T, Command[T]]{}
		metadata = rawReply.CommandMetadata
		if err := serde.UnmarshalJson(rawReply.RawData, &reply); err != nil {
			fmt.Println(err)
			return result, metadata, err
		}

		if reply.Status == "error" {
			return result, metadata, reply.Error
		}

		if reply.Status == "ok" && reply.Data != nil {
			for _, mv := range reply.Data {
				result = mv
			}
		}
	}

	return result, metadata, nil
}
