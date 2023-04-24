package main

import (
	"encoding/json"
	"greenlight.bcc/internal/assert"
	"net/http"
	"testing"
)

func TestRegisterUser(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routesTest())
	defer ts.Close()

	const (
		validName     = "name"
		validEmail    = "email@gmail.com"
		validPassword = "password"
	)
	tests := []struct {
		name     string
		Name     string
		Email    string
		Password string
		Mock     string
		wantCode int
	}{
		{
			name:     "Valid submission",
			Name:     validName,
			Email:    validEmail,
			Password: validPassword,
			wantCode: http.StatusCreated,
		},
		{
			name:     "Empty Name",
			Name:     "",
			Email:    validEmail,
			Password: validPassword,
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name:     "Invalid request",
			Name:     validName,
			Email:    validEmail,
			Password: validPassword,
			Mock:     "mock",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "Too long password",
			Name:     validName,
			Email:    validEmail,
			Password: "0000000000000000000000000000000000000000000000000000000000000000000000000000000001",
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "Invalid name",
			Name:     "invalid",
			Email:    validEmail,
			Password: validPassword,
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "Invalid email",
			Name:     validName,
			Email:    "invalid@gmail.com",
			Password: validPassword,
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name:     "Permission fall",
			Name:     "permissions fall",
			Email:    validEmail,
			Password: validPassword,
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "Token fall",
			Name:     "token fall",
			Email:    validEmail,
			Password: validPassword,
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "Fake json.Write",
			Name:     validName,
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
				Name     string `json:"name"`
				Email    string `json:"email"`
				Password string `json:"password"`
				Mock     string `json:"mock,omitempty"`
			}{
				Name:     tt.Name,
				Email:    tt.Email,
				Password: tt.Password,
			}
			if tt.name == "Invalid request" {
				inputData.Mock = tt.Mock
			}

			b, err := json.Marshal(&inputData)
			if err != nil {
				t.Fatal("wrong input data")
			}
			if tt.name == "test for wrong input" {
				b = append(b, 'a')
			}

			code, _, _ := ts.postForm(t, "/v1/users", b)

			assert.Equal(t, code, tt.wantCode)

		})
	}
}

func TestActivateUser(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routesTest())
	const (
		validToken             = "TokenPlainTextForTokenTest"
		failedValidationToken  = "TokenPlainTextForTokenTes1"
		internalServerErrToken = "TokenPlainTextForTokenTes2"
		emailConflictToken     = "TokenPlainTextForTokenTes3"
		emailErrToken          = "TokenPlainTextForTokenTes4"
		deleteAllForTokenError = "TokenPlainTextForTokenTes5"
	)

	tests := []struct {
		name     string
		Token    string
		Mock     string
		wantCode int
	}{
		{
			name:     "Valid input",
			Token:    validToken,
			wantCode: http.StatusOK,
		},
		{
			name:     "Wrong input",
			Token:    validToken,
			Mock:     "mock",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "Invalid token",
			Token:    "invalid_token",
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name:     "Token not found",
			Token:    failedValidationToken,
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name:     "Token, database fall",
			Token:    internalServerErrToken,
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "Email conflict",
			Token:    emailConflictToken,
			wantCode: http.StatusConflict,
		},
		{
			name:     "Email error",
			Token:    emailErrToken,
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "Delete all for Token error",
			Token:    deleteAllForTokenError,
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "Fake json.Write",
			Token:    validToken,
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputData := struct {
				Token string `json:"token"`
				Mock  string `json:"mock,omitempty"`
			}{
				Token: tt.Token,
			}

			if tt.name == "Fake json.Write" {
				storedMarshal := jsonMarshal
				jsonMarshal = ts.fakeMarshal
				defer ts.restoreMarshal(storedMarshal)
			}

			if tt.Mock == "mock" {
				inputData.Mock = tt.Mock
			}

			b, err := json.Marshal(&inputData)
			if err != nil {
				t.Fatal("wrong input data")
			}
			if tt.name == "test for wrong input" {
				b = append(b, 'a')
			}

			code, _, _ := ts.updateReq(t, "/v1/users/activated", b)

			assert.Equal(t, code, tt.wantCode)

		})
	}
}
