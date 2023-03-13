package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/lenninlasd/Engine/types"
)

const (
	WITHDRAW_STATUS_INICIALIZED = "initialized"
	WITHDRAW_STATUS_BROADCASTED = "broadcasted"
	WITHDRAW_STATUS_CONFIRMED   = "confirmed"
	WITHDRAW_STATUS_CANCELED    = "canceled"

	MIM_NUMBER_CONFIRMATIONS = 3
)

func Withdraw(w http.ResponseWriter, r *http.Request) {
	var withdraw types.Withdraw

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "wrong data")
	}

	withdraw.FromJSON(reqBody)

	user, err := GetUserByEmail(withdraw.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			userNotExist(w)
			return
		}
		panic(err)
	}

	withdraw.UserId = user.Id

	isUnable := checkUserWithdrawalStatus(withdraw.UserId)
	if isUnable {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", "User is not able to withdraw")
		return
	}

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

func GetPendingWithdrawalTxByEmail(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()
	email := mux.Vars(r)["email"]
	var tx types.BroadcastedTx

	err := db.QueryRow(
		`SELECT txid, address, time_stamp, status, confirmations
		FROM users INNER JOIN withdraw_history ON users.id = withdraw_history.user_id
		WHERE users.email=$1 AND (status=$2 OR (status=$3 AND confirmations < $4))`,
		email, WITHDRAW_STATUS_BROADCASTED, WITHDRAW_STATUS_CONFIRMED, MIM_NUMBER_CONFIRMATIONS).Scan(
		&tx.Txid, &tx.Address, &tx.Timestamp, &tx.Status, &tx.Confirmations)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	Conv, _ := json.MarshalIndent(tx, "", " ")
	fmt.Fprintf(w, "%s", Conv)
}

func checkUserWithdrawalStatus(UserId uuid.UUID) bool {
	db := OpenConnection()
	defer db.Close()

	query := `SELECT EXISTS (SELECT 1 FROM withdraw_history WHERE user_id = $1 AND status IN ($2, $3))`

	var exists bool
	err := db.QueryRow(query, UserId, WITHDRAW_STATUS_INICIALIZED, WITHDRAW_STATUS_BROADCASTED).Scan(&exists)
	if err != nil {
		panic(err)
	}

	return exists
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
