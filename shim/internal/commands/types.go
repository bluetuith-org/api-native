//go:build !linux

package commands

import (
	"strings"
	"time"

	"github.com/ugorji/go/codec"
)

const CommandReplyTimeout = 10 * time.Second

type ExecuteFunc func(params []string) (chan CommandResponse, error)
type ArgumentMap = map[Argument]string
type NoResult = struct{}

type OperationID uint32
type RequestID int64

// T is the return value type of the command.
// If T is of type NoResult, it means the command only returns errors, and no other values.
type Command[T any] struct {
	cmd    string
	argmap ArgumentMap
}

type CommandResponse struct {
	Status string `json:"status"`

	OperationId OperationID  `json:"operation_id,omitempty"`
	RequestId   RequestID    `json:"request_id,omitempty"`
	Error       CommandError `json:"error"`
	Data        codec.Raw    `json:"data"`
}

type CommandError struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Metadata    map[string]string `json:"metadata"`
}

func (c CommandError) Error() string {
	sb := strings.Builder{}

	sb.WriteString(c.Name)
	sb.WriteString(": ")
	if c.Description == "" {
		sb.WriteString("No information is provided for this error")
	} else {
		sb.WriteString(c.Description)
	}
	sb.WriteString(". ")

	count := 0
	length := len(c.Metadata)
	if length == 0 {
		goto Print
	}

	sb.WriteString("(")
	for _, v := range c.Metadata {
		count++
		sb.WriteString(v)

		if count < length {
			sb.WriteString(", ")
		}
	}
	sb.WriteString(")")

Print:
	return sb.String()
}

func (c *Command[T]) String() string {
	sb := strings.Builder{}
	sb.Grow(len(c.cmd) + (len(c.argmap) * 2))

	sb.WriteString(c.cmd)
	for param, value := range c.argmap {
		sb.WriteString(" ")
		sb.WriteString(string(param))
		sb.WriteString(" ")
		sb.WriteString(value)
	}

	return sb.String()
}

func (c *Command[T]) Slice() []string {
	return strings.Split(c.String(), " ")
}

func (c *Command[T]) WithArgument(arg Argument, value string) *Command[T] {
	if c.argmap == nil {
		c.argmap = make(ArgumentMap)
	}

	c.argmap[arg] = value

	return c
}

func (c *Command[T]) WithArguments(fn func(ArgumentMap)) *Command[T] {
	if c.argmap == nil {
		c.argmap = make(ArgumentMap)
	}

	fn(c.argmap)

	return c
}
