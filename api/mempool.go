package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/lenninlasd/Engine/pkg/utils"
)

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

	recomendedFees, err := utils.GetMempoolFeeRates()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", "Error loading fees")
		return
	}

	fee, err := utils.GetFeeByFeerate(address, recomendedFees.FastestFee)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("Error loading fees: %v\n", err)
		return
	}

	// send the fee rate as well
	fmt.Fprintf(w, "%d", fee)
}
