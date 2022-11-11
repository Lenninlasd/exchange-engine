package api

import (
	"encoding/json"
	"fmt"
	"errors"
	"io/ioutil"
	"net/http"
	"github.com/gocql/gocql"
	"github.com/lenninlasd/TradingEngine/engine"
	"github.com/lenninlasd/TradingEngine/types"
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

func GetUserAmountInExistingOrder(userId gocql.UUID, side int8) float64 {
	var orders []engine.Order
	if side == 1 {
		orders = OrderBook.BuyOrders
	} else {
		orders = OrderBook.SellOrders
	}

	var existingAmount float64 = 0
	for _, order := range orders {
		if order.UserId != userId {
			continue
		}

		if side == 1 {
			existingAmount += order.Amount * order.Price
			continue
		}
		existingAmount += order.Amount
	}

	return existingAmount
}

func CheckEnoughBalance(order engine.Order) (types.Balance,  types.Balance, error) {
	userExistingOrderAmount := GetUserAmountInExistingOrder(order.UserId, order.Side)

	userBTCBalance := GetUserBalance(order.UserId, "BTC")
	userUSDTBalance := GetUserBalance(order.UserId, "USDT")

	err := errors.New("There is not enough balance in your account")

	if order.Side == 1 {
	 	// userBalance amount(USDT) - userExistingOrderAmount in USDT < Amount(BTC)*Price(USD/BTC)
		hasEnoughBalance := userUSDTBalance[0].Amount - userExistingOrderAmount >=  order.Amount * order.Price
		if len(userUSDTBalance) == 0 || !hasEnoughBalance {
			return types.Balance{}, types.Balance{}, err
		}
	} else {
		if len(userBTCBalance) == 0 || (userBTCBalance[0].Amount - userExistingOrderAmount <  order.Amount) {
			return types.Balance{}, types.Balance{}, err
		}
	}
	return userBTCBalance[0], userUSDTBalance[0], nil
}

func CreateOrder(w http.ResponseWriter, r *http.Request) {
	var NewOrder engine.Order
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "wrong data")
	}

	NewOrder.FromJSON(reqBody)
	NewOrder.Id = gocql.TimeUUID()

	if NewOrder.Amount < MIN_AMOUNT {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "The minimum amount should be greater than %s", fmt.Sprintf("%v", MIN_AMOUNT))
		return
	}

	userBTCBalance, _, err := CheckEnoughBalance(NewOrder)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", err)
		return
	}
 
	trades := OrderBook.Process(NewOrder)

	var newTakerAmount = userBTCBalance.Amount

	for _, trade := range trades {
		var newMakerAmount float64
		makerUserBalance := GetUserBalance(trade.MakerUserId, "BTC")

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
