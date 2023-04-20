package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/google/uuid"
)

const MIN_DEPOSIT = 100000

type RefResponse struct {
	Reference string `json:"reference"`
	Integrity string `json:"integrity"`
}

// "<Referencia><Monto><Moneda><SecretoIntegridad>"
func GetDepositReference(w http.ResponseWriter, r *http.Request) {
	amountStr := r.URL.Query().Get("amount")
	currency := r.URL.Query().Get("currency")

	amount, err := strconv.ParseUint(amountStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid 'amount' parameter: "+err.Error(), http.StatusBadRequest)
		return
	}

	if amount < MIN_DEPOSIT {
		http.Error(w, fmt.Sprintf("Deposit amount must be at least %d", MIN_DEPOSIT), http.StatusBadRequest)
		return
	}

	reference := uuid.New().String()
	integrity := os.Getenv("WOMPY_INTEGRITY")
	strToken := reference + amountStr + currency + integrity

	hash := sha256.Sum256([]byte(strToken))
	hashToken := hex.EncodeToString(hash[:])

	response := RefResponse{
		Reference: reference,
		Integrity: hashToken,
	}

	Conv, _ := json.MarshalIndent(response, "", " ")
	fmt.Fprintf(w, "%s", Conv)
}
