package api

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/lenninlasd/Engine/middleware"
	"github.com/lenninlasd/Engine/types"
)

const (
	PENDING  = "PENDING"
	APPROVED = "APPROVED"
	DECLINED = "DECLINED"
	ERROR    = "ERROR"
	VOIDED   = "VOIDED"

	SATS_PER_BTC = 100000000
	MIN_DEPOSIT  = 100000
)

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
	integrity := os.Getenv("WOMPI_INTEGRITY")
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

func CreateFiatDepositSupa(w http.ResponseWriter, r *http.Request) {
	userClaims, err := middleware.GetClainsFromToken(r.Header)
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
		VALUES ($1, $2, $3, $4)`, depositTx.Reference, depositTx.AmountInCents, PENDING, userClaims.Subject)

	defer db.Close()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
}

func CompareChecksum(wompiEvent types.WompiTransactionUpdatedEvent, AmountInCents uint64) error {
	wompiEventToken := os.Getenv("WOMPI_EVENT")
	amountInCents := strconv.FormatUint(AmountInCents, 10)
	Timestamp := strconv.FormatInt(wompiEvent.Timestamp, 10)

	// * Calculate checksum with the AmountInCents from fiat_deposit_tx that was previosly created
	checksumString := wompiEvent.Data.Transaction.ID + wompiEvent.Data.Transaction.Status + amountInCents + Timestamp + wompiEventToken
	hash := sha256.Sum256([]byte(checksumString))
	checksum := hex.EncodeToString(hash[:])

	if checksum != wompiEvent.Signature.Checksum {
		return errors.New("Checksum mismatch")
	}
	return nil
}

func GetFiatDepositData(reference string) (types.FiatDepositTx, error) {
	db := OpenConnection()
	var fiatDepositTx types.FiatDepositTx

	err := db.QueryRow("SELECT user_id, amount_in_cents, status  FROM fiat_deposit_tx WHERE reference=$1", reference).Scan(&fiatDepositTx.UserId, &fiatDepositTx.AmountInCents, &fiatDepositTx.Status)

	defer db.Close()

	return fiatDepositTx, err
}

func UpdateBalanceAfterDeposit(
	UserId string,
	Transaction types.WompiTx,
	btcCurrentPriceInCents uint64,
	balanceAmount uint64,
) error {
	db := OpenConnection()
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	purchasedSats := SATS_PER_BTC * Transaction.AmountInCents / btcCurrentPriceInCents

	result, err := tx.Exec(
		`UPDATE fiat_deposit_tx SET id=$1, status=$2, purchased_sats=$3 WHERE reference=$4 AND status = $5`,
		Transaction.ID, Transaction.Status, purchasedSats, Transaction.Reference, PENDING,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	numRowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if numRowsAffected != 1 {
		tx.Rollback()
		return errors.New("UpdateBalanceAfterDeposit wrong event: no rows affected")
	}

	newBalanceAmountInSats := balanceAmount + purchasedSats

	result, err = tx.Exec(`UPDATE btcbalances SET amount=$1 WHERE user_id=$2;`, newBalanceAmountInSats, UserId)
	if err != nil {
		tx.Rollback()
		return err
	}

	numRowsAffected, err = result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if numRowsAffected != 1 {
		tx.Rollback()
		return errors.New("UpdateBalanceAfterDeposit wrong event: No rows in btc balance affected")
	}

	return tx.Commit()
}

func UpdateFailedFiatDeposit(Transaction types.WompiTx) error {
	db := OpenConnection()
	defer db.Close()

	_, err := db.Exec(
		`UPDATE fiat_deposit_tx SET id=$1, status=$2 WHERE reference=$3 AND status = $4`,
		Transaction.ID, Transaction.Status, Transaction.Reference, PENDING,
	)
	if err != nil {
		return err
	}

	return errors.New("UpdateBalanceAfterDeposit error on transaction: No Approved")
}

func UptateFiatDeposit(w http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "wrong data")
	}

	var wompiEvent types.WompiTransactionUpdatedEvent
	wompiEvent.FromJSON(reqBody)

	fiatDepositTx, err := GetFiatDepositData(wompiEvent.Data.Transaction.Reference)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", err)
		return
	}

	err = CompareChecksum(wompiEvent, fiatDepositTx.AmountInCents)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", err)
		return
	}

	balance, err := GetUserBalance(fiatDepositTx.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			userNotExist(w)
			return
		}
		panic(err)
	}

	// the wompi event is Decline or Error
	if wompiEvent.Data.Transaction.Status != APPROVED {
		err = UpdateFailedFiatDeposit(wompiEvent.Data.Transaction)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
		return
	}

	btcCurrentPriceInCents, err := Exchangerate()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", "Error loading rates")
		panic(err)
	}

	err = UpdateBalanceAfterDeposit(fiatDepositTx.UserId, wompiEvent.Data.Transaction, btcCurrentPriceInCents, balance.Amount)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
	}

	w.WriteHeader(http.StatusOK)
}
