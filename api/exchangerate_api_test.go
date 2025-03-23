package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock ExchangeRate function
func mockExchangeRateSuccess(repo ExchangeRateProvider) (uint64, error) {
	return 200000000, nil // Simulated BTC/COP price
}

func mockExchangeRateFail(repo ExchangeRateProvider) (uint64, error) {
	err := errors.New("failed to fetch exchange rate")
	return 0, err
}

func MockValidateTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r) // Just pass through without validation
	})
}

func TestExchangeRateApi_Success(t *testing.T) {
	// Inject mock
	exchangerateFunc = mockExchangeRateSuccess
	defer func() { exchangerateFunc = ExchangeRate }() // Reset after test

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

func TestExchangeRateApi_Fail(t *testing.T) {
	// Inject mock
	exchangerateFunc = mockExchangeRateFail
	defer func() { exchangerateFunc = ExchangeRate }() // Reset after test

	router := Routes(MockValidateTokenMiddleware)

	req := httptest.NewRequest("GET", "/supa/exchangerate/BTCCOP", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}
