package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/lenninlasd/Engine/types"
)

const (
	WITHDRAW_STATUS_INICIALIZED = "initialized"
	WITHDRAW_STATUS_BROADCASTED = "broadcasted"
	WITHDRAW_STATUS_CONFIRMED   = "confirmed"
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

	err = SendWithdraw(withdraw)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	// sent withdrawal signal the bdk wallet ✅
	// receive signal from wallet -> broadcasted
	// activate worker to check the mempool every minute if tx is confirmed

	transaction := SendWithdrawToWallet(withdraw)

	Conv, _ := json.MarshalIndent(transaction, "", " ")
	fmt.Fprintf(w, "%s", string(Conv))
}

func SendWithdraw(withdraw types.Withdraw) error {
	db := OpenConnection()
	_, err := db.Exec(
		`INSERT INTO withdraw_history (user_id, amount, address, status) VALUES ($1, $2, $3, $4)`,
		withdraw.UserId, withdraw.Amount, withdraw.Address, WITHDRAW_STATUS_INICIALIZED,
	)

	defer db.Close()

	return err
}

func SendWithdrawToWallet(withdraw types.Withdraw) types.Transaction {
	// create the request body
	requestBody, err := json.Marshal(map[string]interface{}{
		"amount":  withdraw.Amount,
		"address": withdraw.Address,
	})

	request, err := http.NewRequest("POST", "http://wallet:8000/withdraw", bytes.NewBuffer(requestBody))
	if err != nil {
		panic(err)
	}

	// set the Content-Type header
	request.Header.Add("Content-Type", "application/json")

	// create an HTTP client
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}

	defer response.Body.Close()

	transaction := &types.Transaction{}
	derr := json.NewDecoder(response.Body).Decode(transaction)
	if derr != nil {
		panic(derr)
	}

	return *transaction
}
