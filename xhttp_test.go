package xhttp

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

type testCase struct {
	expected *testRequest
	actual   *Response
}

type testRequest struct {
	code int
	mtd  string
	msg  string
	body []byte
	fail bool
}

func TestInterface(t *testing.T) {
	type Interface interface {
		CloseIdleConnections()
		Do(req *http.Request) (*http.Response, error)
		Get(url string) (resp *http.Response, err error)
		Head(url string) (resp *http.Response, err error)
		Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
		PostForm(url string, data url.Values) (resp *http.Response, err error)
	}

	var (
		_ Interface = &http.Client{}
		_ Interface = NewClient()
	)
}

func TestNewClient(t *testing.T) {
	c := NewClient()

	testClientGet(t, c)
}

func TestNewClientWithConfig(t *testing.T) {
	c := NewClientWithConfig(DefaultClientConfig())

	testClientGet(t, c)
}

func TestGET(t *testing.T) {
	tests := []*testCase{
		{expected: &testRequest{code: 200, msg: "success"}},
		{expected: &testRequest{code: 400, msg: "bad request"}},
		{expected: &testRequest{code: 500, msg: "internal server error"}},
	}

	for _, tc := range tests {
		t.Run(tc.expected.msg, func(tt *testing.T) {
			srv := httptest.NewServer(http.Handler(createTestHandler(tc.expected.code, tc.expected.msg)))

			var err error
			tc.actual, err = GET(srv.URL)
			if err != nil {
				srv.Close()
				tt.Errorf("failed to GET: %s", err)
			}
			srv.Close()

			if tc.actual.StatusCode != tc.expected.code {
				tt.Errorf("expected resp code %d, actual is %d", tc.expected.code, tc.actual.StatusCode)
			}
		})
	}
}

func TestPOST(t *testing.T) {
	tests := []*testCase{
		{expected: &testRequest{code: 200, mtd: http.MethodPost, msg: "success", body: []byte("success")}},
		{expected: &testRequest{code: 400, mtd: http.MethodPost, msg: "bad request", body: []byte("bad request")}},
		{expected: &testRequest{code: 500, mtd: http.MethodPost, msg: "internal server error", body: []byte("internal server error")}},
	}

	for _, tc := range tests {
		t.Run(tc.expected.msg, func(tt *testing.T) {
			srv := httptest.NewServer(http.Handler(createTestHandler(tc.expected.code, tc.expected.msg)))

			var err error
			tc.actual, err = POST(srv.URL, "application/octet-stream", tc.expected.body)
			if err != nil {
				srv.Close()
				tt.Errorf("failed to Send: %s", err)
			}
			srv.Close()

			if tc.actual.StatusCode != tc.expected.code {
				tt.Errorf("expected resp code %d, actual is %d", tc.expected.code, tc.actual.StatusCode)
			}
		})
	}
}

func TestSend(t *testing.T) {
	tests := []*testCase{
		{expected: &testRequest{code: 200, mtd: http.MethodGet, msg: "success"}},
		{expected: &testRequest{code: 400, mtd: http.MethodGet, msg: "bad request"}},
		{expected: &testRequest{code: 500, mtd: http.MethodGet, msg: "internal server error"}},
		{expected: &testRequest{code: 0, mtd: "WRONG", msg: "wrong", fail: true}},
	}

	for _, tc := range tests {
		t.Run(tc.expected.msg, func(tt *testing.T) {
			srv := httptest.NewServer(http.Handler(createTestHandler(tc.expected.code, tc.expected.msg)))

			req := NewRequest(tc.expected.mtd, srv.URL, tc.expected.body)

			var err error
			tc.actual, err = Send(req)
			if err != nil {
				srv.Close()
				if tc.expected.fail {
					return
				}

				tt.Errorf("failed to Send: %s", err)
			}
			srv.Close()

			if tc.actual.StatusCode != tc.expected.code {
				tt.Errorf("expected resp code %d, actual is %d", tc.expected.code, tc.actual.StatusCode)
			}
		})
	}
}

func TestDownloadFile(t *testing.T) {
	tests := []*testCase{
		{expected: &testRequest{code: 200, msg: "success"}},
		{expected: &testRequest{code: 400, msg: "bad request", fail: true}},
	}

	for _, tc := range tests {
		t.Run(tc.expected.msg, func(tt *testing.T) {
			tcFile, err := ioutil.TempFile(os.TempDir(), "_xhttp_DownloadFile")
			if err != nil {
				tt.Errorf("failed to create test file: %s", err)
			}
			tcFile.Close()

			srv := httptest.NewServer(http.Handler(createTestHandler(tc.expected.code, tc.expected.msg)))
			if err := DownloadFile(srv.URL, tcFile.Name()); err != nil {
				os.Remove(tcFile.Name())
				srv.Close()

				if !tc.expected.fail {
					tt.Errorf("failed to DownloadFile: %s", err)
				}

				if tc.expected.fail {
					return
				}
			}

			actual, err := ioutil.ReadFile(tcFile.Name())
			if err != nil {
				os.Remove(tcFile.Name())
				tt.Errorf("failed to Read file: %s", err)
			}

			if !bytes.Equal([]byte(tc.expected.msg), actual) {
				os.Remove(tcFile.Name())
				tt.Errorf("expected: %s, got %s", tc.expected.msg, string(actual))
			}

			os.Remove(tcFile.Name())
		})
	}
}

