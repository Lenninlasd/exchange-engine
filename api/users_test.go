package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHomeLink(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Call the homeLink function directly
	homeLink(rr, req)

	// Check if the status code is 200 OK
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	// Check if the response body is correct
	expectedBody := "Restful API using Go and Postgres! from new version"
	if rr.Body.String() != expectedBody {
		t.Errorf("Expected response body %q, got %q", expectedBody, rr.Body.String())
	}
}
