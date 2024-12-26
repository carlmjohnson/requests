package reqtest

import (
	"net/http"
)

// RecorderMode is an argument type controlling [Recorder].
type RecorderMode int8

//go:generate stringer -type=RecorderMode

// Enum values for type RecorderMode
const (
	// Record HTTP requests and responses to text files.
	ModeRecord RecorderMode = iota
	// Replay responses from pre-recorded text files.
	ModeReplay
	// Replay responses from pre-recorded files if present,
	// otherwise record a new request/response pair.
	ModeCache
)

// Recorder returns an HTTP transport that operates in the specified mode.
// Requests and responses are read from or written to
// text files in basepath according to a hash of their contents.
// File names may optionally be prefixed with comments for better human organization.
// The http.RoundTripper is only used in [ModeRecord] and [ModeCache]
// and if nil defaults to [http.DefaultTransport].
func Recorder(mode RecorderMode, rt http.RoundTripper, basepath string) http.RoundTripper {
	switch mode {
	case ModeReplay:
		return Replay(basepath)
	case ModeRecord:
		return Record(rt, basepath)
	case ModeCache:
		return Caching(rt, basepath)
	default:
		panic("invalid reqtest.RecorderMode")
	}
}
