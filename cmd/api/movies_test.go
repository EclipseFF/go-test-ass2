package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"greenlight.bcc/internal/assert"
)

func TestShowMovie(t *testing.T) {
	app := newTestApplication(t)

	ts := newTestServer(t, app.routesTest())
	defer ts.Close()

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody string
	}{
		{
			name:     "Valid ID",
			urlPath:  "/v1/movies/1",
			wantCode: http.StatusOK,
		},
		{
			name:     "Non-existent ID",
			urlPath:  "/v1/movies/3",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Negative ID",
			urlPath:  "/v1/movies/-1",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Decimal ID",
			urlPath:  "/v1/movies/1.23",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "String ID",
			urlPath:  "/v1/movies/foo",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Database fall",
			urlPath:  "/v1/movies/2",
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "Fake json.Write",
			urlPath:  "/v1/movies/1",
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "Fake json.Write" {
				storedMarshal := jsonMarshal
				jsonMarshal = ts.fakeMarshal
				defer ts.restoreMarshal(storedMarshal)
			}

			code, _, body := ts.get(t, tt.urlPath)

			assert.Equal(t, code, tt.wantCode)

			if tt.wantBody != "" {
				assert.StringContains(t, body, tt.wantBody)
			}

		})
	}

}

func TestCreateMovie(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routesTest())
	defer ts.Close()

	const (
		validTitle    = "Test Name"
		validYear     = 2021
		validRuntime  = "105 mins"
		repeatedTitle = "Repeated Title"
	)

	validGenres := []string{"comedy", "drama"}

	tests := []struct {
		name     string
		Title    string
		Year     int32
		Runtime  string
		Genres   []string
		wantCode int
	}{
		{
			name:     "Valid submission",
			Title:    validTitle,
			Year:     validYear,
			Runtime:  validRuntime,
			Genres:   validGenres,
			wantCode: http.StatusCreated,
		},
		{
			name:     "Empty Name",
			Title:    "",
			Year:     validYear,
			Runtime:  validRuntime,
			Genres:   validGenres,
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name:     "year < 1888",
			Title:    validTitle,
			Year:     1500,
			Runtime:  validRuntime,
			Genres:   validGenres,
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name:     "test for wrong input",
			Title:    validTitle,
			Year:     validYear,
			Runtime:  validRuntime,
			Genres:   validGenres,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "same title",
			Title:    repeatedTitle,
			Year:     validYear,
			Runtime:  validRuntime,
			Genres:   validGenres,
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "Fake json.Write",
			Title:    validTitle,
			Year:     validYear,
			Runtime:  validRuntime,
			Genres:   validGenres,
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "Fake json.Write" {
				storedMarshal := jsonMarshal
				jsonMarshal = ts.fakeMarshal
				defer ts.restoreMarshal(storedMarshal)
			}

			inputData := struct {
				Title   string   `json:"title"`
				Year    int32    `json:"year"`
				Runtime string   `json:"runtime"`
				Genres  []string `json:"genres"`
			}{
				Title:   tt.Title,
				Year:    tt.Year,
				Runtime: tt.Runtime,
				Genres:  tt.Genres,
			}

			b, err := json.Marshal(&inputData)
			if err != nil {
				t.Fatal("wrong input data")
			}
			if tt.name == "test for wrong input" {
				b = append(b, 'a')
			}

			code, _, _ := ts.postForm(t, "/v1/movies", b)

			assert.Equal(t, code, tt.wantCode)

		})
	}
}

