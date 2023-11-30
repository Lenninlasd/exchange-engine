package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lenninlasd/Engine/middleware"
	"github.com/lenninlasd/Engine/types"
)

func GetUserBalance(UserId string) (types.Balance, error) {
	db := OpenConnection()
	var balance types.Balance

	err := db.QueryRow("SELECT * FROM btcbalances WHERE user_id=$1 LIMIT 1", UserId).Scan(&balance.UserId, &balance.Amount, &balance.Timestamp)

	defer db.Close()

	return balance, err
}

func GetUserBalanceFromHttpSupa(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.GetClainsFromToken(r.Header)
	balance, err := GetUserBalance(claims.Subject)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "%s", "Balance not found")
			return
		}
		panic(err)
	}

	Conv, _ := json.MarshalIndent(balance, "", " ")
	fmt.Fprintf(w, "%s", string(Conv))
}

func userNotExist(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, "%s", "User do not exist")
}
