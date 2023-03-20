package matchingengine

type dataEventRemoveOrderEvent struct {
	order *Order
}

func (d dataEventRemoveOrderEvent) Execute(e *engine) {
	e.orderbook.deleteOrderFromBook(d.order)
}

func (d dataEventRemoveOrderEvent) Revert(e *engine) {
	e.orderbook.addOrderToBook(d.order)
}

func newRemoveOrderDataEvent(order *Order) dataEvent {
	return &dataEventRemoveOrderEvent{
		order: order,
	}
}