func TestDo(t *testing.T) {
	type testCase struct {
		expected *testRequest
		actual   *http.Response
	}

	tests := []*testCase{
		{expected: &testRequest{code: 200, mtd: http.MethodGet, msg: "success"}},
		{expected: &testRequest{code: 400, mtd: http.MethodGet, msg: "bad request"}},
		{expected: &testRequest{code: 500, mtd: http.MethodGet, msg: "internal server error"}},
	}

	for _, tc := range tests {
		t.Run(tc.expected.msg, func(tt *testing.T) {
			srv := httptest.NewServer(http.Handler(createTestHandler(tc.expected.code, tc.expected.msg)))

			req, err := http.NewRequest(tc.expected.mtd, srv.URL, bytes.NewBuffer(tc.expected.body))
			if err != nil {
				srv.Close()
				tt.Errorf("failed to Send: %s", err)
			}

			tc.actual, err = Do(req)
			if err != nil {
				srv.Close()
				tt.Errorf("failed to Send: %s", err)
			}
			srv.Close()

			if tc.actual.StatusCode != tc.expected.code {
				tt.Errorf("expected resp code %d, actual is %d", tc.expected.code, tc.actual.StatusCode)
			}
		})
	}
}

func TestRequest_SetContentTypeJSON(t *testing.T) {
	expHeader := "Content-Type"
	expValue := "application/json"

	// Case with non-nil Request.Header map.
	req := NewRequest("http://localhost", http.MethodGet, []byte(nil))
	req.SetContentTypeJSON()

	if act := req.Header.Get(expHeader); act != expValue {
		t.Errorf("expected %s, got %s", expValue, act)
	}

	// Case with nil Request.Header map.
	req = &Request{
		BaseURL: "http://localhost",
		Method:  http.MethodGet,
	}

	req.SetContentTypeJSON()

	if act := req.Header.Get(expHeader); act != expValue {
		t.Errorf("expected %s, got %s", expValue, act)
	}
}

func TestRequest_SetContentType(t *testing.T) {
	expHeader := "Content-Type"
	expValue := "application/test-content-type"

	// Case with non-nil Request.Header map.
	req := NewRequest("http://localhost", http.MethodGet, []byte(nil))
	req.SetContentType(expValue)

	if act := req.Header.Get(expHeader); act != expValue {
		t.Errorf("expected %s, got %s", expValue, act)
	}

	// Case with nil Request.Header map.
	req = &Request{
		BaseURL: "http://localhost",
		Method:  http.MethodGet,
	}

	req.SetContentType(expValue)

	if act := req.Header.Get(expHeader); act != expValue {
		t.Errorf("expected %s, got %s", expValue, act)
	}
}

func TestRequest_SetAuthorization(t *testing.T) {
	expHeader := "Authorization"
	expValue := "Test: Authorization"

	// Case with non-nil Request.Header map.
	req := NewRequest("http://localhost", http.MethodGet, []byte(nil))
	req.SetAuthorization(expValue)

	if act := req.Header.Get(expHeader); act != expValue {
		t.Errorf("expected %s, got %s", expValue, act)
	}

	// Case with nil Request.Header map.
	req = &Request{
		BaseURL: "http://localhost",
		Method:  http.MethodGet,
	}

	req.SetAuthorization(expValue)

	if act := req.Header.Get(expHeader); act != expValue {
		t.Errorf("expected %s, got %s", expValue, act)
	}
}

// Helpers.

func createTestHandler(code int, msg string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		w.Write([]byte(msg))
	}
}

func testClientGet(t *testing.T, c *Client) {
	tests := []*testCase{
		{expected: &testRequest{code: 200, msg: "success"}},
		{expected: &testRequest{code: 400, msg: "bad request"}},
		{expected: &testRequest{code: 500, msg: "internal server error"}},
	}

	for _, tc := range tests {
		t.Run(tc.expected.msg, func(tt *testing.T) {
			srv := httptest.NewServer(http.Handler(createTestHandler(tc.expected.code, tc.expected.msg)))

			var err error
			tc.actual, err = c.GET(srv.URL)
			if err != nil {
				srv.Close()
				tt.Errorf("failed to GET: %s", err)
			}
			srv.Close()

			if tc.actual.StatusCode != tc.expected.code {
				tt.Errorf("expected resp code %d, actual is %d", tc.expected.code, tc.actual.StatusCode)
			}
		})
	}
}

//
