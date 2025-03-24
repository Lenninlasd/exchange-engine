package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

const TIME_FETCH_RATES = 45

type ExchangeRateData struct {
	Disclaimer string             `json:"disclaimer"`
	License    string             `json:"license"`
	Timestamp  int64              `json:"timestamp"`
	Base       string             `json:"base"`
	Rates      map[string]float64 `json:"rates"`
}

func (rate *ExchangeRateData) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, rate)
}

type BinancePrice struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

func (price *BinancePrice) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, price)
}

type ExchangeRateProvider interface {
	GetBinanceBTCUSDT() (float64, error)
	GetOpenExchangeRates() (ExchangeRateData, error)
	GetZenerDepositSpread(db *sql.DB) (float64, error)
	GetUSDCOPPriceFromDB(db *sql.DB) (float64, time.Time, error)
	UpdateUSDCOPPriceInDB(db *sql.DB, price float64, updatedAt time.Time) error
	OpenConnection() *sql.DB
}

// ExchangeRateService contains methods for retrieving exchange rates.
type ExchangeRateService struct{}

// GetBinanceBTCUSDT fetches the BTC/USDT price from Binance API.
func (repo *ExchangeRateService) GetBinanceBTCUSDT() (float64, error) {
	var binancePrice BinancePrice
	resp, err := http.Get("https://api.binance.us/api/v3/ticker/price?symbol=BTCUSDT")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	err = json.Unmarshal(body, &binancePrice)
	if err != nil {
		return 0, err
	}

	price, err := strconv.ParseFloat(binancePrice.Price, 64)
	if err == nil && price == 0 {
		return 0, errors.New("BTCUSDT price cannot be zero")
	}

	return price, err
}

// GetOpenExchangeRates fetches USD/COP exchange rate from OpenExchangeRates API.
func (repo *ExchangeRateService) GetOpenExchangeRates() (ExchangeRateData, error) {
	var rate ExchangeRateData
	appID := os.Getenv("OPENEXCHANGE_APP_ID")

	req, err := http.NewRequest("GET", fmt.Sprintf("https://openexchangerates.org/api/latest.json?app_id=%s&symbols=COP", appID), nil)
	if err != nil {
		return rate, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Token %s", appID))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return rate, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return rate, err
	}

	err = json.Unmarshal(body, &rate)
	return rate, err
}

// GetZenerDepositSpread fetches the spread from the database.
func (repo *ExchangeRateService) GetZenerDepositSpread(db *sql.DB) (float64, error) {
	var spread float64
	err := db.QueryRow("SELECT config_value FROM public.app_config WHERE config_key='ZENER_DEPOSIT_SPREAD'").Scan(&spread)
	return spread, err
}

// GetUSDCOPPriceFromDB retrieves the last USDCOP price from the database.
func (repo *ExchangeRateService) GetUSDCOPPriceFromDB(db *sql.DB) (float64, time.Time, error) {
	var price float64
	var updatedAt time.Time
	err := db.QueryRow("SELECT price, updatedat FROM marketprice WHERE symbol='USDCOP'").Scan(&price, &updatedAt)
	return price, updatedAt, err
}

// UpdateUSDCOPPriceInDB updates or inserts the latest USDCOP price into the database.
func (repo *ExchangeRateService) UpdateUSDCOPPriceInDB(db *sql.DB, price float64, updatedAt time.Time) error {
	_, err := db.Exec(`INSERT INTO marketprice(symbol, price, updatedat) VALUES ('USDCOP', $1, $2)
		ON CONFLICT (symbol) DO UPDATE SET price = EXCLUDED.price, updatedat = EXCLUDED.updatedat`,
		price, updatedAt,
	)
	return err
}

func (repo *ExchangeRateService) OpenConnection() *sql.DB {
	return OpenConnection()
}

func ExchangeRate(repo ExchangeRateProvider) (uint64, error) {
	db := repo.OpenConnection()
	defer db.Close()

	currentPrice, updatedAt, err := repo.GetUSDCOPPriceFromDB(db)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	currentTime := time.Now()
	difference := currentTime.Sub(updatedAt)

	var usdCopPrice float64
	if difference >= time.Minute*TIME_FETCH_RATES || currentPrice == 0 {
		rate, err := repo.GetOpenExchangeRates()
		if err != nil {
			return 0, err
		}

		usdCopPrice = rate.Rates["COP"]

		err = repo.UpdateUSDCOPPriceInDB(db, usdCopPrice, currentTime)
		if err != nil {
			return 0, err
		}
	} else {
		usdCopPrice = currentPrice
	}

	btcUsdtPrice, err := repo.GetBinanceBTCUSDT()
	if err != nil {
		return 0, err
	}

	zenerDepositSpread, err := repo.GetZenerDepositSpread(db)
	if err != nil {
		return 0, err
	}

	priceAfterFee := usdCopPrice * btcUsdtPrice * (1 + zenerDepositSpread)

	return uint64(priceAfterFee * 100), nil
}

var exchangerateFunc = ExchangeRate // Default function
var exchangeRateRepo = &ExchangeRateService{}

func ExchangeRateApi(w http.ResponseWriter, r *http.Request) {
	price, err := exchangerateFunc(exchangeRateRepo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", "Error loading rates")
	}

	fmt.Fprintf(w, "%d", price)
}
