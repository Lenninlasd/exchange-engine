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

const TIME_FECHT_RATES = 45

type ExchangeRate struct {
	Disclaimer string             `json:"disclaimer"`
	License    string             `json:"license"`
	Timestamp  int64              `json:"timestamp"`
	Base       string             `json:"base"`
	Rates      map[string]float64 `json:"rates"`
}

func (rate *ExchangeRate) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, rate)
}

type BinancePrice struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

func (price *BinancePrice) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, price)
}

func getbinanceBTCUSDT() (float64, error) {

	var binancePrice BinancePrice
	var price float64

	resp, err := http.Get("https://api.binance.us/api/v3/ticker/price?symbol=BTCUSDT")
	if err != nil {
		return price, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return price, err
	}

	err = binancePrice.FromJSON(body)
	if err != nil {
		return price, err
	}

	price, err = strconv.ParseFloat(binancePrice.Price, 64)
	if err == nil && price == 0 {
		return price, errors.New("error on getbinanceBTCUSDT: BTCUSDT price can not be zero")
	}

	return price, err
}

func getOpenexchangerates() (ExchangeRate, error) {
	var rate ExchangeRate
	appId := os.Getenv("OPENEXCHANGE_APP_ID")

	req, err := http.NewRequest("GET", fmt.Sprintf("https://openexchangerates.org/api/latest.json?app_id=%s&symbols=COP,", appId), nil)
	if err != nil {
		return rate, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Token %s", appId))

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

	err = rate.FromJSON(body)
	if err != nil {
		return rate, err
	}

	return rate, nil
}

func getZenerDepositSpread() (float64, error) {
	db := OpenConnection()
	defer db.Close()
	var zenerDepositSpread float64

	err := db.QueryRow("SELECT config_value FROM public.app_config WHERE config_key='ZENER_DEPOSIT_SPREAD'").Scan(&zenerDepositSpread)
	if err != nil {
		return 0, err
	}

	return zenerDepositSpread, err
}

// TODO: make unit test
func Exchangerate() (uint64, error) {
	db := OpenConnection()
	defer db.Close()

	var currentPrice float64
	var USDCOP_price float64
	var uptatedAt time.Time

	err := db.QueryRow("SELECT price, updatedat FROM marketprice WHERE symbol='USDCOP'").Scan(&currentPrice, &uptatedAt)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	currentTime := time.Now()
	difference := currentTime.Sub(uptatedAt)

	if difference >= time.Minute*TIME_FECHT_RATES || currentPrice == 0 {
		rate, err := getOpenexchangerates()
		if err != nil {
			return 0, err
		}

		USDCOP_price = rate.Rates["COP"]

		_, err = db.Exec(`INSERT INTO marketprice(symbol, price, updatedat) VALUES ('USDCOP', $1, $2)
			ON CONFLICT (symbol) DO UPDATE SET price = EXCLUDED.price, updatedat = EXCLUDED.updatedat`,
			USDCOP_price, currentTime,
		)
		if err != nil {
			return 0, err
		}
	} else {
		USDCOP_price = currentPrice
	}

	BTCUSDT_price, err := getbinanceBTCUSDT()
	if err != nil {
		return 0, err
	}

	zenerDepositSpread, err := getZenerDepositSpread()

	if err != nil {
		return 0, err
	}

	priceAfterFee := USDCOP_price * BTCUSDT_price * (1 + zenerDepositSpread)

	return uint64(priceAfterFee * 100), nil
}

func ExchangerateApi(w http.ResponseWriter, r *http.Request) {
	price, err := Exchangerate()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", "Error loading rates")
		panic(err)
	}

	fmt.Fprintf(w, "%d", price)
}
