package api

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockExchangeRateRepo struct {
	mock.Mock
}

func (m *MockExchangeRateRepo) GetBinanceBTCUSDT() (float64, error) {
	args := m.Called()
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockExchangeRateRepo) GetOpenExchangeRates() (ExchangeRate, error) {
	args := m.Called()
	return args.Get(0).(ExchangeRate), args.Error(1)
}

func (m *MockExchangeRateRepo) GetZenerDepositSpread(db *sql.DB) (float64, error) {
	args := m.Called(db)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockExchangeRateRepo) GetUSDCOPPriceFromDB(db *sql.DB) (float64, time.Time, error) {
	args := m.Called(db)
	return args.Get(0).(float64), args.Get(1).(time.Time), args.Error(2)
}

func (m *MockExchangeRateRepo) UpdateUSDCOPPriceInDB(db *sql.DB, price float64, updatedAt time.Time) error {
	args := m.Called(db, price, updatedAt)
	return args.Error(0)
}

func (m *MockExchangeRateRepo) OpenConnection() *sql.DB {
	args := m.Called()
	return args.Get(0).(*sql.DB)
}

func TestExchangerate(t *testing.T) {
	// Mock database
	mockDB, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	// Mock repo
	repo := new(MockExchangeRateRepo)
	repo.On("OpenConnection").Return(mockDB)
	repo.On("GetUSDCOPPriceFromDB", mockDB).Return(4000.0, time.Now().Add(-time.Hour), nil)
	repo.On("GetBinanceBTCUSDT").Return(60000.0, nil)
	repo.On("GetZenerDepositSpread", mockDB).Return(0.02, nil)
	repo.On("GetOpenExchangeRates").Return(ExchangeRate{
		Rates: map[string]float64{"COP": 4000.0},
	}, nil)
	repo.On("UpdateUSDCOPPriceInDB", mockDB, mock.AnythingOfType("float64"), mock.AnythingOfType("time.Time")).Return(nil)

	// Run test
	result, err := Exchangerate(repo)
	assert.NoError(t, err)
	assert.Greater(t, result, uint64(0))
}
