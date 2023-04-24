package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"greenlight.bcc/internal/assert"
)

func TestCreateToken(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routesTest())
	defer ts.Close()

	const (
		validEmail                      = "test0@test.com"
		validPassword                   = "TestPassword"
		emailInvalidCredentials         = "test1@test.com"
		emailInternalServerErr          = "test2@test.com"
		invalidPasswordEmail            = "test3@test.com"
		invalidCredentialsPasswordEmail = "test4@test.com"
		tokenErrorEmail                 = "test5@test.com"
	)

	tests := []struct {
		name     string
		Email    string
		Password string
		wantCode int
	}{
		{
			name:     "Create token",
			Email:    validEmail,
			Password: validPassword,
			wantCode: http.StatusCreated,
		},
		{
			name:     "InvalidCredentials after	GetByEmail",
			Email:    emailInvalidCredentials,
			Password: validPassword,
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "InternalServerError after GetByEmail",
			Email:    emailInternalServerErr,
			Password: validPassword,
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "Email validation fail",
			Email:    "invalid",
			Password: validPassword,
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name:     "Invalid password",
			Email:    invalidPasswordEmail,
			Password: validPassword,
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "InvalidCredentials password",
			Email:    invalidCredentialsPasswordEmail,
			Password: "asdasdasdasdasdasdasdasdasd",
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "token error",
			Email:    tokenErrorEmail,
			Password: validPassword,
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "Fake json.Write",
			Email:    validEmail,
			Password: validPassword,
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
				Email    string `json:"email"`
				Password string `json:"password"`
			}{
				Email:    tt.Email,
				Password: tt.Password,
			}

			b, err := json.Marshal(&inputData)
			if err != nil {
				t.Fatal("wrong input data")
			}
			if tt.name == "test for wrong input" {
				b = append(b, 'a')
			}

			code, _, _ := ts.postForm(t, "/v1/tokens/authentication", b)

			assert.Equal(t, code, tt.wantCode)
		})
	}

	code, _, _ := ts.postForm(t, "/v1/tokens/authentication", []byte{})

	assert.Equal(t, code, http.StatusBadRequest)
}
