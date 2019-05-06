package xhttp

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func createTestHandler(code int, msg string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		w.Write([]byte(msg))
	}
}

func TestGet(t *testing.T) {
	type testItem struct {
		code int
		msg  string
	}

	type testCase struct {
		expected *testItem
		actual   *Response
	}

	tests := []*testCase{
		{expected: &testItem{200, "success"}},
		{expected: &testItem{400, "bad request"}},
		{expected: &testItem{500, "internal server error"}},
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
