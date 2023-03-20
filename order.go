package matchingengine

type OrderID = int

type OrderType = string

const (
	OrderTypeBuy  OrderType = "buy"
	OrderTypeSell OrderType = "sell"
)

type OrderStatus = string

const (
	OrderStatusPending            OrderStatus = "pending"
	OrderStatusPartiallyCompleted OrderStatus = "partially_completed"
	OrderStatusCompleted          OrderStatus = "completed"
	OrderStatusCancelled          OrderStatus = "cancelled"
	OrderStatusDeleted            OrderStatus = "deleted"
)

type Order struct {
	ID              OrderID
	Amount          int
	CompletedAmount int
	Price           int
	Type            OrderType
	Status          OrderStatus
}

func (o Order) RemainingAmount() int {
	return o.Amount - o.CompletedAmount
}
