package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/lenninlasd/Engine/types"
)

type RecomendedFees struct {
	FastestFee  float32 `json:"fastestFee"`
	HalfHourFee float32 `json:"halfHourFee"`
	HourFee     float32 `json:"hourFee"`
	EconomyFee  float32 `json:"economyFee"`
	MinimumFee  float32 `json:"minimumFee"`
}

func (fees *RecomendedFees) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, fees)
}

func GetMempoolFeeRates() (RecomendedFees, error) {
	var recomendedFees RecomendedFees

	resp, err := http.Get("https://mempool.space/api/v1/fees/recommended")
	if err != nil {
		return recomendedFees, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return recomendedFees, err
	}

	err = recomendedFees.FromJSON(body)
	if err != nil {
		return recomendedFees, err
	}

	return recomendedFees, nil
}

func GetCurrentBlockHeight() (uint64, error) {
	resp, err := http.Get("https://mempool.space/testnet/api/blocks/tip/height")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	height, err := strconv.Atoi(string(body))
	if err != nil {
		return 0, err
	}

	return uint64(height), nil
}

func GetTransactionFee(w http.ResponseWriter, r *http.Request) {
	address := mux.Vars(r)["address"]
	if address == "" {
		address = "tb1q30grxr89ln5hmrxecr6d0l4vxsffv5q2arwcu3" // dummy address
	}

	recomendedFees, err := GetMempoolFeeRates()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", "Error loading fees")
		return
	}

	url := fmt.Sprintf("http://wallet:8000/transactiondetails?address=%s&amount=%d&fee_rate=%f", address, 600, recomendedFees.FastestFee)
	resp, err := http.Get(url)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("Error fetching transaction details: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return
	}

	var txDetails types.Transaction
	txDetails.FromJSON(body)

	// send the fee rate as well
	fmt.Fprintf(w, "%d", txDetails.Fee)
}
