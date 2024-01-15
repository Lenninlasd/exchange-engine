package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/lenninlasd/Engine/constants"
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

func GetFeeByFeerate(address string, feeRate float32) (uint64, error) {
	if address == "" {
		address = "tb1q30grxr89ln5hmrxecr6d0l4vxsffv5q2arwcu3" // dummy address
	}

	var fee uint64

	url := fmt.Sprintf("%s/transactiondetails?address=%s&amount=%d&fee_rate=%f", constants.WALLET_ADDRESS, address, 600, feeRate)
	resp, err := http.Get(url)
	if err != nil {
		return fee, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fee, err
	}

	var txDetails types.Transaction
	txDetails.FromJSON(body)

	fee = txDetails.Fee + constants.ZENER_FEE
	return fee, nil
}
