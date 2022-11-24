package main

import (
	// "log"
	// "fmt"
	// "encoding/json"
	// "net/http"

	"github.com/lenninlasd/Engine/engine"
	"github.com/lenninlasd/Engine/api"
)

// create the order book
var OrderBook = engine.OrderBook{
	BuyOrders:  make([]engine.Order, 0, 100),
	SellOrders: make([]engine.Order, 0, 100),
}

// var bookStreamList = []engine.Order{
// 	{Amount: 100, Price: 105, Id: "lennin0", Side: 1, UserId: "zener"},
// 	{Amount: 50, Price: 105, Id: "lennin1", Side: 0 },
// 	{Amount: 20, Price: 105, Id: "lennin2", Side: 0 },
// 	{Amount: 200, Price: 110, Id: "lennin3", Side: 0 },
// 	{Amount: 300, Price: 115, Id: "lennin4", Side: 1 },
// }

// func handleRequests() {
// 	http.HandleFunc("/orderbook", processOrderBook)
// 	log.Fatal(http.ListenAndServe(":8000", nil))
// }

// func processOrderBook(w http.ResponseWriter, r *http.Request) {
// 	for _, order := range bookStreamList {		
// 		OrderBook.Process(order)
// 	}

// 	fmt.Println("Endpoint Hit: processOrderBook")
// 	json.NewEncoder(w).Encode(OrderBook)
// }

func main() {
	// handleRequests()
	api.Routes()
}
