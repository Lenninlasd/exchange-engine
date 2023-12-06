package api

import (
	"fmt"
	"net/http"

	"github.com/lenninlasd/Engine/middleware"
)

func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Restful API using Go and Postgres! from new version")
}

func CreateUserAfterSignUpSupa(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.GetClainsFromToken(r.Header)
	if err != nil {
		panic(err)
	}

	db := OpenConnection()
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}

	_, err = tx.Exec(`
		INSERT INTO users (id, email) SELECT $1, $2
		WHERE NOT EXISTS (SELECT 1 FROM users WHERE id = $1)`, claims.Subject, claims.Email)
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

	w.WriteHeader(http.StatusOK)
}
