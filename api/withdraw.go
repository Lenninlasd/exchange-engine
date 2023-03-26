package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/lenninlasd/Engine/types"
)

const (
	WITHDRAW_STATUS_INICIALIZED = "initialized"
	WITHDRAW_STATUS_BROADCASTED = "broadcasted"
	WITHDRAW_STATUS_CONFIRMED   = "confirmed"
	WITHDRAW_STATUS_CANCELED    = "canceled"

	MIM_NUMBER_CONFIRMATIONS        = 3
	TIME_TO_FECHT_MEMPOOL           = 1
	ZENER_FEE                uint64 = 500
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

	go watchTransactionOnMempool(transaction)

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
		email, WITHDRAW_STATUS_BROADCASTED, WITHDRAW_STATUS_CONFIRMED, MIM_NUMBER_CONFIRMATIONS+1).Scan(
		&tx.Txid, &tx.Address, &tx.Timestamp, &tx.Status, &tx.Confirmations)

	if err != nil && err != sql.ErrNoRows {
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

func updateTxConfirmations(txid string, confirmations uint64) error {
	db := OpenConnection()

	_, err := db.Exec(
		`UPDATE withdraw_history SET status=$1, confirmations=$2 WHERE txid=$3`,
		WITHDRAW_STATUS_CONFIRMED, confirmations, txid,
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

func getCurrentBlockHeight() (uint64, error) {
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

func watchTransactionOnMempool(transaction types.Transaction) {
	// TODO Implement websockets to watch new block confirmations
	ticker := time.NewTicker(TIME_TO_FECHT_MEMPOOL * time.Minute)
	done := make(chan bool)
	isConfirmed := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				ticker.Stop()
				return
			case <-ticker.C:
				currentBlockHeight, err := getCurrentBlockHeight()
				if err != nil {
					fmt.Printf("Error fetching current block Height: %v\n", err)
					continue
				}

				resp, err := http.Get(fmt.Sprintf("https://mempool.space/testnet/api/tx/%s", transaction.Txid))
				if err != nil {
					fmt.Printf("Error fetching transaction data: %v\n", err)
					continue
				}
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Printf("Error reading response body: %v\n", err)
					continue
				}

				var txData map[string]interface{}
				err = json.Unmarshal(body, &txData)
				if err != nil {
					fmt.Printf("Error decoding JSON response: %v\n", err)
					continue
				}

				status := txData["status"].(map[string]interface{})
				var confirmations uint64 = 0

				// Check if transaction has 4 confirmations
				if status["confirmed"] == true {
					blockHeight := uint64(status["block_height"].(float64))
					confirmations = currentBlockHeight - blockHeight + 1

					if confirmations == 1 {
						isConfirmed <- true
					}

					err = updateTxConfirmations(transaction.Txid, confirmations)
					if err != nil {
						fmt.Printf("Error updating the withdraw_history table and its confirmation number %v\n", err)
						continue
					}

					if confirmations > MIM_NUMBER_CONFIRMATIONS {
						done <- true
						return
					}
				}

				fmt.Printf("Transaction still has %d confirmations\n", confirmations)
			}
		}
	}()

	go func() {
		for r := range isConfirmed {
			if r {
				updateBalance(transaction)
			}
		}
	}()

	<-done
	fmt.Println("Transaction has 4 confirmations, cancelling ticker...")
}

func updateBalance(transaction types.Transaction) { //, balance types.Balance) {
	totalUserSpend := transaction.Sent - transaction.Received - transaction.Fee + ZENER_FEE
	fmt.Printf("🐵 isConfirmed Updating balance... totalUserSpend: %d", totalUserSpend)
}
