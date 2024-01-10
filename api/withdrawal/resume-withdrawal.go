package withdrawal

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/lenninlasd/Engine/pkg/withdrawal"
	"github.com/lenninlasd/Engine/postgres"
	"github.com/lenninlasd/Engine/types"
)

func ResumeWithdrawal(w http.ResponseWriter, r *http.Request) {
	withdrawalId, err := uuid.Parse(mux.Vars(r)["withdrawalId"])
	if err != nil {
		panic(err)
	}
	// TODO: and check this person can withdraw that amount

	var withdraw types.Withdraw
	var userId string

	db := postgres.OpenConnection()

	query := "SELECT user_id, amount, address FROM public.withdraw_history WHERE id=$1;"
	err = db.QueryRow(query, withdrawalId).Scan(&userId, &withdraw.Amount, &withdraw.Address)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var balance types.Balance
	err = db.QueryRow("SELECT * FROM btcbalances WHERE user_id=$1 LIMIT 1", userId).Scan(&balance.UserId, &balance.Amount, &balance.Timestamp)

	transaction, err := withdrawal.Execute(withdrawalId, withdraw, 10, balance)
	if err != nil {
		panic(err)
	}

	Conv, _ := json.MarshalIndent(transaction, "", " ")
	fmt.Fprintf(w, "%s", string(Conv))
}
