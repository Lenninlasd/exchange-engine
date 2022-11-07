package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"github.com/gocql/gocql"

	// "github.com/gorilla/mux"
	// "github.com/lenninlasd/TradingEngine/types"
	"github.com/lenninlasd/TradingEngine/engine"
)

var MIN_AMOUNT float64 = 0.001

// create the order book
var OrderBook = engine.OrderBook{
	BuyOrders:  make([]engine.Order, 0, 100),
	SellOrders: make([]engine.Order, 0, 100),
}

func CreateOrder(w http.ResponseWriter, r *http.Request) {
	var NewOrder engine.Order
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "wrong data")
	}

	NewOrder.FromJSON(reqBody)
	NewOrder.ID = gocql.TimeUUID()
	userBalance := GetUserBalance(NewOrder.UserId)
	
	if NewOrder.Amount < MIN_AMOUNT {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", "The minimum amount should be greater than MIN_AMOUNT")
		return
	}

	if len(userBalance) == 0 || userBalance[0].Amount <=  NewOrder.Amount {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", "There is not enough balance in your account")
		return
	}
 
	trades := OrderBook.Process(NewOrder)

	fmt.Println("NewOrder", NewOrder)
	fmt.Println("trades", trades)


	// if err != nil {
	// 	fmt.Println("Error while inserting")
	// 	fmt.Println(err)
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }
	w.WriteHeader(http.StatusCreated)
	Conv, _ := json.MarshalIndent(OrderBook, "", " ")
	fmt.Fprintf(w, "%s", string(Conv))
}

func GetOrderBook(w http.ResponseWriter, r *http.Request) {
	var orders []engine.Order
	m := map[string]interface{}{}

	iter := Session.Query("SELECT * FROM orderbook").Iter()
	for iter.MapScan(m) {
		orders = append(orders, engine.Order{
			ID:		m["id"].(gocql.UUID),
			UserId:	m["user_id"].(gocql.UUID),
			Amount:	m["amount"].(float64),
			Price:	m["price"].(float64),
			Side:	m["side"].(int8),
		})
		m = map[string]interface{}{}
	}

	Conv, _ := json.MarshalIndent(orders, "", " ")
	fmt.Fprintf(w, "%s", string(Conv))

}

// func GetOneUser(w http.ResponseWriter, r *http.Request) {
// 	UserID := mux.Vars(r)["id"]
// 	var users []types.User
// 	m := map[string]interface{}{}

// 	iter := Session.Query("SELECT * FROM users WHERE id=?", UserID).Iter()
// 	for iter.MapScan(m) {
// 		users = append(users, types.User{
// 			ID:		m["id"].(gocql.UUID),
// 			Email:	m["email"].(string),
// 		})
// 		m = map[string]interface{}{}
// 	}

// 	Conv, _ := json.MarshalIndent(users, "", " ")
// 	fmt.Fprintf(w, "%s", string(Conv))

// }

// func ExistUser(UserID gocql.UUID) bool {
// 	var Count int
// 	err := Session.Query("SELECT count(*) FROM users WHERE id=?", UserID).Scan(&Count)
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
// 	UserID := mux.Vars(r)["id"]
// 	err := Session.Query("DELETE FROM users WHERE id = ?", UserID).Exec();
// 	if err != nil {
// 		fmt.Println("Error while deleting")
// 		fmt.Println(err)
// 	}
// 	fmt.Fprintf(w, "deleted successfully the user num %s ", UserID)
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
// 	UserID := mux.Vars(r)["id"]
// 	var UpdateUser types.User
// 	reqBody, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		fmt.Fprintf(w, "Kindly enter data properly")
// 	}
// 	json.Unmarshal(reqBody, &UpdateUser)
	
// 	err = Session.Query("UPDATE users SET email = ? WHERE id = ?", UpdateUser.Email, UserID).Exec();
// 	if err != nil {
// 		fmt.Println("Error while updating")
// 		fmt.Println(err)
// 	}
// 	fmt.Fprintf(w, "updated successfully")

// }