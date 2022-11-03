package api

import (
	"encoding/json"
	"fmt"
	"time"
	"io/ioutil"
	"net/http"
	"github.com/gocql/gocql"
	"github.com/gorilla/mux"

	"github.com/lenninlasd/TradingEngine/types"
)

func CreateBalance(w http.ResponseWriter, r *http.Request) {
	var NewBalance types.Balance
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "wrong data")
	}

	NewBalance.FromJSON(reqBody)

	hasUser := ExistUser(NewBalance.UserId)

	if !hasUser {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", "User do not exist")
		return
	}

	err = Session.Query("INSERT INTO balances(id, user_id, coin_type, amount, time_stamp) VALUES(uuid(), ?, ?, ?, toTimeStamp(now()))",
		NewBalance.UserId, NewBalance.CoinType, NewBalance.Amount).Exec();

	if err != nil {
		fmt.Println("Error while inserting")
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	Conv, _ := json.MarshalIndent(NewBalance, "", " ")
	fmt.Fprintf(w, "%s", string(Conv))
}

func GetAllBalances(w http.ResponseWriter, r *http.Request) {
	var balances []types.Balance
	m := map[string]interface{}{}

	iter := Session.Query("SELECT * FROM balances").Iter()
	for iter.MapScan(m) {
		balances = append(balances, types.Balance{
			Id:			m["id"].(gocql.UUID),
			UserId:		m["user_id"].(gocql.UUID),
			CoinType:	m["coin_type"].(string),
			Amount:		m["amount"].(int64),
			Timestamp:	m["time_stamp"].(time.Time),
		})
		m = map[string]interface{}{}
	}

	Conv, _ := json.MarshalIndent(balances, "", " ")
	fmt.Fprintf(w, "%s", string(Conv))
}

func GetOneBalance(w http.ResponseWriter, r *http.Request) {
	UserId := mux.Vars(r)["id"]
	var balances []types.Balance
	m := map[string]interface{}{}

	iter := Session.Query("SELECT * FROM balances WHERE id=?", UserId).Iter()
	for iter.MapScan(m) {
		balances = append(balances, types.Balance{
			Id:			m["id"].(gocql.UUID),
			UserId:		m["user_id"].(gocql.UUID),
			CoinType:	m["coin_type"].(string),
			Amount:		m["amount"].(int64),
			Timestamp:	m["time_stamp"].(time.Time),
		})
		m = map[string]interface{}{}
	}

	Conv, _ := json.MarshalIndent(balances, "", " ")
	fmt.Fprintf(w, "%s", string(Conv))
}

func CountAllBalances(w http.ResponseWriter, r *http.Request) {
	var Count string
	err := Session.Query("SELECT count(*) FROM balances").Scan(&Count)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(w, "%s ", Count)

}
func DeleteOneBalance(w http.ResponseWriter, r *http.Request) {
	UserId := mux.Vars(r)["id"]
	err := Session.Query("DELETE FROM balances WHERE user_id = ?", UserId).Exec();
	if err != nil {
		fmt.Println("Error while deleting")
		fmt.Println(err)
	}
	fmt.Fprintf(w, "deleted successfully the balance num %s ", UserId)
}

func DeleteAllBalances(w http.ResponseWriter, r *http.Request) {
	err := Session.Query("TRUNCATE balances").Exec();
	if err != nil {
		fmt.Println("Error while deleting all balances")
		fmt.Println(err)
	}
	fmt.Fprintf(w, "deleted all successfully")

}

func UpdateBalance(w http.ResponseWriter, r *http.Request) {
	UserId := mux.Vars(r)["user_id"]
	var UpdateBalance types.Balance
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Kindly enter data properly")
	}
	json.Unmarshal(reqBody, &UpdateBalance)
	
	err = Session.Query("UPDATE balances SET coin_type = ?, amount = ?, time_stamp = ? WHERE user_id = ?",
		UpdateBalance.CoinType, UpdateBalance.Amount, UpdateBalance.Timestamp, UserId).Exec();
	
	if err != nil {
		fmt.Println("Error while updating")
		fmt.Println(err)
	}
	fmt.Fprintf(w, "updated successfully")
}