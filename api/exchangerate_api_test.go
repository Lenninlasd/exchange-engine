package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock Exchangerate function
func mockExchangerateSuccess(repo *ExchangeRateRepo) (uint64, error) {
	return 200000000, nil // Simulated BTC/COP price
}

func mockExchangerateFail(repo *ExchangeRateRepo) (uint64, error) {
	err := errors.New("failed to fetch exchange rate")
	return 0, err
}

func MockValidateTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r) // Just pass through without validation
	})
}

func TestExchangerateApiSuccess(t *testing.T) {
	// Inject mock
	exchangerateFunc = mockExchangerateSuccess
	defer func() { exchangerateFunc = Exchangerate }() // Reset after test

	router := Routes(MockValidateTokenMiddleware) // Get router

	req := httptest.NewRequest("GET", "/supa/exchangerate/BTCCOP", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}

	expectedBody := "200000000"
	if rec.Body.String() != expectedBody {
		t.Errorf("Expected response %s, got %s", expectedBody, rec.Body.String())
	}
}

func TestExchangerateApiFail(t *testing.T) {
	// Inject mock
	exchangerateFunc = mockExchangerateFail
	defer func() { exchangerateFunc = Exchangerate }() // Reset after test

	router := Routes(MockValidateTokenMiddleware)

	req := httptest.NewRequest("GET", "/supa/exchangerate/BTCCOP", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}
