package matchingengine

type dataEventUpdateOrder struct {
	orderID OrderID
	updates orderUpdates
}

type orderUpdates struct {
	OldStatus   OrderStatus
	NewStatus   OrderStatus
	AmountDelta int
}

func newUpdateOrderDataEvent(id OrderID, updates orderUpdates) dataEvent {
	return &dataEventUpdateOrder{
		orderID: id,
		updates: updates,
	}
}

func (d dataEventUpdateOrder) Execute(e *engine) {
	order := e.Order(d.orderID)

	if d.updates.NewStatus != "" {
		order.Status = d.updates.NewStatus
	}

	order.CompletedAmount += d.updates.AmountDelta
}

func (d dataEventUpdateOrder) Revert(e *engine) {
	order := e.Order(d.orderID)

	if d.updates.OldStatus != "" {
		order.Status = d.updates.OldStatus
	}

	order.CompletedAmount -= d.updates.AmountDelta
}