func TestUpdateMovie(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routesTest())
	defer ts.Close()

	const (
		validTitle    = "Test Title"
		validYear     = 2021
		validRuntime  = "105 mins"
		conflictTitle = "Conflict Title"
	)

	validGenres := []string{"comedy", "drama"}

	tests := []struct {
		name     string
		urlPath  string
		Title    string
		Year     int32
		Runtime  string
		Genres   []string
		wantCode int
	}{
		{
			name:     "Valid submission",
			urlPath:  "/v1/movies/1",
			Title:    validTitle,
			Year:     validYear,
			Runtime:  validRuntime,
			Genres:   validGenres,
			wantCode: http.StatusOK,
		},
		{
			name:     "Empty Name",
			urlPath:  "/v1/movies/1",
			Title:    "",
			Year:     validYear,
			Runtime:  validRuntime,
			Genres:   validGenres,
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name:     "year < 1888",
			urlPath:  "/v1/movies/1",
			Title:    validTitle,
			Year:     1500,
			Runtime:  validRuntime,
			Genres:   validGenres,
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name:     "test for wrong input",
			urlPath:  "/v1/movies/1",
			Title:    validTitle,
			Year:     validYear,
			Runtime:  validRuntime,
			Genres:   validGenres,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "String ID",
			urlPath:  "/v1/movies/string",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "database falls",
			urlPath:  "/v1/movies/2",
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "get not found",
			urlPath:  "/v1/movies/4",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "conflict err",
			urlPath:  "/v1/movies/1",
			Title:    conflictTitle,
			Year:     validYear,
			Runtime:  validRuntime,
			Genres:   validGenres,
			wantCode: http.StatusConflict,
		},
		{
			name:     "database falls",
			urlPath:  "/v1/movies/1",
			Title:    "fall database",
			Year:     validYear,
			Runtime:  validRuntime,
			Genres:   validGenres,
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "Fake json.Write",
			urlPath:  "/v1/movies/1",
			Title:    validTitle,
			Year:     validYear,
			Runtime:  validRuntime,
			Genres:   validGenres,
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Fake json.Write" {
				storedMarshal := jsonMarshal
				jsonMarshal = ts.fakeMarshal
				defer ts.restoreMarshal(storedMarshal)
			}

			inputData := struct {
				Title   string   `json:"title"`
				Year    int32    `json:"year"`
				Runtime string   `json:"runtime"`
				Genres  []string `json:"genres"`
			}{
				Title:   tt.Title,
				Year:    tt.Year,
				Runtime: tt.Runtime,
				Genres:  tt.Genres,
			}

			b, err := json.Marshal(&inputData)
			if err != nil {
				t.Fatal("wrong input data")
			}
			if tt.name == "test for wrong input" {
				b = append(b, 'a')
			}

			code, _, _ := ts.patchReq(t, tt.urlPath, b)

			assert.Equal(t, code, tt.wantCode)

		})
	}
}

func TestDeleteMovie(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routesTest())
	defer ts.Close()

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody string
	}{
		{
			name:     "deleting existing movie",
			urlPath:  "/v1/movies/1",
			wantCode: http.StatusOK,
		},
		{
			name:     "Non-existent ID",
			urlPath:  "/v1/movies/3",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "String ID",
			urlPath:  "/v1/movies/string",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Database fall",
			urlPath:  "/v1/movies/2",
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "Fake json.Write",
			urlPath:  "/v1/movies/1",
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "Fake json.Write" {
				storedMarshal := jsonMarshal
				jsonMarshal = ts.fakeMarshal
				defer ts.restoreMarshal(storedMarshal)
			}

			code, _, body := ts.deleteReq(t, tt.urlPath)

			assert.Equal(t, code, tt.wantCode)

			if tt.wantBody != "" {
				assert.StringContains(t, body, tt.wantBody)
			}

		})
	}

}

func TestListMovies(t *testing.T) {
	app := newTestApplication(t)

	ts := newTestServer(t, app.routesTest())
	defer ts.Close()

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody string
	}{
		{
			name:     "Valid req",
			urlPath:  "/v1/movies",
			wantCode: http.StatusOK,
		},
		{
			name:     "String page",
			urlPath:  "/v1/movies?page=string",
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name:     "Negative page",
			urlPath:  "/v1/movies?page=-1",
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name:     "Database fall cause of sort by genres",
			urlPath:  "/v1/movies?sort=title",
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "Fake json.Write",
			urlPath:  "/v1/movies",
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "Fake json.Write" {
				storedMarshal := jsonMarshal
				jsonMarshal = ts.fakeMarshal
				defer ts.restoreMarshal(storedMarshal)
			}

			code, _, body := ts.get(t, tt.urlPath)

			assert.Equal(t, code, tt.wantCode)

			if tt.wantBody != "" {
				assert.StringContains(t, body, tt.wantBody)
			}

		})
	}
}
