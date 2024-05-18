package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lenninlasd/Engine/constants"
	"github.com/lenninlasd/Engine/middleware"
	"github.com/lenninlasd/Engine/postgres"
	"github.com/lenninlasd/Engine/types"
)

func CheckReferrerUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var isReferrerValid bool

	db := postgres.OpenConnection()
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM public.users WHERE referral_id = $1)", id).Scan(&isReferrerValid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}
	defer db.Close()

	fmt.Fprintf(w, "%t", isReferrerValid)
}

func GetReferralHistory(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.GetClainsFromToken(r.Header)
	if err != nil {
		panic(err)
	}

	db := postgres.OpenConnection()
	query := `
		SELECT t1.id, t1.user_id, t2.user_id as referred_user_id, users.email as referred_user_email, t1.purchased_sats, t1.amount_in_cents, t1.time_stamp
		FROM public.fiat_deposit_tx AS t1
		JOIN public.fiat_deposit_tx AS t2 ON t1.id = t2.id AND t1.reference <> t2.reference
		JOIN users ON users.id=t2.user_id
		WHERE t1.origin=$1
			AND t1.status=$2 AND t2.status=$2
			AND t1.user_id=$3;`

	rows, err := db.Query(query, constants.REFERRING_ORIGIN, constants.APPROVED, claims.Subject)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	defer db.Close()

	var referralHistoryList []types.ReferralHistory

	for rows.Next() {
		var record types.ReferralHistory
		err := rows.Scan(
			&record.Id,
			&record.UserId,
			&record.ReferredUserId,
			&record.ReferredUserEmail,
			&record.PurchasedSats,
			&record.AmountInCents,
			&record.Timestamp,
		)
		if err != nil {
			panic(err)
		}
		referralHistoryList = append(referralHistoryList, record)
	}

	if err := rows.Err(); err != nil {
		panic(err)
	}

	if len(referralHistoryList) == 0 {
		referralHistoryList = []types.ReferralHistory{}
	}

	w.Header().Set("Content-Type", "application/json")
	Conv, _ := json.MarshalIndent(referralHistoryList, "", " ")
	fmt.Fprintf(w, "%s", Conv)
}
