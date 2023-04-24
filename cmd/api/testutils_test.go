package main

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"greenlight.bcc/internal/data"
	"greenlight.bcc/internal/jsonlog"
)

func newTestApplication(t *testing.T) *application {

	cfg := config{limiter: struct {
		rps     float64
		burst   int
		enabled bool
	}{2, 4, true}}

	return &application{
		config: cfg,
		logger: jsonlog.New(io.Discard, jsonlog.LevelFatal),
		models: data.NewMockModels(),
	}
}

type testServer struct {
	*httptest.Server
}

func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewServer(h)

	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{ts}
}

func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, string) {
	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	bytes.TrimSpace(body)

	return rs.StatusCode, rs.Header, string(body)
}

func (ts *testServer) deleteReq(t *testing.T, urlPath string) (int, http.Header, string) {
	req, err := http.NewRequest(http.MethodDelete, ts.URL+urlPath, nil)
	if err != nil {
		t.Fatal(err)
	}

	rs, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	bytes.TrimSpace(body)

	return rs.StatusCode, rs.Header, string(body)
}

func (ts *testServer) postForm(t *testing.T, urlPath string, data []byte) (int, http.Header, string) {
	reader := bytes.NewReader(data)
	rs, err := ts.Client().Post(ts.URL+urlPath, "application/json", reader)
	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	bytes.TrimSpace(body)

	return rs.StatusCode, rs.Header, string(body)
}

func (ts *testServer) patchReq(t *testing.T, urlPath string, data []byte) (int, http.Header, string) {
	reader := bytes.NewReader(data)

	req, err := http.NewRequest(http.MethodPatch, ts.URL+urlPath, reader)
	if err != nil {
		t.Fatal(err)
	}

	rs, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	bytes.TrimSpace(body)

	return rs.StatusCode, rs.Header, string(body)
}

func (ts *testServer) fakeMarshal(v any) ([]byte, error) {
	return []byte{}, errors.New("marshalling failed")
}

func (ts *testServer) restoreMarshal(replace func(v interface{}) ([]byte, error)) {
	jsonMarshal = replace
}
func (ts *testServer) fakeUnmarshal(w http.ResponseWriter, r *http.Request, v any) error {
	return errors.New("marshalling failed")
}

func (ts *testServer) updateReq(t *testing.T, urlPath string, data []byte) (int, http.Header, string) {
	reader := bytes.NewReader(data)

	req, err := http.NewRequest(http.MethodPut, ts.URL+urlPath, reader)
	if err != nil {
		t.Fatal(err)
	}

	rs, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Body.Close()

	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	bytes.TrimSpace(body)

	return rs.StatusCode, rs.Header, string(body)
}

func (ts *testServer) patchForAuth(t *testing.T, urlPath string, data []byte, token string) (int, http.Header, string) {
	reader := bytes.NewReader(data)

	req, err := http.NewRequest(http.MethodPatch, ts.URL+urlPath, reader)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	rs, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Body.Close()

	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	bytes.TrimSpace(body)

	return rs.StatusCode, rs.Header, string(body)
}
