package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"github.com/gocql/gocql"
	"github.com/lenninlasd/TradingEngine/engine"
)

var MIN_AMOUNT float64 = 0.001

// create the order book
var OrderBook = engine.OrderBook{
	BuyOrders:  make([]engine.Order, 0, 100),
	SellOrders: make([]engine.Order, 0, 100),
}

func GetOrderBook(w http.ResponseWriter, r *http.Request) {
	Conv, _ := json.MarshalIndent(OrderBook, "", " ")
	fmt.Fprintf(w, "%s", string(Conv))
}

func CreateMakerTakerBalances(
	takerId gocql.UUID,
	makerId gocql.UUID,
	newTakerAmount float64,
	newMakerAmount float64,
) error {
	return Session.Query(
		`BEGIN BATCH 
			INSERT INTO BTCbalances(user_id, time_stamp, amount) VALUES(?, toTimeStamp(now()), ?)
			INSERT INTO BTCbalances(user_id, time_stamp, amount) VALUES(?, toTimeStamp(now()), ?)
		APPLY BATCH`, takerId, newTakerAmount, makerId, newMakerAmount).Exec()
}

func CreateOrder(w http.ResponseWriter, r *http.Request) {
	var NewOrder engine.Order
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "wrong data")
	}

	NewOrder.FromJSON(reqBody)
	NewOrder.Id = gocql.TimeUUID()
	userBalance := GetUserBalance(NewOrder.UserId)

	if NewOrder.Amount < MIN_AMOUNT {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "The minimum amount should be greater than %s", fmt.Sprintf("%v", MIN_AMOUNT))
		return
	}

	if len(userBalance) == 0 || userBalance[0].Amount <=  NewOrder.Amount {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", "There is not enough balance in your account")
		return
	}
 
	trades := OrderBook.Process(NewOrder)

	var newTakerAmount = userBalance[0].Amount

	for _, trade := range trades {
		var newMakerAmount float64
		makerUserBalance := GetUserBalance(trade.MakerUserId)

		if NewOrder.Side == 1 {
			newTakerAmount = newTakerAmount + trade.Amount
			newMakerAmount = makerUserBalance[0].Amount - trade.Amount
		} else {
			newTakerAmount = newTakerAmount - trade.Amount
			newMakerAmount = makerUserBalance[0].Amount + trade.Amount
		}

		err = CreateMakerTakerBalances(
			NewOrder.UserId,
			makerUserBalance[0].UserId,
			newTakerAmount,
			newMakerAmount,
		)

		if err != nil {
			break
		}
	}

	if err != nil {
		fmt.Println("Error while inserting")
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	Conv, _ := json.MarshalIndent(OrderBook, "", " ")
	fmt.Fprintf(w, "%s", string(Conv))
}
