package reqtest

import (
	"io/fs"
	"net/http"

	"github.com/carlmjohnson/requests"
)

// ReplayString returns an http.RoundTripper that always responds with a
// request built from rawResponse. It is intended for use in one-off tests.
func ReplayString(rawResponse string) requests.Transport {
	return requests.ReplayString(rawResponse)
}

// Record returns an http.RoundTripper that writes out its
// requests and their responses to text files in basepath.
// Requests are named according to a hash of their contents.
// Responses are named according to the request that made them.
//
// Deprecated: Use [Recorder].
func Record(rt http.RoundTripper, basepath string) requests.Transport {
	return requests.Record(rt, basepath)
}

// Replay returns an http.RoundTripper that reads its
// responses from text files in basepath.
// Responses are looked up according to a hash of the request.
// Response file names may optionally be prefixed with comments for better human organization.
//
// Deprecated: Use [Recorder].
func Replay(basepath string) requests.Transport {
	return requests.Replay(basepath)
}

// ReplayFS returns an http.RoundTripper that reads its
// responses from text files in the fs.FS.
// Responses are looked up according to a hash of the request.
// Response file names may optionally be prefixed with comments for better human organization.
//
// Deprecated: Use os.CopyFS and [Recorder].
func ReplayFS(fsys fs.FS) requests.Transport {
	return requests.ReplayFS(fsys)
}

// Caching returns an http.RoundTripper that attempts to read its
// responses from text files in basepath. If the response is absent,
// it caches the result of issuing the request with rt in basepath.
// Requests are named according to a hash of their contents.
// Responses are named according to the request that made them.
//
// Deprecated: Use [Recorder].
func Caching(rt http.RoundTripper, basepath string) requests.Transport {
	return requests.Caching(rt, basepath)
}
