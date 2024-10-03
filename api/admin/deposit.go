package admin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/lenninlasd/Engine/constants"
	"github.com/lenninlasd/Engine/middleware"
	"github.com/lenninlasd/Engine/pkg/utils"
	"github.com/lenninlasd/Engine/postgres"
	"github.com/lenninlasd/Engine/types"
)

func CheckIsAdmin(header http.Header) (bool, error) {
	claims, err := middleware.GetClainsFromToken(header)
	if err != nil {
		return false, err
	}

	role, err := utils.GetUserRole(claims.Subject)
	if err != nil {
		return false, err
	}

	if role != constants.ADMIN_ROLE {
		return false, nil
	}

	return true, nil
}

func GetApprovedDepositHistory(w http.ResponseWriter, r *http.Request) {
	isAdmin, err := CheckIsAdmin(r.Header)
	if err != nil {
		panic(err)
	}

	if !isAdmin {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "%s", "Forbidden: User does not have admin role.")
		return
	}

	db := postgres.OpenConnection()
	query := `
		SELECT reference, id, amount_in_cents, status, user_id, purchased_sats, origin, time_stamp, is_backed
		FROM public.fiat_deposit_tx WHERE status='APPROVED'
		ORDER BY time_stamp DESC;`

	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	defer db.Close()

	var fiatDepositTxHistory []types.FiatDepositTxFull

	for rows.Next() {
		var record types.FiatDepositTxFull
		err := rows.Scan(
			&record.Reference,
			&record.Id,
			&record.AmountInCents,
			&record.Status,
			&record.UserId,
			&record.PurchasedSats,
			&record.Origin,
			&record.Timestamp,
			&record.IsBacked,
		)
		if err != nil {
			panic(err)
		}
		fiatDepositTxHistory = append(fiatDepositTxHistory, record)
	}

	if err := rows.Err(); err != nil {
		panic(err)
	}

	if len(fiatDepositTxHistory) == 0 {
		fiatDepositTxHistory = []types.FiatDepositTxFull{}
	}

	w.Header().Set("Content-Type", "application/json")
	Conv, _ := json.MarshalIndent(fiatDepositTxHistory, "", " ")
	fmt.Fprintf(w, "%s", Conv)
}

func UpdateFundBackup(w http.ResponseWriter, r *http.Request) {
	isAdmin, err := CheckIsAdmin(r.Header)
	if err != nil {
		panic(err)
	}

	if !isAdmin {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "%s", "Forbidden: User does not have admin role.")
		return
	}

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "wrong data")
	}

	var reference types.Reference
	reference.FromJSON(reqBody)

	db := postgres.OpenConnection()
	defer db.Close()

	query := `UPDATE public.fiat_deposit_tx
		SET is_backed=true
		WHERE reference=$1;`

	_, err = db.Exec(query, reference.Reference)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
}
