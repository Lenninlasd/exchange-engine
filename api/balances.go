package api

import (
	"encoding/json"
	"fmt"
	"time"
	"io/ioutil"
	"net/http"
	"github.com/gocql/gocql"
	"github.com/gorilla/mux"

	"github.com/lenninlasd/Engine/types"
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

	err = Session.Query("INSERT INTO BTCbalances(user_id, time_stamp, amount) VALUES(?, toTimeStamp(now()), ?)",
		NewBalance.UserId, NewBalance.Amount).Exec();

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

func GetUserBalance(UserId gocql.UUID, tokenType string) types.Balance {
	var balances []types.Balance
	m := map[string]interface{}{}

	iter := Session.Query("SELECT * FROM " + tokenType + "balances WHERE user_id=? ORDER BY time_stamp DESC LIMIT 1", UserId).Iter()
	for iter.MapScan(m) {
		balances = append(balances, types.Balance{
			UserId:		m["user_id"].(gocql.UUID),
			Timestamp:	m["time_stamp"].(time.Time),
			Amount:		m["amount"].(float64),
		})
		m = map[string]interface{}{}
	}

	if len(balances) > 0 {
		return balances[0]
	}

	return types.Balance{
		UserId: UserId,
	}
}

func GetAllBalances(w http.ResponseWriter, r *http.Request) {
	var balances []types.Balance
	m := map[string]interface{}{}

	iter := Session.Query("SELECT * FROM BTCbalances").Iter()
	for iter.MapScan(m) {
		balances = append(balances, types.Balance{
			UserId:		m["user_id"].(gocql.UUID),
			Timestamp:	m["time_stamp"].(time.Time),
			Amount:		m["amount"].(float64),
		})
		m = map[string]interface{}{}
	}

	Conv, _ := json.MarshalIndent(balances, "", " ")
	fmt.Fprintf(w, "%s", string(Conv))
}

func GetUserBalanceFromHttp(w http.ResponseWriter, r *http.Request) {
	UserId := mux.Vars(r)["id"]
	
	Useruuid, err := gocql.ParseUUID(UserId)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", "User do not exist")
		return
	}

	balances := GetUserBalance(Useruuid, "BTC")

	Conv, _ := json.MarshalIndent(balances, "", " ")
	fmt.Fprintf(w, "%s", string(Conv))
}

func CountAllBalances(w http.ResponseWriter, r *http.Request) {
	var Count string
	err := Session.Query("SELECT count(*) FROM BTCbalances").Scan(&Count)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(w, "%s ", Count)

}

func DeleteAllBalances(w http.ResponseWriter, r *http.Request) {
	err := Session.Query("TRUNCATE BTCbalances").Exec();
	if err != nil {
		fmt.Println("Error while deleting all BTCbalances")
		fmt.Println(err)
	}
	fmt.Fprintf(w, "deleted all successfully")
}
