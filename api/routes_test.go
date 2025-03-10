package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHomeRoute(t *testing.T) {
	router := Routes(MockValidateTokenMiddleware) // Get the router

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	// Check if the response body is correct
	expectedBody := "Restful API using Go and Postgres! from new version"
	if rr.Body.String() != expectedBody {
		t.Errorf("Expected response body %q, got %q", expectedBody, rr.Body.String())
	}
}
