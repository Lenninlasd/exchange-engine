package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
	"github.com/lenninlasd/Engine/types"
)

const (
	WITHDRAW_STATUS_INICIALIZED = "initialized"
	WITHDRAW_STATUS_BROADCASTED = "broadcasted"
	WITHDRAW_STATUS_CONFIRMED   = "confirmed"
	WITHDRAW_STATUS_CANCELED    = "canceled"
)

func Withdraw(w http.ResponseWriter, r *http.Request) {
	var withdraw types.Withdraw

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "wrong data")
	}

	withdraw.FromJSON(reqBody)

	balance, err := GetUserBalance(withdraw.UserId)

	if err != nil {
		if err == sql.ErrNoRows {
			userNotExist(w)
			return
		}
		panic(err)
	}

	if withdraw.Amount > balance.Amount {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", "User does not have sufficient funds")
		return
	}

	withdrawId, err := initWithdrawTx(withdraw)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	// activate worker to check the mempool every minute if tx is confirmed
	transaction, err := SendWithdrawToWallet(withdraw)
	if err != nil {
		cancelWithdrawalTx(withdrawId)
		panic(err)
	}

	err = broadcastedWithdrawTx(withdrawId, transaction.Txid)
	if err != nil {
		cancelWithdrawalTx(withdrawId)
		panic(err)
	}

	Conv, _ := json.MarshalIndent(transaction, "", " ")
	fmt.Fprintf(w, "%s", string(Conv))
}

func initWithdrawTx(withdraw types.Withdraw) (uuid.UUID, error) {
	db := OpenConnection()

	withdrawId := uuid.New()

	_, err := db.Exec(
		`INSERT INTO withdraw_history (id, user_id, amount, address, status) VALUES ($1, $2, $3, $4, $5)`,
		withdrawId, withdraw.UserId, withdraw.Amount, withdraw.Address, WITHDRAW_STATUS_INICIALIZED,
	)

	defer db.Close()

	return withdrawId, err
}

func broadcastedWithdrawTx(withdrawId uuid.UUID, txid string) error {
	db := OpenConnection()

	_, err := db.Exec(
		`UPDATE withdraw_history SET status=$1, txid=$2, confirmations=0 WHERE id=$3`,
		WITHDRAW_STATUS_BROADCASTED, txid, withdrawId,
	)

	defer db.Close()

	return err
}

func cancelWithdrawalTx(withdrawId uuid.UUID) error {
	db := OpenConnection()
	_, err := db.Exec(
		`UPDATE withdraw_history SET status=$1 WHERE id=$2`,
		WITHDRAW_STATUS_CANCELED, withdrawId,
	)
	defer db.Close()
	return err
}

func SendWithdrawToWallet(withdraw types.Withdraw) (types.Transaction, error) {
	// create the request body
	requestBody, err := json.Marshal(map[string]interface{}{
		"amount":  withdraw.Amount,
		"address": withdraw.Address,
	})

	request, err := http.NewRequest("POST", "http://wallet:8000/withdraw", bytes.NewBuffer(requestBody))
	if err != nil {
		return types.Transaction{}, err
	}

	// set the Content-Type header
	request.Header.Add("Content-Type", "application/json")

	// create an HTTP client
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return types.Transaction{}, err
	}

	transaction := &types.Transaction{}
	err = json.NewDecoder(response.Body).Decode(transaction)
	if err != nil {
		return types.Transaction{}, err
	}

	defer response.Body.Close()

	return *transaction, err
}
