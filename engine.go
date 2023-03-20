package matchingengine

import (
	"errors"
)

type Engine interface {
	AddLimitOrder(orderType OrderType, price int, amount int) *Order
	CancelOrder(OrderID)

	AskOrders() []*Order
	BidOrders() []*Order

	Order(OrderID) *Order

	Bookmark()
	Rollback() error
}

type engine struct {
	orderbook  *orderbook
	dataEvents []dataEvent
	bookmark   int

	lastOrderID OrderID
}

func New() Engine {
	return &engine{
		orderbook: newOrderBook(),
		bookmark:  -1,
	}
}

func (e *engine) AddLimitOrder(orderType OrderType, price int, amount int) *Order {
	e.addDataEvent(newAddLimitOrderDataEvent(orderType, price, amount))
	order := e.Order(e.lastOrderID - 1)
	e.matchBuyOrder(order)
	return order
}

func (e *engine) CancelOrder(orderID OrderID) {
	order := e.Order(orderID)
	if order == nil {
		return // ignoring
	}

	e.addDataEvent(newUpdateOrderDataEvent(orderID, orderUpdates{
		OldStatus: order.Status,
		NewStatus: OrderStatusCancelled,
	}))
	e.addDataEvent(newRemoveOrderDataEvent(order))
}

func (e *engine) addDataEvent(event dataEvent) {
	event.Execute(e)

	e.dataEvents = append(e.dataEvents, event)
}

func (e *engine) Bookmark() {
	e.bookmark = len(e.dataEvents)
}

func (e *engine) Rollback() error {
	if e.bookmark < 0 {
		return errors.New("no bookmark")
	}

	if e.bookmark >= len(e.dataEvents) {
		return nil
	}

	var event dataEvent
	for len(e.dataEvents) > e.bookmark {
		event = e.dataEvents[len(e.dataEvents)-1]
		e.dataEvents = e.dataEvents[:len(e.dataEvents)-1]

		event.Revert(e)
	}
	return nil
}

func (e *engine) AskOrders() []*Order {
	return e.orderbook.AskOrders()
}

func (e *engine) BidOrders() []*Order {
	return e.orderbook.BidOrders()
}

func (e *engine) Order(id OrderID) *Order {
	return e.orderbook.orders[id]
}

func (e *engine) matchBuyOrder(order *Order) {
	processingOrders := e.orderbook.asks
	if order.Type == OrderTypeSell {
		processingOrders = e.orderbook.bids
	}
	iter := processingOrders.Iterator()

mainLoop:
	for iter.Next() {
		oppositeOrderQueue := iter.Value().([]*Order)
		for _, oppositeOrder := range oppositeOrderQueue {
			if oppositeOrder.Price > order.Price {
				break mainLoop
			}

			closingAmount := order.RemainingAmount()
			if oppositeOrder.RemainingAmount() < order.RemainingAmount() {
				closingAmount = oppositeOrder.RemainingAmount()
			}

			e.closeAmountInOrder(oppositeOrder, closingAmount)
			e.closeAmountInOrder(order, closingAmount)
		}
	}
}

func (e *engine) closeAmountInOrder(order *Order, closingAmount int) {
	newStatus := OrderStatusPartiallyCompleted
	if closingAmount+order.CompletedAmount == order.Amount {
		newStatus = OrderStatusCompleted
	}

	oppositeOrderUpdates := orderUpdates{
		AmountDelta: closingAmount,
		OldStatus:   order.Status,
		NewStatus:   newStatus,
	}

	e.addDataEvent(newUpdateOrderDataEvent(order.ID, oppositeOrderUpdates))
	if newStatus == OrderStatusCompleted {
		e.addDataEvent(newRemoveOrderDataEvent(order))
	}
}
