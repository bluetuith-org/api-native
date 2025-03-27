package bluetooth

import (
	"context"
	"strconv"
	"time"

	"github.com/bluetuith-org/api-native/api/errorkinds"
	"github.com/google/uuid"
)

// SessionAuthorizer describes an authentication interface for authorizing session functions.
type SessionAuthorizer interface {
	AuthorizeReceiveFile
	AuthorizeDevicePairing
}

// AuthTimeout describes an authentication timeout duration.
// The context value is created with 'context.WithTimeout()'.
type AuthTimeout struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type AuthEventID string

const (
	AuthEventNone     AuthEventID = "auth-event-none"
	DisplayPinCode    AuthEventID = "display-pincode"
	DisplayPasskey    AuthEventID = "display-passkey"
	ConfirmPasskey    AuthEventID = "confirm-passkey"
	AuthorizePairing  AuthEventID = "authorize-pairing"
	AuthorizeService  AuthEventID = "authorize-service"
	AuthorizeTransfer AuthEventID = "authorize-transfer"
)

type AuthReplyMethod string

const (
	ReplyNone      AuthReplyMethod = "reply-none"
	ReplyYesNo     AuthReplyMethod = "reply-yes-no"
	ReplyWithInput AuthReplyMethod = "reply-with-input"
)

type AuthReply struct {
	ReplyMethod AuthReplyMethod
	Reply       string
}

// AuthEventData describes an authentication event.
type AuthEventData struct {
	AuthID      int             `json:"auth_id,omitempty"`
	EventID     AuthEventID     `json:"auth_event,omitempty"`
	ReplyMethod AuthReplyMethod `json:"auth_reply_method,omitempty"`

	TimeoutMs int        `json:"timeout_ms,omitempty"`
	Address   MacAddress `json:"address,omitempty"`

	Pincode string `json:"pincode,omitempty"`

	Passkey uint32 `json:"passkey,omitempty"`
	Entered uint16 `json:"entered,omitempty"`

	UUID uuid.UUID `json:"uuid,omitempty"`

	FileTransfer FileTransferData `json:"file_transfer,omitempty"`
}

func (a *AuthEventData) CallAuthorizer(authorizer SessionAuthorizer, cb func(authEvent AuthEventData, reply AuthReply, err error)) error {
	if authorizer == nil {
		return errorkinds.ErrMethodCall
	}

	var authfn func() (AuthReply, error)

	switch a.EventID {
	case DisplayPinCode:
		authfn = func() (AuthReply, error) {
			return AuthReply{ReplyWithInput, a.Pincode}, authorizer.DisplayPinCode(NewAuthTimeout(time.Duration(a.TimeoutMs)), a.Address, a.Pincode)
		}

	case DisplayPasskey:
		authfn = func() (AuthReply, error) {
			return AuthReply{ReplyWithInput, strconv.FormatUint(uint64(a.Passkey), 10)}, authorizer.DisplayPasskey(NewAuthTimeout(time.Duration(a.TimeoutMs)), a.Address, a.Passkey, a.Entered)
		}

	case ConfirmPasskey:
		authfn = func() (AuthReply, error) {
			return AuthReply{ReplyYesNo, "yes"}, authorizer.ConfirmPasskey(NewAuthTimeout(time.Duration(a.TimeoutMs)), a.Address, a.Passkey)
		}

	case AuthorizePairing:
		authfn = func() (AuthReply, error) {
			return AuthReply{ReplyYesNo, "yes"}, authorizer.AuthorizePairing(NewAuthTimeout(time.Duration(a.TimeoutMs)), a.Address)
		}

	case AuthorizeService:
		authfn = func() (AuthReply, error) {
			return AuthReply{ReplyYesNo, "yes"}, authorizer.AuthorizeService(NewAuthTimeout(time.Duration(a.TimeoutMs)), a.Address, a.UUID)
		}

	case AuthorizeTransfer:
		authfn = func() (AuthReply, error) {
			return AuthReply{ReplyYesNo, "yes"}, authorizer.AuthorizeTransfer(NewAuthTimeout(time.Duration(a.TimeoutMs)), a.FileTransfer)
		}
	}

	if authfn == nil {
		return errorkinds.ErrMethodCall
	}

	reply, err := authfn()
	cb(*a, reply, err)

	return nil
}

// NewAuthTimeout returns a new authentication timeout token.
func NewAuthTimeout(timeout time.Duration) AuthTimeout {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	return AuthTimeout{ctx, cancel}
}

// Done returns the inner context's Done() channel.
func (a *AuthTimeout) Done() <-chan struct{} {
	return a.ctx.Done()
}

// Cancel cancels the inner context.
func (a *AuthTimeout) Cancel() {
	a.cancel()
}

// DefaultAuthorizer describes a default authentication handler.
type DefaultAuthorizer struct{}

// AuthorizeTransfer accepts all file transfer authorization requests.
func (DefaultAuthorizer) AuthorizeTransfer(AuthTimeout, FileTransferData) error {
	return nil
}

// DisplayPinCode accepts all display pincode requests.
func (DefaultAuthorizer) DisplayPinCode(AuthTimeout, MacAddress, string) error {
	return nil
}

// DisplayPasskey accepts all display passkey requests.
func (DefaultAuthorizer) DisplayPasskey(AuthTimeout, MacAddress, uint32, uint16) error {
	return nil
}

// ConfirmPasskey accepts all passkey confirmation requests.
func (DefaultAuthorizer) ConfirmPasskey(AuthTimeout, MacAddress, uint32) error {
	return nil
}

// AuthorizePairing accepts all pairing authorization requests.
func (DefaultAuthorizer) AuthorizePairing(AuthTimeout, MacAddress) error {
	return nil
}

// AuthorizeService accepts all service (Bluetooth profile) authorization requests.
func (DefaultAuthorizer) AuthorizeService(AuthTimeout, MacAddress, uuid.UUID) error {
	return nil
}
