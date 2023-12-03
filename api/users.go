package api

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/lenninlasd/Engine/middleware"
	"github.com/lenninlasd/Engine/types"
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

func GetUserFromHeader(r *http.Request) (types.Auth0User, error) {
	var auth0User types.Auth0User
	accessToken := middleware.GetTokenFromHeader(r.Header)
	if accessToken == "" {
		return auth0User, errors.New("Authorization header missing")
	}

	auth0User, err := getAuth0UserFromToken(accessToken)
	if err != nil {
		return auth0User, err
	}

	return auth0User, nil
}

func getAuth0UserFromToken(accessToken string) (types.Auth0User, error) {
	var auth0User types.Auth0User

	protocol, present := os.LookupEnv("AUTH0_PROTOCOL")
	if !present {
		protocol = "https"
	}

	issuerURL, err := url.Parse(protocol + "://" + os.Getenv("AUTH0_DOMAIN") + "/userinfo")
	if err != nil {
		return auth0User, err
	}

	req, err := http.NewRequest("GET", issuerURL.String(), nil)
	if err != nil {
		return auth0User, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return auth0User, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return auth0User, err
	}

	err = auth0User.FromJSON(body)
	if err != nil {
		return auth0User, err
	}

	return auth0User, nil
}
