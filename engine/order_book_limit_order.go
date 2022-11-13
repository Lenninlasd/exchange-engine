package engine

// Process an order and return the trades generated before adding the remaining amount to the market
func (book *OrderBook) Process(order Order) []Trade {
	if order.Side == 1 {
		return book.processLimitBuy(order)
	}
	return book.processLimitSell(order)
}

// Process a limit buy order
/* if the price of the order is bigger than all the sell order, nothing ot do.
if not, and the order amount is lower than the sell order just subtract the order amount of the sell order
*/
func (book *OrderBook) processLimitBuy(order Order) []Trade {
	trades := make([]Trade, 0, 1)
	n := len(book.SellOrders)

	// check if we have at least one matching order
	if n != 0 && book.SellOrders[n-1].Price <= order.Price {
		// traverse all orders that match
		for i := n - 1; i >= 0; i-- {
			sellOrder := &book.SellOrders[i]
			if sellOrder.Price > order.Price {
				break
			}
			// fill the entire order
			if sellOrder.Amount >= order.Amount {
				trades = append(trades, Trade{
					TakerOrderId: 	order.Id,
					TakerUserId: 	order.UserId,
					MakerOrderId: 	sellOrder.Id,
					MakerUserId:	sellOrder.UserId,
					Amount: 		order.Amount,
					Price: 			sellOrder.Price,
				})
				sellOrder.Amount -= order.Amount
				if sellOrder.Amount == 0 {
					book.removeSellOrder(i)
				}
				return trades
			}
			// fill a partial order and continue
			if sellOrder.Amount < order.Amount {
				trades = append(trades, Trade{
					TakerOrderId: 	order.Id,
					TakerUserId: 	order.UserId,
					MakerOrderId: 	sellOrder.Id,
					MakerUserId: 	sellOrder.UserId,
					Amount:			sellOrder.Amount,
					Price:			sellOrder.Price,
				})
				order.Amount -= sellOrder.Amount
				book.removeSellOrder(i)
				continue
			}
		}
	}
	// finally add the remaining order to the list
	book.addBuyOrder(order)
	return trades
}

// Process a limit sell order
func (book *OrderBook) processLimitSell(order Order) []Trade {
	trades := make([]Trade, 0, 1)
	n := len(book.BuyOrders)

	// check if we have at least one matching order
	if n != 0 && book.BuyOrders[n-1].Price >= order.Price {
		// traverse all orders that match
		for i := n - 1; i >= 0; i-- {
			buyOrder := &book.BuyOrders[i]
			if buyOrder.Price < order.Price {
				break
			}

			// fill the entire order
			if buyOrder.Amount >= order.Amount {
				trades = append(trades, Trade{
					TakerOrderId: order.Id, 
					TakerUserId: order.UserId,
					MakerOrderId: buyOrder.Id,
					MakerUserId: buyOrder.UserId,
					Amount: order.Amount,
					Price: buyOrder.Price,
				})
				buyOrder.Amount -= order.Amount
				
				if buyOrder.Amount == 0 {
					book.removeBuyOrder(i)
				}

				return trades
			}
			// fill a partial order and continue
			if buyOrder.Amount < order.Amount {
				trades = append(trades, Trade{
					TakerOrderId: order.Id,
					TakerUserId: order.UserId,
					MakerOrderId: buyOrder.Id,
					MakerUserId: buyOrder.UserId,
					Amount: buyOrder.Amount,
					Price: buyOrder.Price,
				})
				order.Amount -= buyOrder.Amount
				book.removeBuyOrder(i)
				continue
			}
		}
	}
	// finally add the remaining order to the list
	book.addSellOrder(order)
	return trades
}