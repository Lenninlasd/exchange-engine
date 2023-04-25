package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/lenninlasd/Engine/types"
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

func CreateFiatDeposit(w http.ResponseWriter, r *http.Request) {
	auth0User, err := GetUserFromHeader(r)
	if err != nil {
		panic(err)
	}

	var depositTx types.DepositTx

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "wrong data")
	}

	depositTx.FromJSON(reqBody)

	db := OpenConnection()
	_, err = db.Exec(`
		INSERT INTO fiat_deposit_tx (reference, amount_in_cents, status, user_id)
		VALUES ($1, $2, $3, $4)`, depositTx.Reference, depositTx.AmountInCents, "INITIATE", auth0User.Sub)

	defer db.Close()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
}
