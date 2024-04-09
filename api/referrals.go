package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lenninlasd/Engine/postgres"
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
