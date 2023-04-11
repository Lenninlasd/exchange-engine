package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/lenninlasd/Engine/types"
)

func TestGetAuth0UserFromToken(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer mock_access_token" {
			t.Errorf("Expected Authorization header 'Bearer mock_access_token', got %s", authHeader)
		}

		responseBody := `{
			"sub": "mock_sub",
			"email": "mockuser@example.com",
			"email_verified": true
		}`
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, responseBody)
	}))
	defer server.Close()

	os.Setenv("AUTH0_PROTOCOL", "http")
	os.Setenv("AUTH0_DOMAIN", strings.TrimPrefix(server.URL, "http://"))

	auth0User, err := getAuth0UserFromToken("mock_access_token")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check that the correct Auth0User was returned
	expectedUser := types.Auth0User{
		Sub:           "mock_sub",
		Email:         "mockuser@example.com",
		EmailVerified: true,
	}
	if !reflect.DeepEqual(auth0User, expectedUser) {
		t.Errorf("Expected Auth0User %+v, got %+v", expectedUser, auth0User)
	}
}

func TestGetUserFromHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer mock_access_token" {
			t.Errorf("Expected Authorization header 'Bearer mock_access_token', got %s", authHeader)
		}

		responseBody := `{
			"sub": "mock_sub",
			"email": "mockuser@example.com",
			"email_verified": true
		}`
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, responseBody)
	}))
	defer server.Close()

	os.Setenv("AUTH0_PROTOCOL", "http")
	os.Setenv("AUTH0_DOMAIN", strings.TrimPrefix(server.URL, "http://"))

	req := &http.Request{
		Header: http.Header{"Authorization": []string{"Bearer mock_access_token"}},
	}

	auth0User, err := GetUserFromHeader(req)

	// Assert that the returned auth0User matches the expected value
	expectedAuth0User := types.Auth0User{
		Sub:           "mock_sub",
		Email:         "mockuser@example.com",
		EmailVerified: true,
	}
	if !reflect.DeepEqual(expectedAuth0User, auth0User) {
		t.Errorf("Expected auth0User to be %+v, but got %+v", expectedAuth0User, auth0User)
	}

	// Assert that no error was returned
	if err != nil {
		t.Errorf("Expected no error, but got %s", err.Error())
	}

	// Mock request without an access token
	reqWithoutToken := &http.Request{}

	// Call GetUserFromHeader with the mock request
	auth0UserWithoutToken, err := GetUserFromHeader(reqWithoutToken)

	// Assert that an error was returned and the auth0User is empty
	if err == nil {
		t.Errorf("Expected an error, but got nil")
	}
	if !reflect.DeepEqual(types.Auth0User{}, auth0UserWithoutToken) {
		t.Errorf("Expected auth0UserWithoutToken to be empty, but got %+v", auth0UserWithoutToken)
	}
	expectedError := errors.New("Authorization header missing")
	if err.Error() != expectedError.Error() {
		t.Errorf("Expected error message to be '%s', but got '%s'", expectedError.Error(), err.Error())
	}
}
