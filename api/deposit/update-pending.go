package deposit

import (
	"fmt"
	"net/http"

	"github.com/lenninlasd/Engine/constants"
	"github.com/lenninlasd/Engine/middleware"
	"github.com/lenninlasd/Engine/postgres"
)

func UpdatePendingExpiredDeposits(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.GetClainsFromToken(r.Header)
	if err != nil {
		panic(err)
	}

	db := postgres.OpenConnection()

	defer db.Close()

	query := `UPDATE public.fiat_deposit_tx
		SET status=$1
		WHERE status=$2
		AND user_id=$3
		AND time_stamp <= (CURRENT_TIMESTAMP AT TIME ZONE 'UTC' - INTERVAL '10 minutes');`

	result, err := db.Exec(query, constants.VOIDED, constants.PENDING, claims.Subject)
	if err != nil {
		http.Error(w, "Error updating status", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Error fetching update result", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("No updates performed"))
		return
	}

	responseMessage := fmt.Sprintf("Status updated successfully. Updated: %d", rowsAffected)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(responseMessage))
}
