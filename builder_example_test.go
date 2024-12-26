package requests_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/requests/reqtest"
)

func init() {
	http.DefaultClient.Transport = reqtest.Recorder(reqtest.ModeReplay, nil, "testdata")
}

func Example() {
	// Simple GET into a string
	var s string
	err := requests.
		URL("http://example.com").
		ToString(&s).
		Fetch(context.Background())
	if err != nil {
		fmt.Println("could not connect to example.com:", err)
	}
	fmt.Println(strings.Contains(s, "Example Domain"))
	// Output:
	// true
}

func Example_getJSON() {
	// GET a JSON object
	id := 1
	var post placeholder
	err := requests.
		URL("https://jsonplaceholder.typicode.com").
		Pathf("/posts/%d", id).
		ToJSON(&post).
		Fetch(context.Background())
	if err != nil {
		fmt.Println("could not connect to jsonplaceholder.typicode.com:", err)
	}
	fmt.Println(post.Title)
	// Output:
	// sunt aut facere repellat provident occaecati excepturi optio reprehenderit
}

func Example_postJSON() {
	// POST a JSON object and parse the response
	var res placeholder
	req := placeholder{
		Title:  "foo",
		Body:   "baz",
		UserID: 1,
	}
	err := requests.
		URL("/posts").
		Host("jsonplaceholder.typicode.com").
		BodyJSON(&req).
		ToJSON(&res).
		Fetch(context.Background())
	if err != nil {
		fmt.Println("could not connect to jsonplaceholder.typicode.com:", err)
	}
	fmt.Println(res)
	// Output:
	// {101 foo baz 1}
}

func ExampleBuilder_ToBytesBuffer() {
	// Simple GET into a buffer
	var buf bytes.Buffer
	err := requests.
		URL("http://example.com").
		ToBytesBuffer(&buf).
		Fetch(context.Background())
	if err != nil {
		fmt.Println("could not connect to example.com:", err)
	}
	fmt.Println(strings.Contains(buf.String(), "Example Domain"))
	// Output:
	// true
}

func ExampleBuilder_ToWriter() {
	f, err := os.CreateTemp("", "*.to_writer.html")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name()) // clean up

	// suppose there is some io.Writer you want to stream to
	err = requests.
		URL("http://example.com").
		ToWriter(f).
		Fetch(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	if err = f.Close(); err != nil {
		log.Fatal(err)
	}
	stat, err := os.Stat(f.Name())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("file is %d bytes\n", stat.Size())

	// Output:
	// file is 1256 bytes
}

