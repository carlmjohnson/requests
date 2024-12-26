package reqtest_test

import (
	"context"
	"fmt"
	"os"
	"testing/fstest"

	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/requests/reqtest"
)

func ExampleReplayString() {
	const res = `HTTP/1.1 200 OK

An example response.`

	var s string
	const expected = `An example response.`
	if err := requests.
		URL("http://response.example").
		Transport(reqtest.ReplayString(res)).
		ToString(&s).
		Fetch(context.Background()); err != nil {
		panic(err)
	}
	fmt.Println(s == expected)
	// Output:
	// true
}

func copyToTempDir(m map[string]string) string {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		panic(err)
	}
	fsys := make(fstest.MapFS, len(m))
	for path, content := range m {
		fsys[path] = &fstest.MapFile{
			Data: []byte(content),
		}
	}
	if err = os.CopyFS(dir, fsys); err != nil {
		panic(err)
	}
	return dir
}

func ExampleRecorder() {
	// Given a directory with the following file
	dir := copyToTempDir(map[string]string{
		"fsys.example - MKIYDwjs.res.txt": `HTTP/1.1 200 OK
Content-Type: text/plain; charset=UTF-8
Date: Mon, 24 May 2021 18:48:50 GMT

An example response.`,
	})
	defer os.RemoveAll(dir)

	// Make a test transport that reads the directory
	tr := reqtest.Recorder(reqtest.ModeReplay, nil, dir)

	// And test that it produces the correct response
	var s string
	const expected = `An example response.`
	if err := requests.
		URL("http://fsys.example").
		Transport(tr).
		ToString(&s).
		Fetch(context.Background()); err != nil {
		panic(err)
	}
	fmt.Println(s == expected)

	// Output:
	// true
}
