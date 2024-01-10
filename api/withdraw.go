package api

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/lenninlasd/Engine/constants"
	"github.com/lenninlasd/Engine/middleware"
	"github.com/lenninlasd/Engine/telegrambot"
	"github.com/lenninlasd/Engine/types"
)

const (
	WITHDRAW_STATUS_INICIALIZED = "initialized"
	WITHDRAW_STATUS_BROADCASTED = "broadcasted"
	WITHDRAW_STATUS_CONFIRMED   = "confirmed"
	WITHDRAW_STATUS_CANCELED    = "canceled"

	WITHDRAW_INSUFFICIENT_FUNDS = "InsufficientFunds"

	TIME_TO_FECHT_MEMPOOL = 1
)

func Withdraw(w http.ResponseWriter, r *http.Request) {
	var withdraw types.Withdraw
	claims, err := middleware.GetClainsFromToken(r.Header)

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "wrong data")
		return
	}

	withdraw.FromJSON(reqBody)

	isUnable := checkUserWithdrawalStatus(claims.Subject)
	if isUnable {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", "User is not able to withdraw")
		return
	}

	balance, err := GetUserBalance(claims.Subject)
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
	withdrawalId := uuid.New()
	err = initWithdrawTx(withdrawalId, withdraw, claims.Subject)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	recomendedFees, err := GetMempoolFeeRates()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", "Error loading fees")
		return
	}

	transaction, err := ExecuteWithdrawal(withdrawalId, withdraw, recomendedFees.FastestFee, balance)
	if err != nil {
		panic(err)
	}

	Conv, _ := json.MarshalIndent(transaction, "", " ")
	fmt.Fprintf(w, "%s", string(Conv))
}

func ExecuteWithdrawal(withdrawalId uuid.UUID, withdraw types.Withdraw, FastestFee float32, balance types.Balance) (types.Transaction, error) {
	transaction, err := SendWithdrawToWallet(withdrawalId, withdraw, FastestFee)
	if err != nil {
		if err.Error() == WITHDRAW_INSUFFICIENT_FUNDS {
			telegrambot.InsufficientFundsAlert()
			updateError := updateInternalState(withdrawalId, WITHDRAW_INSUFFICIENT_FUNDS)
			if updateError != nil {
				cancelWithdrawalTx(withdrawalId)
				return transaction, updateError
			}
			return transaction, nil
		}
		cancelWithdrawalTx(withdrawalId)
		return transaction, err
	}

	err = broadcastedWithdrawTx(withdrawalId, transaction.Txid)
	if err != nil {
		cancelWithdrawalTx(withdrawalId)
		return transaction, err
	}

	go watchTransactionOnMempool(transaction, balance)

	return transaction, nil
}

func updateInternalState(id uuid.UUID, newState string) error {
	db := OpenConnection()
	defer db.Close()

	_, err := db.Exec(`UPDATE withdraw_history SET internal_state = $2 WHERE id = $1`, id, newState)
	return err
}

func GetPendingWithdrawalTxSupa(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()

	claims, err := middleware.GetClainsFromToken(r.Header)
	if err != nil {
		panic(err)
	}

	var tx types.BroadcastedTx

	err = db.QueryRow(
		`SELECT txid, address, amount, time_stamp, status, confirmations, blockheight
		FROM public.withdraw_history
		WHERE user_id=$1 AND (status=$2 OR status=$3 OR status=$4) ORDER BY time_stamp DESC LIMIT 1`,
		claims.Subject, WITHDRAW_STATUS_INICIALIZED, WITHDRAW_STATUS_BROADCASTED, WITHDRAW_STATUS_CONFIRMED).Scan(
		&tx.Txid, &tx.Address, &tx.Amount, &tx.Timestamp, &tx.Status, &tx.Confirmations, &tx.BlockHeight)

	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	defer db.Close()

	Conv, _ := json.MarshalIndent(tx, "", " ")
	fmt.Fprintf(w, "%s", Conv)
}

