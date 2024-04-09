package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/lenninlasd/Engine/middleware"
	"github.com/lenninlasd/Engine/pkg/utils"
	"github.com/lenninlasd/Engine/postgres"
)

func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Restful API using Go and Postgres! from new version")
}

type ReferrerIdBody struct {
	ReferrerId string `json:"referrer_id"`
}

func (referrerIdBody *ReferrerIdBody) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, referrerIdBody)
}

func createReferralForUser(userId string, r *http.Request) error {
	var referrerIdBody ReferrerIdBody

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	referrerIdBody.FromJSON(reqBody)

	if referrerIdBody.ReferrerId == "" {
		return nil
	}

	db := postgres.OpenConnection()

	_, err = db.Exec(
		`INSERT INTO public.referrals (user_id, referrer_id) VALUES ($1, $2)`,
		userId, referrerIdBody.ReferrerId,
	)

	defer db.Close()

	return err
}

func CreateUserAfterSignUpSupa(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.GetClainsFromToken(r.Header)
	if err != nil {
		panic(err)
	}

	db := postgres.OpenConnection()
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}

	_, err = tx.Exec(`
		INSERT INTO users (id, email, referral_id) SELECT $1, $2, $3
		WHERE NOT EXISTS (SELECT 1 FROM users WHERE id = $1)`, claims.Subject, claims.Email, utils.GenerateShortID(claims.Subject))
	if err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}

	_, err = tx.Exec(`
		INSERT INTO btcbalances (user_id, amount) SELECT $1, $2
		WHERE NOT EXISTS (SELECT 1 FROM btcbalances WHERE user_id = $1)`, claims.Subject, 0)
	if err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}

	err = tx.Commit()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}

	err = createReferralForUser(claims.Subject, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
}
