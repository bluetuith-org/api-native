package bluetooth

// MediaPlayer describes a function call interface to invoke media player/control
// related functions on a device.
type MediaPlayer interface {
	Properties() (MediaData, error)

	Play() error
	Pause() error
	TogglePlayPause() error

	Next() error
	Previous() error
	FastForward() error
	Rewind() error

	Stop() error
}

// MediaData holds the media player information.
type MediaData struct {
	Status   string
	Position uint32

	TrackData
}

// MediaEventData holds the media player event information.
type MediaEventData struct {
	Address MacAddress

	MediaData
}

// TrackData describes the track properties of
// the currently playing media.
type TrackData struct {
	Title       string
	Album       string
	Artist      string
	Duration    uint32
	TrackNumber uint32
	TotalTracks uint32
}
