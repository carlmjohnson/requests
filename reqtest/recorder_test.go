package reqtest_test

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/requests/internal/be"
	"github.com/carlmjohnson/requests/reqtest"
)

func TestRecorder(t *testing.T) {
	baseTrans := requests.ReplayString(`HTTP/1.1 200 OK

Test Document 1`)
	dir := t.TempDir()

	var s1, s2 string
	err := requests.URL("http://example.com").
		Transport(reqtest.Recorder(reqtest.ModeRecord, baseTrans, dir)).
		ToString(&s1).
		Fetch(context.Background())
	be.NilErr(t, err)

	err = requests.URL("http://example.com").
		Transport(reqtest.Recorder(reqtest.ModeReplay, nil, dir)).
		ToString(&s2).
		Fetch(context.Background())
	be.NilErr(t, err)
	be.Equal(t, s1, s2)
	be.Equal(t, "Test Document 1", s1)
}

func TestCaching(t *testing.T) {
	dir := t.TempDir()
	hasRun := false
	content := "some content"
	var onceTrans requests.RoundTripFunc = func(req *http.Request) (res *http.Response, err error) {
		be.False(t, hasRun)
		hasRun = true
		res = &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(content)),
		}
		return
	}
	trans := reqtest.Recorder(reqtest.ModeCache, onceTrans, dir)
	var s1, s2 string
	err := requests.URL("http://example.com").
		Transport(trans).
		ToString(&s1).
		Fetch(context.Background())
	be.NilErr(t, err)
	err = requests.URL("http://example.com").
		Transport(trans).
		ToString(&s2).
		Fetch(context.Background())
	be.NilErr(t, err)
	be.Equal(t, content, s1)
	be.Equal(t, s1, s2)

	entries, err := os.ReadDir(dir)
	be.NilErr(t, err)
	be.Equal(t, 2, len(entries))
}
