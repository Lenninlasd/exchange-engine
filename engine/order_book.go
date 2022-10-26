package engine

import (
	"sort"
)

type OrderBook struct {
	BuyOrders  []Order
	SellOrders []Order
}

func insertAt(data []Order, i int, order Order) []Order {
    if i == len(data) {
        return append(data, order)
    }

    data = append(data[:i+1], data[i:]...)

    data[i] = order

    return data
}

func insertOrderInSortedList(data []Order, order Order, asc int8) []Order {
	i := sort.Search(len(data), func(i int) bool {
		if asc == 1 {
			return data[i].Price >= order.Price
		}
		return data[i].Price <= order.Price
	})
	return insertAt(data, i, order)
}

func (book *OrderBook) addBuyOrder(order Order) {
	book.BuyOrders = insertOrderInSortedList(book.BuyOrders, order, 1)
}

func (book *OrderBook) addSellOrder(order Order) {
	book.SellOrders = insertOrderInSortedList(book.SellOrders, order, 0)
}

func (book *OrderBook) removeBuyOrder(index int) {
	book.BuyOrders = append(book.BuyOrders[:index], book.BuyOrders[index+1:]...)
}

func (book *OrderBook) removeSellOrder(index int) {
	book.SellOrders = append(book.SellOrders[:index], book.SellOrders[index+1:]...)
}
