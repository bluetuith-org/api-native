package bluetooth

import (
	"context"
)

// Obex describes a function call interface to invoke Obex related functions
// on specified devices.
type Obex interface {
	// FileTransfer returns a function call interface to invoke device file transfer
	// related functions.
	FileTransfer() ObexFileTransfer
}

// ObexFileTransfer describes a function call interface to manage file-transfer
// related functions on specified devices.
type ObexFileTransfer interface {
	// CreateSession creates a new Obex session with a device.
	// The context (ctx) can be provided in case this function call
	// needs to be cancelled, since this function call can take some time
	// to complete.
	CreateSession(ctx context.Context) error

	// RemoveSession removes a created Obex session.
	RemoveSession() error

	// SendFile sends a file to the device. The 'filepath' must be a full path to the file.
	SendFile(filepath string) (FileTransferData, error)

	// CancelTransfer cancels the transfer.
	CancelTransfer() error

	// SuspendTransfer suspends the transfer.
	SuspendTransfer() error

	// ResumeTransfer resumes the transfer.
	ResumeTransfer() error
}

// FileTransferStatus describes the status of the file transfer.
type FileTransferStatus string

// The different transfer status types.
const (
	TransferInProgress FileTransferStatus = "in-progress"
	TransferSuccess    FileTransferStatus = "success"
	TransferError      FileTransferStatus = "error"
)

// FileTransferData holds the static file transfer data for a device.
type FileTransferData struct {
	// Name is the name of the file.
	Name string

	// Type is the type of the file (mime-type).
	Type string

	// Status indicates the file transfer status.
	Status FileTransferStatus

	// Filename is the name of the file.
	Filename string

	FileTransferEventData
}

// FileTransferEventData holds the dynamic (variable) file transfer data for a device.
// This is primarily used to send file transfer event related data.
type FileTransferEventData struct {
	Address MacAddress

	Size        uint64
	Transferred uint64
}

// ReceiveFileAuthHandler describes an authentication interface, which is used
// to authorize a file transfer being received, before starting the transfer.
type ReceiveFileAuthHandler interface {
	AuthorizeTransfer(ctx context.Context, path string, props FileTransferData) error
}
