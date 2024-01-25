package withdrawal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lenninlasd/Engine/middleware"
	"github.com/lenninlasd/Engine/postgres"
)

type WithdrawalsHistory struct {
	Id        uuid.UUID `json:"id"`
	Amount    uint64    `json:"amount"`
	Address   string    `json:"address"`
	Txid      string    `json:"txid"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Fee       uint64    `json:"fee"`
}

func GetWithdrawalsHistory(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.GetClainsFromToken(r.Header)

	db := postgres.OpenConnection()
	query := `
		SELECT id, amount, address, txid, status, time_stamp, fee
		FROM public.withdraw_history
		WHERE user_id=$1
		ORDER BY time_stamp DESC;`

	rows, err := db.Query(query, claims.Subject)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	defer db.Close()

	// Iterate over the result set and store the data in a struct
	var withdrawalsHistory []WithdrawalsHistory
	for rows.Next() {
		var record WithdrawalsHistory
		err := rows.Scan(
			&record.Id,
			&record.Amount,
			&record.Address,
			&record.Txid,
			&record.Status,
			&record.Timestamp,
			&record.Fee,
		)
		if err != nil {
			panic(err)
		}
		withdrawalsHistory = append(withdrawalsHistory, record)
	}

	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		panic(err)
	}

	if len(withdrawalsHistory) == 0 {
		withdrawalsHistory = []WithdrawalsHistory{}
	}

	w.Header().Set("Content-Type", "application/json")
	Conv, _ := json.MarshalIndent(withdrawalsHistory, "", " ")
	fmt.Fprintf(w, "%s", Conv)
}
