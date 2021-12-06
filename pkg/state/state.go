package state

import (
	"encoding/json"

	"github.com/pkg/errors"
)

var _ json.Unmarshaler = (*State)(nil)

type State uint8

func (s *State) UnmarshalJSON(bytes []byte) error {
	v := string(bytes)

	switch v {
	case `"moving"`:
		*s = Moving
	case `"uploading"`:
		*s = Uploading
	case `"stalledUP"`:
		*s = StalledUploading
	case `"downloading"`:
		*s = Downloading
	case `"stalledDL"`:
		*s = StalledDownloading
	case `"pausedUP"`:
		*s = PausedUploading
	case `"pausedDL"`:
		*s = PausedDownloading
	case `"checkingUP"`:
		*s = CheckingUploading
	case `"checkingDL"`:
		*s = CheckingDownloading
	default:
		return errors.New("unknown state " + v)
	}

	return nil
}

func (s *State) String() string {
	switch *s {
	case Moving:
		return `"moving"`
	case Uploading:
		return `"uploading"`
	case StalledUploading:
		return `"stalledUP"`
	case Downloading:
		return `"downloading"`
	case StalledDownloading:
		return `"stalledDL"`
	case PausedUploading:
		return `"pausedUP"`
	case PausedDownloading:
		return `"pausedDL"`
	case CheckingUploading:
		return `"checkingUP"`
	case CheckingDownloading:
		return `"checkingDL"`
	}

	return "unknown"
}

const (
	BasicStalled State = 1 << iota
	BasicPaused
	BasicUploading
	BasicDownloading
	BasicChecking
	BasicMoving
)

const (
	Moving      = BasicMoving
	Checking    = BasicChecking
	Uploading   = BasicUploading
	Downloading = BasicDownloading

	StalledUploading = BasicUploading | BasicStalled
	PausedUploading  = BasicUploading | BasicPaused

	StalledDownloading = BasicDownloading | BasicStalled
	PausedDownloading  = BasicDownloading | BasicPaused

	CheckingUploading   = Checking | BasicUploading
	CheckingDownloading = Checking | BasicDownloading
)
