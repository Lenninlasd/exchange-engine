package withdrawal

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lenninlasd/Engine/constants"
	"github.com/lenninlasd/Engine/middleware"
	"github.com/lenninlasd/Engine/pkg/utils"
	"github.com/lenninlasd/Engine/postgres"
	"github.com/lenninlasd/Engine/types"
)

func GetBlockedWithdrawals(w http.ResponseWriter, r *http.Request) {
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

	db := postgres.OpenConnection()

	query := `
		SELECT withdraw_history.id, user_id, email, amount, address, status, time_stamp
		FROM withdraw_history
		JOIN users ON users.id = user_id
		WHERE internal_state=$1;
	`
	rows, err := db.Query(query, constants.WITHDRAW_INSUFFICIENT_FUNDS)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	defer db.Close()

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

	if len(withdrawalRecords) == 0 {
		withdrawalRecords = []types.WithdrawalRecord{}
	}

	w.Header().Set("Content-Type", "application/json")
	Conv, _ := json.MarshalIndent(withdrawalRecords, "", " ")
	fmt.Fprintf(w, "%s", Conv)
}
