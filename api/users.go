package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/lenninlasd/Engine/middleware"
	"github.com/lenninlasd/Engine/types"
)

func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Restful API using Go and Postgres!")
}

func CreateUserAfterSignUp(w http.ResponseWriter, r *http.Request) {
	auth0User, err := GetUserFromHeader(r)
	if err != nil {
		panic(err)
	}

	db := OpenConnection()
	_, err = db.Exec(`
		INSERT INTO users (id, email) SELECT $1, $2
		WHERE NOT EXISTS (SELECT 1 FROM users WHERE id = $1)`, auth0User.Sub, auth0User.Email)
	defer db.Close()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
}

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()

	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		panic(err)
	}

	var users []types.User
	for rows.Next() {
		var user types.User
		err := rows.Scan(&user.Id, &user.Email)
		CheckError(err)
		users = append(users, user)
	}

	Conv, _ := json.MarshalIndent(users, "", " ")
	fmt.Fprintf(w, "%s", string(Conv))

	defer rows.Close()
	defer db.Close()
}

// func GetOneUser(w http.ResponseWriter, r *http.Request) {
// 	UserId := mux.Vars(r)["id"]
// 	var users []types.UserC
// 	m := map[string]interface{}{}

// 	iter := Session.Query("SELECT * FROM users WHERE id=?", UserId).Iter()
// 	for iter.MapScan(m) {
// 		users = append(users, types.UserC{
// 			Id:		m["id"].(gocql.UUID),
// 			Email:	m["email"].(string),
// 		})
// 		m = map[string]interface{}{}
// 	}

// 	Conv, _ := json.MarshalIndent(users, "", " ")
// 	fmt.Fprintf(w, "%s", string(Conv))

// }

// func ExistUser(UserId gocql.UUID) bool {
// 	var Count int
// 	err := Session.Query("SELECT count(*) FROM users WHERE id=?", UserId).Scan(&Count)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return Count == 1
// }

// func CountAllUsers(w http.ResponseWriter, r *http.Request) {

// 	var Count string
// 	err := Session.Query("SELECT count(*) FROM users").Scan(&Count)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Fprintf(w, "%s ", Count)

// }
// func DeleteOneUser(w http.ResponseWriter, r *http.Request) {
// 	UserId := mux.Vars(r)["id"]
// 	err := Session.Query("DELETE FROM users WHERE id = ?", UserId).Exec();
// 	if err != nil {
// 		fmt.Println("Error while deleting")
// 		fmt.Println(err)
// 	}
// 	fmt.Fprintf(w, "deleted successfully the user num %s ", UserId)
// }

// func DeleteAllUsers(w http.ResponseWriter, r *http.Request) {
// 	err := Session.Query("TRUNCATE users").Exec();
// 	if err != nil {
// 		fmt.Println("Error while deleting all users")
// 		fmt.Println(err)
// 	}
// 	fmt.Fprintf(w, "deleted all successfully")

// }

func GetUserFromHeader(r *http.Request) (types.Auth0User, error) {
	var auth0User types.Auth0User
	accessToken, err := middleware.GetTokenFromHeader(r.Header)
	if err != nil {
		return auth0User, err
	}

	auth0User, err = getAuth0UserFromToken(accessToken)
	if err != nil {
		return auth0User, err
	}

	return auth0User, nil
}

func getAuth0UserFromToken(accessToken string) (types.Auth0User, error) {
	var auth0User types.Auth0User

	issuerURL, err := url.Parse("https://" + os.Getenv("AUTH0_DOMAIN") + "/userinfo")
	if err != nil {
		return auth0User, err
	}

	fmt.Println("accessToken", accessToken)

	fmt.Println("issuerURL.String()", issuerURL.String())

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
