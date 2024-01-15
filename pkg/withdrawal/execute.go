package withdrawal

import (
	"bytes"
	"crypto/sha256"
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
	"github.com/lenninlasd/Engine/postgres"
	"github.com/lenninlasd/Engine/telegrambot"
	"github.com/lenninlasd/Engine/types"
)

func Execute(withdrawalId uuid.UUID, withdraw types.Withdraw, FastestFee float32) (types.Transaction, error) {
	transaction, err := SendWithdrawToWallet(withdrawalId, withdraw, FastestFee)
	if err != nil {
		if err.Error() == constants.WITHDRAW_INSUFFICIENT_FUNDS {
			telegrambot.InsufficientFundsAlert()
			updateError := updateInternalState(withdrawalId, constants.WITHDRAW_INSUFFICIENT_FUNDS)
			if updateError != nil {
				cancelWithdrawalTx(withdrawalId)
				return transaction, updateError
			}
			return transaction, err
		}
		cancelWithdrawalTx(withdrawalId)
		return transaction, err
	}

	err = broadcastedWithdrawTx(withdrawalId, transaction.Txid)
	if err != nil {
		cancelWithdrawalTx(withdrawalId)
		return transaction, err
	}

	return transaction, nil
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

func updateInternalState(id uuid.UUID, newState string) error {
	db := postgres.OpenConnection()
	defer db.Close()

	_, err := db.Exec(`UPDATE withdraw_history SET internal_state = $2 WHERE id = $1`, id, newState)
	return err
}

func cancelWithdrawalTx(withdrawalId uuid.UUID) error {
	db := postgres.OpenConnection()
	_, err := db.Exec(
		`UPDATE withdraw_history SET status=$1 WHERE id=$2`,
		constants.WITHDRAW_STATUS_CANCELED, withdrawalId,
	)
	defer db.Close()
	return err
}

func broadcastedWithdrawTx(withdrawalId uuid.UUID, txid string) error {
	db := postgres.OpenConnection()

	_, err := db.Exec(
		`UPDATE withdraw_history SET status=$1, txid=$2, confirmations=0, internal_state=null WHERE id=$3`,
		constants.WITHDRAW_STATUS_BROADCASTED, txid, withdrawalId,
	)

	defer db.Close()

	return err
}

func WatchTransactionOnMempool(transaction types.Transaction, balance types.Balance) {
	// TODO Implement websockets to watch new block confirmations
	ticker := time.NewTicker(constants.TIME_TO_FECHT_MEMPOOL * time.Minute)
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

func updateByConfirmations(transaction types.Transaction, blockHeight uint64, balance types.Balance) error {
	db := postgres.OpenConnection()
	defer db.Close()

	sentToUser := transaction.Sent - transaction.Received - transaction.Fee

	feeToCharge := transaction.Fee
	if transaction.FeeToCharge != nil {
		feeToCharge = *transaction.FeeToCharge
	}

	// MODEL 1: in this case the user is charged with on chain fees and the ZENER fee
	totalUserSpend := sentToUser + constants.ZENER_FEE + feeToCharge

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
    `, constants.WITHDRAW_STATUS_CONFIRMED, blockHeight, transaction.Txid)
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
