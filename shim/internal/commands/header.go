package commands

import "encoding/binary"

const RawCommandHeaderSize = 18

type RawCommandHeaderBuffer = [RawCommandHeaderSize]byte

type RawCommandHeader struct {
	ApiVersion  byte
	InfoHeader  byte
	RequestId   int64
	OperationId uint32
	ContentSize uint32
}

type UnpackedRawCommandHeader struct {
	RawCommandHeader

	// Parsed from InfoHeader.
	IsOperationComplete bool
	EventID             byte
}

func UnpackReplyHeader(rawheader [RawCommandHeaderSize]byte) (UnpackedRawCommandHeader, error) {
	var unpacked UnpackedRawCommandHeader

	var header RawCommandHeader
	if _, err := binary.Decode(rawheader[:], binary.BigEndian, &header); err != nil {
		return unpacked, err
	}

	unpacked.RawCommandHeader = header

	flags := (header.InfoHeader >> 4) & 0x0f
	unpacked.IsOperationComplete = (flags >> 0) > 0
	unpacked.EventID = (header.InfoHeader) & 0x0f

	return unpacked, nil
}
