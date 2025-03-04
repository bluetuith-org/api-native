package commands

type Argument string

const (
	SocketArgument  Argument = "--socket-path"
	AddressArgument Argument = "--address"
	StateArgument   Argument = "--state"
	ProfileArgument Argument = "--uuid"
	FileArgument    Argument = "--file"
)

func (a Argument) String() string {
	return string(a)
}

func StateArgumentValue(enable bool) string {
	if !enable {
		return "off"
	}

	return "on"
}
