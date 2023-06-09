package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lenninlasd/Engine/types"
)

// func CreateBalance(w http.ResponseWriter, r *http.Request) {
// 	var NewBalance types.Balance
// 	reqBody, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		fmt.Fprintf(w, "wrong data")
// 	}

// 	NewBalance.FromJSON(reqBody)

// 	hasUser := ExistUser(NewBalance.UserId)

// 	if !hasUser {
// 		w.WriteHeader(http.StatusBadRequest)
// 		fmt.Fprintf(w, "%s", "User do not exist")
// 		return
// 	}

// 	err = Session.Query("INSERT INTO BTCbalances(user_id, time_stamp, amount) VALUES(?, toTimeStamp(now()), ?)",
// 		NewBalance.UserId, NewBalance.Amount).Exec();

// 	if err != nil {
// 		fmt.Println("Error while inserting")
// 		fmt.Println(err)
// 		w.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}
// 	w.WriteHeader(http.StatusCreated)
// 	Conv, _ := json.MarshalIndent(NewBalance, "", " ")
// 	fmt.Fprintf(w, "%s", string(Conv))
// }

func GetUserBalance(UserId string) (types.Balance, error) {
	db := OpenConnection()
	var balance types.Balance

	err := db.QueryRow("SELECT * FROM btcbalances WHERE user_id=$1 LIMIT 1", UserId).Scan(&balance.UserId, &balance.Amount, &balance.Timestamp)

	defer db.Close()

	return balance, err
}

func GetUserBalanceFromHttp(w http.ResponseWriter, r *http.Request) {
	UserId := mux.Vars(r)["userid"]
	balance, err := GetUserBalance(UserId)

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

// func GetAllBalances(w http.ResponseWriter, r *http.Request) {
// 	var balances []types.Balance
// 	m := map[string]interface{}{}

// 	iter := Session.Query("SELECT * FROM BTCbalances").Iter()
// 	for iter.MapScan(m) {
// 		balances = append(balances, types.Balance{
// 			UserId:		m["user_id"].(gocql.UUID),
// 			Timestamp:	m["time_stamp"].(time.Time),
// 			Amount:		m["amount"].(float64),
// 		})
// 		m = map[string]interface{}{}
// 	}

// 	Conv, _ := json.MarshalIndent(balances, "", " ")
// 	fmt.Fprintf(w, "%s", string(Conv))
// }

// func CountAllBalances(w http.ResponseWriter, r *http.Request) {
// 	var Count string
// 	err := Session.Query("SELECT count(*) FROM BTCbalances").Scan(&Count)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Fprintf(w, "%s ", Count)

// }

// func DeleteAllBalances(w http.ResponseWriter, r *http.Request) {
// 	err := Session.Query("TRUNCATE BTCbalances").Exec()
// 	if err != nil {
// 		fmt.Println("Error while deleting all BTCbalances")
// 		fmt.Println(err)
// 	}
// 	fmt.Fprintf(w, "deleted all successfully")
// }
