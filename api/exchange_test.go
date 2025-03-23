package api

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockExchangeRateService struct {
	mock.Mock
}

func (m *MockExchangeRateService) GetBinanceBTCUSDT() (float64, error) {
	args := m.Called()
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockExchangeRateService) GetOpenExchangeRates() (ExchangeRateData, error) {
	args := m.Called()
	return args.Get(0).(ExchangeRateData), args.Error(1)
}

func (m *MockExchangeRateService) GetZenerDepositSpread(db *sql.DB) (float64, error) {
	args := m.Called(db)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockExchangeRateService) GetUSDCOPPriceFromDB(db *sql.DB) (float64, time.Time, error) {
	args := m.Called(db)
	return args.Get(0).(float64), args.Get(1).(time.Time), args.Error(2)
}

func (m *MockExchangeRateService) UpdateUSDCOPPriceInDB(db *sql.DB, price float64, updatedAt time.Time) error {
	args := m.Called(db, price, updatedAt)
	return args.Error(0)
}

func (m *MockExchangeRateService) OpenConnection() *sql.DB {
	args := m.Called()
	return args.Get(0).(*sql.DB)
}

func CreateMockExchangeRateService(mockDB *sql.DB) *MockExchangeRateService {
	service := new(MockExchangeRateService)
	service.On("OpenConnection").Return(mockDB)
	service.On("GetBinanceBTCUSDT").Return(60000.0, nil)
	service.On("GetZenerDepositSpread", mockDB).Return(0.02, nil)
	service.On("GetOpenExchangeRates").Return(ExchangeRateData{
		Rates: map[string]float64{"COP": 4010.0},
	}, nil)
	service.On("UpdateUSDCOPPriceInDB", mockDB, mock.AnythingOfType("float64"), mock.AnythingOfType("time.Time")).Return(nil)
	return service
}

func TestExchangeRate_ReturnsPriceFromOpenExchangeRates(t *testing.T) {
	// Mock database
	mockDB, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	// Mock service
	service := CreateMockExchangeRateService(mockDB)
	service.On("GetUSDCOPPriceFromDB", mockDB).Return(4000.0, time.Now().Add(-time.Hour), nil)

	expectedPrice := 24541200000 // 4010 * 60000 * (1 + 0.02) * 100

	// Run test
	result, err := ExchangeRate(service)
	assert.NoError(t, err)
	assert.Equal(t, uint64(expectedPrice), result)

	// if price from database is zero, it should return the price from OpenExchangeRates
	service.On("GetUSDCOPPriceFromDB", mockDB).Return(0, time.Now(), nil)
	result, err = ExchangeRate(service)
	assert.NoError(t, err)
	assert.Equal(t, uint64(expectedPrice), result)
}

func TestExchangeRate_ReturnsPriceFromDatabase(t *testing.T) {
	// Mock database
	mockDB, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	// Mock service
	service := CreateMockExchangeRateService(mockDB)
	service.On("GetUSDCOPPriceFromDB", mockDB).Return(4000.0, time.Now().Add((1-TIME_FETCH_RATES)*time.Minute), nil)

	expectedPrice := 24480000000 // 4000 * 60000 * (1 + 0.02) * 100

	// Run test
	result, err := ExchangeRate(service)
	assert.NoError(t, err)
	assert.Equal(t, uint64(expectedPrice), result)
}

func TestExchangeRate_Fails(t *testing.T) {
	// Mock database
	mockDB, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	// Mock service
	service := CreateMockExchangeRateService(mockDB)
	service.On("GetUSDCOPPriceFromDB", mockDB).Return(0.0, time.Now(), errors.New("failed to fetch Binance BTCUSDT"))

	// Run test
	result, err := ExchangeRate(service)
	assert.Error(t, err)
	assert.Equal(t, uint64(0), result)
}
