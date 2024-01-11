package withdrawal

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/lenninlasd/Engine/constants"
	"github.com/lenninlasd/Engine/middleware"
	"github.com/lenninlasd/Engine/pkg/utils"
	"github.com/lenninlasd/Engine/pkg/withdrawal"
	"github.com/lenninlasd/Engine/postgres"
	"github.com/lenninlasd/Engine/types"
)

func ResumeWithdrawal(w http.ResponseWriter, r *http.Request) {
	withdrawalId, err := uuid.Parse(mux.Vars(r)["withdrawalId"])
	if err != nil {
		panic(err)
	}

	claims, err := middleware.GetClainsFromToken(r.Header)
	if err != nil {
		panic(err)
	}

	role, err := utils.GetUserRole(claims.Subject)
	if err != nil {
		panic(err)
	}

	if role != constants.ADMIN_ROLE {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "%s", "Forbidden: User does not have admin role.")
		return
	}

	var withdraw types.Withdraw
	var userId string
	var internalState string

	db := postgres.OpenConnection()
	defer db.Close()

	query := "SELECT user_id, amount, address, internal_state FROM public.withdraw_history WHERE id=$1;"
	err = db.QueryRow(query, withdrawalId).Scan(&userId, &withdraw.Amount, &withdraw.Address, &internalState)
	if err != nil {
		panic(err)
	}

	if internalState != constants.WITHDRAW_INSUFFICIENT_FUNDS {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "%s", "Forbidden: this transaction is not allawed to be resume.")
		return
	}

	var balance types.Balance
	err = db.QueryRow("SELECT * FROM btcbalances WHERE user_id=$1 LIMIT 1", userId).Scan(&balance.UserId, &balance.Amount, &balance.Timestamp)
	transaction, err := withdrawal.Execute(withdrawalId, withdraw, 10, balance)
	if err != nil {
		panic(err)
	}

	if withdraw.Amount > balance.Amount {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", "User does not have sufficient funds, the withdrawal must be less than the current balance")
		return
	}

	Conv, _ := json.MarshalIndent(transaction, "", " ")
	fmt.Fprintf(w, "%s", string(Conv))
}