func checkUserWithdrawalStatus(UserId string) bool {
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

func initWithdrawTx(withdrawalId uuid.UUID, withdraw types.Withdraw, UserId string) error {
	db := OpenConnection()

	_, err := db.Exec(
		`INSERT INTO withdraw_history (id, user_id, amount, address, status) VALUES ($1, $2, $3, $4, $5)`,
		withdrawalId, UserId, withdraw.Amount, withdraw.Address, WITHDRAW_STATUS_INICIALIZED,
	)

	defer db.Close()

	return err
}

func broadcastedWithdrawTx(withdrawalId uuid.UUID, txid string) error {
	db := OpenConnection()

	_, err := db.Exec(
		`UPDATE withdraw_history SET status=$1, txid=$2, confirmations=0 WHERE id=$3`,
		WITHDRAW_STATUS_BROADCASTED, txid, withdrawalId,
	)

	defer db.Close()

	return err
}

func updateByConfirmations(transaction types.Transaction, blockHeight uint64, balance types.Balance) error {
	db := OpenConnection()
	defer db.Close()

	sentToUser := transaction.Sent - transaction.Received - transaction.Fee

	// MODEL 1: in this case the user is charged with on chain fees and the ZENER fee
	totalUserSpend := sentToUser + constants.ZENER_FEE + transaction.Fee

	// MODEL 2:in this case ZENER_FEE contains the fee,
	// but if the on chain fees are grater than the zener fee the user pay the on chain fees
	// totalUserSpend := sentToUser + transaction.Fee
	// if transaction.Fee <= ZENER_FEE {
	// 	totalUserSpend = sentToUser + ZENER_FEE - transaction.Fee
	// }

	newBalanceAmount := balance.Amount - totalUserSpend
	if totalUserSpend > balance.Amount {
		newBalanceAmount = 0
	}

	// Begin a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
		UPDATE withdraw_history SET status=$1, confirmations=1, blockheight=$2 WHERE txid=$3;
    `, WITHDRAW_STATUS_CONFIRMED, blockHeight, transaction.Txid)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`
        UPDATE btcbalances SET amount=$1 WHERE user_id=$2;
    `, newBalanceAmount, balance.UserId)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	return err
}

func cancelWithdrawalTx(withdrawalId uuid.UUID) error {
	db := OpenConnection()
	_, err := db.Exec(
		`UPDATE withdraw_history SET status=$1 WHERE id=$2`,
		WITHDRAW_STATUS_CANCELED, withdrawalId,
	)
	defer db.Close()
	return err
}

func SendWithdrawToWallet(withdrawalId uuid.UUID, withdraw types.Withdraw, feeRate float32) (types.Transaction, error) {
	reference := withdrawalId.String()
	integrity := os.Getenv("WITHDRAWAL_INTEGRITY")
	feeRateString := fmt.Sprintf("%g", feeRate)
	amountString := fmt.Sprintf("%d", withdraw.Amount)
	strToken := reference + feeRateString + amountString + withdraw.Address + integrity
	hash := sha256.Sum256([]byte(strToken))
	checksum := hex.EncodeToString(hash[:])

	// create the request body
	requestBody, err := json.Marshal(map[string]interface{}{
		"reference": reference,
		"amount":    withdraw.Amount,
		"address":   withdraw.Address,
		"fee_rate":  feeRate,
		"checksum":  checksum,
	})

	request, err := http.NewRequest("POST", fmt.Sprintf("%s/withdraw", constants.WALLET_ADDRESS), bytes.NewBuffer(requestBody))
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

	if response.StatusCode >= 400 || response.StatusCode >= 500 {
		var responseBody struct {
			Error string `json:"error"`
		}
		body, _ := ioutil.ReadAll(response.Body)
		json.Unmarshal(body, &responseBody)
		err := errors.New(responseBody.Error)
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

func watchTransactionOnMempool(transaction types.Transaction, balance types.Balance) {
	// TODO Implement websockets to watch new block confirmations
	ticker := time.NewTicker(TIME_TO_FECHT_MEMPOOL * time.Minute)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				ticker.Stop()
				return
			case <-ticker.C:
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

				if status["confirmed"] == true {
					blockHeight := uint64(status["block_height"].(float64))

					err = updateByConfirmations(transaction, blockHeight, balance)
					if err != nil {
						fmt.Printf("Error updating the withdraw_history table and its confirmation number %v\n", err)
						continue
					}

					done <- true
					return
				}

				fmt.Printf("Transaction still has %d confirmations\n", confirmations)
			}
		}
	}()

	<-done
	fmt.Println("Transaction is confirmed, cancelling ticker...")
}

func getApi(url string, baseKey string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", baseKey)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", baseKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func GetBlockedWithdrawals(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.GetClainsFromToken(r.Header)
	var tokenString = os.Getenv("SUPA_SERVICE_ROLE_TOKEN")
	var supabaseProjectId = os.Getenv("SUPA_PROJECT_ID")
	supabaseApiURL := fmt.Sprintf("https://%s.supabase.co/rest/v1/users?select=role&id=eq.%s", supabaseProjectId, claims.Subject)

	body, err := getApi(supabaseApiURL, tokenString)
	if err != nil {
		panic(err)
	}

	var userRoles []types.UserRole
	err = json.Unmarshal(body, &userRoles)

	if len(userRoles) == 0 || userRoles[0].Role != "admin" {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "%s", "Forbidden: User does not have admin role.")
		return
	}

	db := OpenConnection()

	query := `
		SELECT withdraw_history.id, user_id, email, amount, address, status, time_stamp
		FROM withdraw_history
		JOIN users ON users.id = user_id
		WHERE internal_state=$1;
	`
	rows, err := db.Query(query, WITHDRAW_INSUFFICIENT_FUNDS)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	// Iterate over the result set and store the data in a struct
	var withdrawalRecords []types.WithdrawalRecord
	for rows.Next() {
		var record types.WithdrawalRecord
		err := rows.Scan(
			&record.Id,
			&record.UserId,
			&record.Email,
			&record.Amount,
			&record.Address,
			&record.Status,
			&record.Timestamp,
		)
		if err != nil {
			panic(err)
		}
		withdrawalRecords = append(withdrawalRecords, record)
	}

	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		panic(err)
	}

	defer db.Close()

	Conv, _ := json.MarshalIndent(withdrawalRecords, "", " ")
	fmt.Fprintf(w, "%s", Conv)
}
