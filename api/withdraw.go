package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
	"github.com/lenninlasd/Engine/constants"
	"github.com/lenninlasd/Engine/middleware"
	"github.com/lenninlasd/Engine/pkg/utils"
	"github.com/lenninlasd/Engine/pkg/withdrawal"
	"github.com/lenninlasd/Engine/postgres"
	"github.com/lenninlasd/Engine/types"
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

	recomendedFees, err := utils.GetMempoolFeeRates()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", "Error loading fees")
		return
	}

	withdrawalId := uuid.New()
	err = initWithdrawTx(withdrawalId, withdraw, claims.Subject, recomendedFees.FastestFee)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	transaction, err := withdrawal.Execute(withdrawalId, withdraw, recomendedFees.FastestFee)
	if err != nil {
		if err.Error() == constants.WITHDRAW_INSUFFICIENT_FUNDS {
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", "Error Executing withdrawal")
		panic(err)
	}

	go withdrawal.WatchTransactionOnMempool(transaction, balance)

	Conv, _ := json.MarshalIndent(transaction, "", " ")
	fmt.Fprintf(w, "%s", string(Conv))
}

func GetPendingWithdrawalTxSupa(w http.ResponseWriter, r *http.Request) {
	db := postgres.OpenConnection()

	claims, err := middleware.GetClainsFromToken(r.Header)
	if err != nil {
		panic(err)
	}

	var tx types.BroadcastedTx

	err = db.QueryRow(
		`SELECT txid, address, amount, time_stamp, status, confirmations, blockheight
		FROM public.withdraw_history
		WHERE user_id=$1 AND (status=$2 OR status=$3 OR status=$4) ORDER BY time_stamp DESC LIMIT 1`,
		claims.Subject, constants.WITHDRAW_STATUS_INICIALIZED, constants.WITHDRAW_STATUS_BROADCASTED, constants.WITHDRAW_STATUS_CONFIRMED).Scan(
		&tx.Txid, &tx.Address, &tx.Amount, &tx.Timestamp, &tx.Status, &tx.Confirmations, &tx.BlockHeight)

	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	defer db.Close()

	Conv, _ := json.MarshalIndent(tx, "", " ")
	fmt.Fprintf(w, "%s", Conv)
}

func checkUserWithdrawalStatus(UserId string) bool {
	db := postgres.OpenConnection()
	defer db.Close()

	query := `SELECT EXISTS (SELECT 1 FROM withdraw_history WHERE user_id = $1 AND status IN ($2, $3))`

	var exists bool
	err := db.QueryRow(query, UserId, constants.WITHDRAW_STATUS_INICIALIZED, constants.WITHDRAW_STATUS_BROADCASTED).Scan(&exists)
	if err != nil {
		panic(err)
	}

	return exists
}

func initWithdrawTx(withdrawalId uuid.UUID, withdraw types.Withdraw, UserId string, feeRate float32) error {
	db := postgres.OpenConnection()

	_, err := db.Exec(
		`INSERT INTO withdraw_history (id, user_id, amount, fee_rate, address, status) VALUES ($1, $2, $3, $4, $5, $6)`,
		withdrawalId, UserId, withdraw.Amount, feeRate, withdraw.Address, constants.WITHDRAW_STATUS_INICIALIZED,
	)

	defer db.Close()

	return err
}