func ExampleBuilder_ToFile() {
	dir, err := os.MkdirTemp("", "to_file_*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir) // clean up

	exampleFilename := filepath.Join(dir, "example.txt")

	err = requests.
		URL("http://example.com").
		ToFile(exampleFilename).
		Fetch(context.Background())

	if err != nil {
		log.Fatal(err)
	}
	stat, err := os.Stat(exampleFilename)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("file is %d bytes\n", stat.Size())

	// Output:
	// file is 1256 bytes
}

type placeholder struct {
	ID     int    `json:"id,omitempty"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	UserID int    `json:"userId"`
}

func ExampleBuilder_Path() {
	// Add an ID to a base URL path
	id := 1
	u, err := requests.
		URL("https://api.example.com/posts/").
		// inherits path /posts from base URL
		Pathf("%d", id).
		URL()
	if err != nil {
		fmt.Println("Error!", err)
	}
	fmt.Println(u.String())
	// Output:
	// https://api.example.com/posts/1
}
func ExampleBuilder_CheckStatus() {
	// Expect a specific status code
	err := requests.
		URL("https://jsonplaceholder.typicode.com").
		Pathf("/posts/%d", 9001).
		CheckStatus(404).
		CheckContentType("application/json").
		Fetch(context.Background())
	if err != nil {
		fmt.Println("should be a 404:", err)
	} else {
		fmt.Println("OK")
	}
	// Output:
	// OK
}

func ExampleBuilder_CheckContentType() {
	// Expect a specific status code
	err := requests.
		URL("https://jsonplaceholder.typicode.com").
		Pathf("/posts/%d", 1).
		CheckContentType("application/bison").
		Fetch(context.Background())
	if err != nil {
		if re := new(requests.ResponseError); errors.As(err, &re) {
			fmt.Println("content-type was", re.Header.Get("Content-Type"))
		}
	}
	// Output:
	// content-type was application/json; charset=utf-8
}

// Examples with the Postman echo server
type postman struct {
	Args    map[string]string `json:"args"`
	Data    string            `json:"data"`
	Headers map[string]string `json:"headers"`
	JSON    map[string]string `json:"json"`
}

func Example_queryParam() {
	subdomain := "dev1"
	c := 4

	u, err := requests.
		URL("https://prod.example.com/get?a=1&b=2").
		Hostf("%s.example.com", subdomain).
		Param("b", "3").
		ParamInt("c", c).
		URL()
	if err != nil {
		fmt.Println("Error!", err)
	}
	fmt.Println(u.String())

	// Output:
	// https://dev1.example.com/get?a=1&b=3&c=4
}

func ExampleBuilder_Params() {
	// Conditionally add parameters
	values := url.Values{"a": {"1"}}
	values.Set("b", "3")
	if "cond" != "example" {
		values.Add("b", "4")
		values.Set("c", "5")
	}

	// Then add them to the URL
	u, err := requests.
		URL("https://www.example.com/get?a=0&z=6").
		Params(values).
		URL()
	if err != nil {
		fmt.Println("Error!", err)
	}
	fmt.Println(u.String())

	// Output:
	// https://www.example.com/get?a=1&b=3&b=4&c=5&z=6
}

func ExampleBuilder_Header() {
	// Set headers
	var headers postman
	err := requests.
		URL("https://postman-echo.com/get").
		UserAgent("bond/james-bond").
		BasicAuth("bondj", "007!").
		ContentType("secret").
		Header("martini", "shaken").
		ToJSON(&headers).
		Fetch(context.Background())
	if err != nil {
		fmt.Println("problem with postman:", err)
	}
	fmt.Println(headers.Headers["user-agent"])
	fmt.Println(headers.Headers["authorization"])
	fmt.Println(headers.Headers["content-type"])
	fmt.Println(headers.Headers["martini"])
	// Output:
	// bond/james-bond
	// Basic Ym9uZGo6MDA3IQ==
	// secret
	// shaken
}

func ExampleBuilder_Headers() {
	// Set headers conditionally
	h := make(http.Header)
	if "x-forwarded-for" != "true" {
		h.Add("x-forwarded-for", "127.0.0.1")
	}
	if "has-trace-id" != "true" {
		h.Add("x-trace-id", "abc123")
	}
	// Then add them to a request
	req, err := requests.
		URL("https://example.com").
		Headers(h).
		Request(context.Background())
	if err != nil {
		fmt.Println("Error!", err)
	}
	fmt.Println(req.Header)
	// Output:
	// map[X-Forwarded-For:[127.0.0.1] X-Trace-Id:[abc123]]

}
func ExampleBuilder_Bearer() {
	// We get a 401 response if no bearer token is provided
	err := requests.
		URL("http://httpbin.org/bearer").
		CheckStatus(http.StatusUnauthorized).
		Fetch(context.Background())
	if err != nil {
		fmt.Println("problem with httpbin:", err)
	}
	// But our response is accepted when we provide a bearer token
	var res struct {
		Authenticated bool
		Token         string
	}
	err = requests.
		URL("http://httpbin.org/bearer").
		Bearer("whatever").
		ToJSON(&res).
		Fetch(context.Background())
	if err != nil {
		fmt.Println("problem with httpbin:", err)
	}
	fmt.Println(res.Authenticated)
	fmt.Println(res.Token)
	// Output:
	// true
	// whatever
}

func ExampleBuilder_BodyBytes() {
	// Post a raw body
	var data postman
	err := requests.
		URL("https://postman-echo.com/post").
		BodyBytes([]byte(`hello, world`)).
		ContentType("text/plain").
		ToJSON(&data).
		Fetch(context.Background())
	if err != nil {
		fmt.Println("problem with postman:", err)
	}
	fmt.Println(data.Data)
	// Output:
	// hello, world
}

func ExampleBuilder_BodyReader() {
	// temp file creation boilerplate
	dir, err := os.MkdirTemp("", "body_reader_*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir) // clean up

	exampleFilename := filepath.Join(dir, "example.txt")
	exampleContent := `hello, world`
	if err := os.WriteFile(exampleFilename, []byte(exampleContent), 0644); err != nil {
		log.Fatal(err)
	}

	// suppose there is some io.Reader you want to stream from
	f, err := os.Open(exampleFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// send the raw file to server
	var echo postman
	err = requests.
		URL("https://postman-echo.com/post").
		ContentType("text/plain").
		BodyReader(f).
		ToJSON(&echo).
		Fetch(context.Background())
	if err != nil {
		fmt.Println("problem with postman:", err)
	}
	fmt.Println(echo.Data)
	// Output:
	// hello, world
}

func ExampleBuilder_CopyHeaders() {
	// Get headers while also getting body
	var s string
	headers := http.Header{}
	err := requests.
		URL("http://example.com").
		CopyHeaders(headers).
		// CopyHeaders disables status validation, so add it back
		CheckStatus(http.StatusOK).
		ToString(&s).
		Fetch(context.Background())
	if err != nil {
		fmt.Println("problem with example.com:", err)
	}
	fmt.Println(headers.Get("Etag"))
	fmt.Println(strings.Contains(s, "Example Domain"))
	// Output:
	// "3147526947+gzip"
	// true
}

func ExampleBuilder_ToHeaders() {
	// Send a HEAD request and look at headers
	headers := http.Header{}
	err := requests.
		URL("http://example.com").
		ToHeaders(headers).
		Fetch(context.Background())
	if err != nil {
		fmt.Println("problem with example.com:", err)
	}
	fmt.Println(headers.Get("Etag"))
	// Output:
	// "3147526947"
}

func ExampleBuilder_BodyWriter() {
	var echo postman
	err := requests.
		URL("https://postman-echo.com/post").
		ContentType("text/plain").
		BodyWriter(func(w io.Writer) error {
			cw := csv.NewWriter(w)
			cw.Write([]string{"col1", "col2"})
			cw.Write([]string{"val1", "val2"})
			cw.Flush()
			return cw.Error()
		}).
		ToJSON(&echo).
		Fetch(context.Background())
	if err != nil {
		fmt.Println("problem with postman:", err)
	}
	fmt.Printf("%q\n", echo.Data)
	// Output:
	// "col1,col2\nval1,val2\n"
}

func ExampleBuilder_BodyForm() {
	// Submit form values
	var echo postman
	err := requests.
		URL("https://postman-echo.com/put").
		Put().
		BodyForm(url.Values{
			"hello": []string{"world"},
		}).
		ToJSON(&echo).
		Fetch(context.Background())
	if err != nil {
		fmt.Println("problem with postman:", err)
	}
	fmt.Println(echo.JSON)
	// Output:
	// map[hello:world]
}

func ExampleBuilder_BodyFile() {
	// Make a file to read from
	dir, err := os.MkdirTemp("", "body_file_*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir) // clean up

	exampleFilename := filepath.Join(dir, "example.txt")
	exampleContent := `hello, world`
	if err = os.WriteFile(exampleFilename, []byte(exampleContent), 0644); err != nil {
		log.Fatal(err)
	}

	// Post a raw file
	var data postman
	err = requests.
		URL("https://postman-echo.com/post").
		BodyFile(exampleFilename).
		ContentType("text/plain").
		ToJSON(&data).
		Fetch(context.Background())
	if err != nil {
		fmt.Println("problem with postman:", err)
	}
	fmt.Println(data.Data)
	// Output:
	// hello, world
}

func ExampleBuilder_CheckPeek() {
	// Check that a response has a doctype
	const doctype = "<!doctype html>"
	var s string
	err := requests.
		URL("http://example.com").
		CheckPeek(len(doctype), func(b []byte) error {
			if string(b) != doctype {
				return fmt.Errorf("missing doctype: %q", b)
			}
			return nil
		}).
		ToString(&s).
		Fetch(context.Background())
	if err != nil {
		fmt.Println("could not connect to example.com:", err)
	}
	fmt.Println(
		// Final result still has the prefix
		strings.HasPrefix(s, doctype),
		// And the full body
		strings.HasSuffix(s, "</html>\n"),
	)
	// Output:
	// true true
}

func ExampleBuilder_Transport() {
	const text = "Hello, from transport!"
	var myCustomTransport requests.RoundTripFunc = func(req *http.Request) (res *http.Response, err error) {
		res = &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(text)),
		}
		return
	}
	var s string
	err := requests.
		URL("x://transport.example").
		Transport(myCustomTransport).
		ToString(&s).
		Fetch(context.Background())
	if err != nil {
		fmt.Println("transport failed:", err)
	}
	fmt.Println(s == text) // true
	// Output:
	// true
}

func ExampleBuilder_ErrorJSON() {
	{
		trans := requests.ReplayString(`HTTP/1.1 200 OK

	{"x": 1}`)

		var goodJSON struct{ X int }
		var errJSON struct{ Error string }
		err := requests.
			URL("http://example.com/").
			Transport(trans).
			ToJSON(&goodJSON).
			ErrorJSON(&errJSON).
			Fetch(context.Background())
		if err != nil {
			fmt.Println("Error!", err)
		} else {
			fmt.Println("X", goodJSON.X)
		}
	}
	{
		trans := requests.ReplayString(`HTTP/1.1 418 I'm a teapot

	{"error": "brewing"}`)

		var goodJSON struct{ X int }
		var errJSON struct{ Error string }
		err := requests.
			URL("http://example.com/").
			Transport(trans).
			ToJSON(&goodJSON).
			ErrorJSON(&errJSON).
			Fetch(context.Background())
		switch {
		case errors.Is(err, requests.ErrInvalidHandled):
			fmt.Println(errJSON.Error)
		case err != nil:
			fmt.Println("Error!", err)
		case err == nil:
			fmt.Println("unexpected success")
		}
	}
	// Output:
	// X 1
	// brewing
}

func ExampleBuilder_BodySerializer() {
	// Have some binary data
	data := struct {
		Header  [3]byte
		Payload uint32
	}{
		Header:  [3]byte([]byte("ABC")),
		Payload: 0xbadc0fee,
	}
	// Serialize it by just shoving data onto the wire
	serializer := func(v any) ([]byte, error) {
		var buf bytes.Buffer
		err := binary.Write(&buf, binary.BigEndian, v)
		return buf.Bytes(), err
	}
	// Make a request using the serializer
	req, err := requests.
		New().
		BodySerializer(serializer, data).
		Request(context.Background())
	if err != nil {
		panic(err)
	}
	b, err := io.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	// Request body is just serialized bytes
	fmt.Printf("%q", b)
	// Output:
	// "ABC\xba\xdc\x0f\xee"
}

func ExampleBuilder_ToDeserializer() {
	trans := requests.ReplayString(
		"HTTP/1.1 200 OK\r\n\r\nXYZ\x00\xde\xca\xff",
	)
	// Have some binary structure
	var data struct {
		Header  [3]byte
		Payload uint32
	}
	// Deserialize it by just pulling data off the wire
	deserializer := func(data []byte, v any) error {
		buf := bytes.NewReader(data)
		return binary.Read(buf, binary.BigEndian, v)
	}
	// Make a request using the deserializer
	err := requests.
		New().
		Transport(trans).
		ToDeserializer(deserializer, &data).
		Fetch(context.Background())
	if err != nil {
		panic(err)
	}

	// We read the data out of the response body
	fmt.Printf("%q, %X", data.Header, data.Payload)
	// Output:
	// "XYZ", DECAFF
}

func ExampleBuilder_BodyJSON() {
	// Restore defaults after this test
	defaultSerializer := requests.JSONSerializer
	defer func() {
		requests.JSONSerializer = defaultSerializer
	}()

	data := struct {
		A string `json:"a"`
		B int    `json:"b"`
		C []bool `json:"c"`
	}{
		"Hello", 42, []bool{true, false},
	}

	// Build a request using the default JSON serializer
	req, err := requests.
		New().
		BodyJSON(&data).
		Request(context.Background())
	if err != nil {
		panic(err)
	}

	// JSON is packed in with no whitespace
	io.Copy(os.Stdout, req.Body)
	fmt.Println()

	// Change the default JSON serializer to indent with two spaces
	requests.JSONSerializer = func(v any) ([]byte, error) {
		return json.MarshalIndent(v, "", "  ")
	}

	// Build a new request using the new indenting serializer
	req, err = requests.
		New().
		BodyJSON(&data).
		Request(context.Background())
	if err != nil {
		panic(err)
	}

	// Now the request body is indented
	io.Copy(os.Stdout, req.Body)
	fmt.Println()

	// Output:
	// {"a":"Hello","b":42,"c":[true,false]}
	// {
	//   "a": "Hello",
	//   "b": 42,
	//   "c": [
	//     true,
	//     false
	//   ]
	// }
}

func ExampleBuilder_ParamOptional() {
	// Suppose we have some variables from some external source
	yes := "1"
	no := ""

	u, err := requests.
		URL("https://www.example.com/?c=something").
		ParamOptional("a", yes).
		ParamOptional("b", no).  // Won't set ?b= because no is blank
		ParamOptional("c", yes). // Won't set ?c= because it was already in the base URL
		URL()
	if err != nil {
		fmt.Println("Error!", err)
	}
	fmt.Println(u.String())

	// Output:
	// https://www.example.com/?a=1&c=something
}

func ExampleBuilder_HeaderOptional() {
	// Suppose we have some environment variables
	// which may or may not be set
	env := map[string]string{
		"FOO": "1",
		"BAR": "",
		"BAZ": "",
	}
	req, err := requests.
		URL("https://example.com").
		HeaderOptional("X-FOO", env["FOO"]).
		HeaderOptional("X-BAR", env["BAR"]). // Won't set because BAR is blank
		Header("X-BAZ", env["BAZ"]).         // Will set to "" because it's not optional
		Request(context.Background())
	if err != nil {
		fmt.Println("Error!", err)
	}
	fmt.Println(req.Header)

	// Output:
	// map[X-Baz:[] X-Foo:[1]]
}
