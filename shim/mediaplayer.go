//go:build !linux

package shim

import (
	"github.com/bluetuith-org/api-native/api/bluetooth"
	"github.com/bluetuith-org/api-native/api/errorkinds"
)

type mediaPlayer struct {
}

func (m *mediaPlayer) Properties() (bluetooth.MediaData, error) {
	return bluetooth.MediaData{}, errorkinds.ErrNotSupported
}

func (m *mediaPlayer) Play() error {
	return errorkinds.ErrNotSupported
}

func (m *mediaPlayer) Pause() error {
	return errorkinds.ErrNotSupported
}

func (m *mediaPlayer) TogglePlayPause() error {
	return errorkinds.ErrNotSupported
}

func (m *mediaPlayer) Next() error {
	return errorkinds.ErrNotSupported
}

func (m *mediaPlayer) Previous() error {
	return errorkinds.ErrNotSupported
}

func (m *mediaPlayer) FastForward() error {
	return errorkinds.ErrNotSupported
}

func (m *mediaPlayer) Rewind() error {
	return errorkinds.ErrNotSupported
}

func (m *mediaPlayer) Stop() error {
	return errorkinds.ErrNotSupported
}
