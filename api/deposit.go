package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/google/uuid"
)

type DepositFiatTx struct {
	Amount   uint64 `json:"amount"`
	Currency string `json:"currency"`
}

type RefResponse struct {
	Reference string `json:"reference"`
	Integrity string `json:"integrity"`
}

func (depositFiatTx *DepositFiatTx) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, depositFiatTx)
}

// "<Referencia><Monto><Moneda><SecretoIntegridad>"
func GetDepositReference(w http.ResponseWriter, r *http.Request) {
	var depositFiatTx DepositFiatTx

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "wrong data")
	}

	depositFiatTx.FromJSON(reqBody)

	reference := uuid.New().String()

	integrity := os.Getenv("WOMPY_INTEGRITY")

	strToken := reference + fmt.Sprint(depositFiatTx.Amount) + depositFiatTx.Currency + integrity

	hash := sha256.Sum256([]byte(strToken))
	hashToken := hex.EncodeToString(hash[:])

	response := RefResponse{
		Reference: reference,
		Integrity: hashToken,
	}

	Conv, _ := json.MarshalIndent(response, "", " ")
	fmt.Fprintf(w, "%s", Conv)
}
