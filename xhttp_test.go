package xhttp

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testRequest struct {
	code int
	mtd  string
	msg  string
	body []byte
	err  error
}

func createTestHandler(code int, msg string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		w.Write([]byte(msg))
	}
}

func TestGet(t *testing.T) {
	type testCase struct {
		expected *testRequest
		actual   *Response
	}

	tests := []*testCase{
		{expected: &testRequest{code: 200, msg: "success"}},
		{expected: &testRequest{code: 400, msg: "bad request"}},
		{expected: &testRequest{code: 500, msg: "internal server error"}},
	}

	for _, tc := range tests {
		t.Run(tc.expected.msg, func(tt *testing.T) {
			srv := httptest.NewServer(http.Handler(createTestHandler(tc.expected.code, tc.expected.msg)))

			var err error
			tc.actual, err = Get(srv.URL)
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

func TestSend(t *testing.T) {
	type testCase struct {
		expected *testRequest
		actual   *Response
	}

	tests := []*testCase{
		{expected: &testRequest{code: 200, mtd: http.MethodGet, msg: "success"}},
		{expected: &testRequest{code: 400, mtd: http.MethodGet, msg: "bad request"}},
		{expected: &testRequest{code: 500, mtd: http.MethodGet, msg: "internal server error"}},
	}

	for _, tc := range tests {
		t.Run(tc.expected.msg, func(tt *testing.T) {
			srv := httptest.NewServer(http.Handler(createTestHandler(tc.expected.code, tc.expected.msg)))

			req := NewRequest(tc.expected.mtd, srv.URL, tc.expected.body)

			var err error
			tc.actual, err = Send(req)
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

func TestSendRaw(t *testing.T) {
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

			tc.actual, err = SendRaw(req)
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

func TestPost(t *testing.T) {
	type testCase struct {
		expected *testRequest
		actual   *Response
	}

	tests := []*testCase{
		{expected: &testRequest{code: 200, mtd: http.MethodPost, msg: "success", body: []byte("success")}},
		{expected: &testRequest{code: 400, mtd: http.MethodPost, msg: "bad request", body: []byte("bad request")}},
		{expected: &testRequest{code: 500, mtd: http.MethodPost, msg: "internal server error", body: []byte("internal server error")}},
	}

	for _, tc := range tests {
		t.Run(tc.expected.msg, func(tt *testing.T) {
			srv := httptest.NewServer(http.Handler(createTestHandler(tc.expected.code, tc.expected.msg)))

			var err error
			tc.actual, err = Post(srv.URL, "application/octet-stream", tc.expected.body)
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
