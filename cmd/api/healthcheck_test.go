package main

import (
	"net/http"
	"testing"
)

func TestHealthcheck(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	for i := 0; i < 2; i++ {
		if i == 1 {
			storedMarshal := jsonMarshal
			jsonMarshal = ts.fakeMarshal
			defer ts.restoreMarshal(storedMarshal)
		}
		code, _, _ := ts.get(t, "/v1/healthcheck")

		if i == 0 && code != http.StatusOK {
			t.Errorf("want %d; got %d", http.StatusOK, code)
		}
		if i == 1 && code != http.StatusInternalServerError {
			t.Errorf("want %d; got %d", http.StatusOK, code)
		}
	}
}
