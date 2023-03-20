package matchingengine

type dataEventAddLimitOrder struct {
	orderType OrderType
	price     int
	amount    int
}

func newAddLimitOrderDataEvent(orderType OrderType, price int, amount int) dataEvent {
	return &dataEventAddLimitOrder{
		orderType: orderType,
		price:     price,
		amount:    amount,
	}
}

func (d dataEventAddLimitOrder) Execute(e *engine) {
	order := &Order{
		ID:     e.lastOrderID,
		Amount: d.amount,
		Price:  d.price,
		Type:   d.orderType,
		Status: OrderStatusPending,
	}
	e.lastOrderID++

	e.orderbook.addOrderToBook(order)
	e.orderbook.orders[order.ID] = order
}

func (d dataEventAddLimitOrder) Revert(e *engine) {
	order := e.Order(e.lastOrderID - 1)
	e.lastOrderID--

	order.Status = OrderStatusDeleted

	e.orderbook.deleteOrderFromBook(order)
	delete(e.orderbook.orders, order.ID)
}
