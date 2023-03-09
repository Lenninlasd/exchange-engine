package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	// "github.com/gorilla/mux"
	"github.com/lenninlasd/Engine/types"
)

func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Restful API using Go and Postgres!")
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()
	var NewUser types.User
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "wrong data")
	}

	NewUser.FromJSON(reqBody)

	_, err = db.Exec(`INSERT INTO users (email) VALUES ($1)`, NewUser.Email)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	w.WriteHeader(http.StatusCreated)
	Conv, _ := json.MarshalIndent(NewUser, "", " ")
	fmt.Fprintf(w, "%s", string(Conv))
	defer db.Close()
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

func GetUserByEmail(email string) (types.User, error) {
	db := OpenConnection()
	var user types.User

	err := db.QueryRow("SELECT id, email FROM users WHERE email=$1", email).Scan(&user.Id, &user.Email)

	defer db.Close()
	return user, err
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

// func UpdateUser(w http.ResponseWriter, r *http.Request) {
// 	UserId := mux.Vars(r)["id"]
// 	var UpdateUser types.User
// 	reqBody, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		fmt.Fprintf(w, "Kindly enter data properly")
// 	}
// 	json.Unmarshal(reqBody, &UpdateUser)

// 	err = Session.Query("UPDATE users SET email = ? WHERE id = ?", UpdateUser.Email, UserId).Exec();
// 	if err != nil {
// 		fmt.Println("Error while updating")
// 		fmt.Println(err)
// 	}
// 	fmt.Fprintf(w, "updated successfully")

// }
