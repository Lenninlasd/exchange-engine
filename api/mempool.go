package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lenninlasd/Engine/pkg/utils"
)

func GetTransactionFee(w http.ResponseWriter, r *http.Request) {
	address := mux.Vars(r)["address"]

	recomendedFees, err := utils.GetMempoolFeeRates()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", "Error loading fees")
		return
	}

	fee, err := utils.GetFeeByFeerate(address, recomendedFees.HalfHourFee)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("Error loading fees: %v\n", err)
		return
	}

	// send the fee rate as well
	fmt.Fprintf(w, "%d", fee)
}
