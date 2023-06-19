package matchingengine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	engine := New()

	assert.Len(t, engine.AskOrders(), 0)
	assert.Len(t, engine.BidOrders(), 0)

	assert.ErrorContains(t, engine.Rollback(), "no bookmark")
}

func TestEngine_AddLimitOrder(t *testing.T) {
	engine := New()

	order := engine.AddLimitOrder(OrderTypeBuy, 100, 5)
	if assert.NotNil(t, order) {
		assert.Equal(t, 0, order.ID)
		assert.Equal(t, OrderTypeBuy, order.Type)
		assert.Equal(t, 100, order.Price)
		assert.Equal(t, 5, order.Amount)
		assert.Equal(t, OrderStatusPending, order.Status)
	}
}

func TestEngine_AddLimitOrder_Second(t *testing.T) {
	engine := New()

	_ = engine.AddLimitOrder(OrderTypeBuy, 100, 5)
	order := engine.AddLimitOrder(OrderTypeSell, 200, 50)
	if assert.NotNil(t, order) {
		assert.Equal(t, 1, order.ID)
		assert.Equal(t, OrderTypeSell, order.Type)
		assert.Equal(t, 200, order.Price)
		assert.Equal(t, 50, order.Amount)
	}
}

func TestEngine_BidOrders(t *testing.T) {
	engine := New()

	order := engine.AddLimitOrder(OrderTypeBuy, 100, 5)

	assert.EqualValues(t, []*Order{order}, engine.BidOrders())
	assert.Len(t, engine.AskOrders(), 0)
}

func TestEngine_AskOrders(t *testing.T) {
	engine := New()

	order := engine.AddLimitOrder(OrderTypeSell, 100, 5)
	assert.EqualValues(t, []*Order{order}, engine.AskOrders())
	assert.Len(t, engine.BidOrders(), 0)
}

func TestEngine_AskOrders_SamePrice(t *testing.T) {
	engine := New()

	order1 := engine.AddLimitOrder(OrderTypeSell, 100, 5)
	order2 := engine.AddLimitOrder(OrderTypeSell, 100, 10)
	asks := engine.AskOrders()

	assert.EqualValues(t, []*Order{order1, order2}, asks)
}

func TestEngine_AskOrders_OrderedPrices(t *testing.T) {
	engine := New()

	order1 := engine.AddLimitOrder(OrderTypeSell, 100, 5)
	order2 := engine.AddLimitOrder(OrderTypeSell, 50, 10)
	asks := engine.AskOrders()

	assert.EqualValues(t, []*Order{order2, order1}, asks)
}

func TestEngine_Order(t *testing.T) {
	engine := New()

	order1 := engine.AddLimitOrder(OrderTypeSell, 100, 5)
	order2 := engine.AddLimitOrder(OrderTypeSell, 50, 10)

	assert.Equal(t, order1, engine.Order(0))
	assert.Equal(t, order2, engine.Order(1))
	assert.Nil(t, engine.Order(99))
}

func TestEngine_Rollback_AddingSell(t *testing.T) {
	engine := New()

	engine.Bookmark()
	order := engine.AddLimitOrder(OrderTypeSell, 100, 5)

	assert.NoError(t, engine.Rollback())

	assert.Len(t, engine.AskOrders(), 0)
	assert.Len(t, engine.BidOrders(), 0)
	assert.Nil(t, engine.Order(0))
	assert.Equal(t, OrderStatusDeleted, order.Status)
}

func TestEngine_Rollback_Subsequent(t *testing.T) {
	engine := New()

	engine.Bookmark()
	order := engine.AddLimitOrder(OrderTypeSell, 100, 5)

	assert.NoError(t, engine.Rollback())

	assert.Len(t, engine.AskOrders(), 0)
	assert.Len(t, engine.BidOrders(), 0)
	assert.Nil(t, engine.Order(0))
	assert.Equal(t, OrderStatusDeleted, order.Status)
}

func TestEngine_Rollback_AddingBuy(t *testing.T) {
	engine := New()

	engine.Bookmark()
	order := engine.AddLimitOrder(OrderTypeBuy, 100, 5)

	assert.NoError(t, engine.Rollback())
	assert.NoError(t, engine.Rollback())

	assert.Len(t, engine.AskOrders(), 0)
	assert.Len(t, engine.BidOrders(), 0)
	assert.Nil(t, engine.Order(0))
	assert.Equal(t, OrderStatusDeleted, order.Status)
}

func TestEngine_CancelOrder(t *testing.T) {
	engine := New()
	order := engine.AddLimitOrder(OrderTypeBuy, 100, 5)
	engine.CancelOrder(order.ID)

	cancelledOrder := engine.Order(order.ID)

	assert.Len(t, engine.AskOrders(), 0)
	assert.Len(t, engine.BidOrders(), 0)

	if assert.NotNil(t, cancelledOrder) {
		assert.Equal(t, 0, cancelledOrder.ID)
		assert.Equal(t, OrderTypeBuy, cancelledOrder.Type)
		assert.Equal(t, 100, cancelledOrder.Price)
		assert.Equal(t, 5, cancelledOrder.Amount)
		assert.Equal(t, OrderStatusCancelled, cancelledOrder.Status)
	}
}

func TestEngine_CancelOrder_NonExistent(t *testing.T) {
	engine := New()
	order := engine.AddLimitOrder(OrderTypeBuy, 100, 5)
	engine.CancelOrder(55)

	cancelledOrder := engine.Order(order.ID)

	assert.Len(t, engine.AskOrders(), 0)
	assert.Len(t, engine.BidOrders(), 1)

	if assert.NotNil(t, cancelledOrder) {
		assert.Equal(t, 0, cancelledOrder.ID)
		assert.Equal(t, OrderTypeBuy, cancelledOrder.Type)
		assert.Equal(t, 100, cancelledOrder.Price)
		assert.Equal(t, 5, cancelledOrder.Amount)
		assert.Equal(t, OrderStatusPending, cancelledOrder.Status)
	}
}

func TestEngine_Rollback_CancelOrder_Buy(t *testing.T) {
	engine := New()

	order := engine.AddLimitOrder(OrderTypeBuy, 100, 5)

	engine.Bookmark()

	engine.CancelOrder(order.ID)

	assert.NoError(t, engine.Rollback())

	cancelledOrder := engine.Order(order.ID)
	assert.EqualValues(t, []*Order{cancelledOrder}, engine.BidOrders())
	if assert.NotNil(t, cancelledOrder) {
		assert.Equal(t, 0, cancelledOrder.ID)
		assert.Equal(t, OrderTypeBuy, cancelledOrder.Type)
		assert.Equal(t, 100, cancelledOrder.Price)
		assert.Equal(t, 5, cancelledOrder.Amount)
		assert.Equal(t, OrderStatusPending, cancelledOrder.Status)
	}
}

func TestEngine_Rollback_CancelOrder_Sell(t *testing.T) {
	engine := New()

	order := engine.AddLimitOrder(OrderTypeSell, 100, 5)

	engine.Bookmark()

	engine.CancelOrder(order.ID)

	assert.NoError(t, engine.Rollback())

	cancelledOrder := engine.Order(order.ID)
	assert.EqualValues(t, []*Order{cancelledOrder}, engine.AskOrders())
	if assert.NotNil(t, cancelledOrder) {
		assert.Equal(t, 0, cancelledOrder.ID)
		assert.Equal(t, OrderTypeSell, cancelledOrder.Type)
		assert.Equal(t, 100, cancelledOrder.Price)
		assert.Equal(t, 5, cancelledOrder.Amount)
		assert.Equal(t, OrderStatusPending, cancelledOrder.Status)
	}
}

func TestEngine_FullMatch_OnSell(t *testing.T) {
	engine := New()

	buyOrder := engine.AddLimitOrder(OrderTypeBuy, 100, 5)
	sellOrder := engine.AddLimitOrder(OrderTypeSell, 100, 5)

	// Should complete both orders
	assert.Len(t, engine.AskOrders(), 0)
	assert.Len(t, engine.BidOrders(), 0)

	assert.Equal(t, OrderStatusCompleted, buyOrder.Status)
	assert.Equal(t, OrderStatusCompleted, sellOrder.Status)

	assert.Equal(t, 5, buyOrder.Amount)
	assert.Equal(t, 5, sellOrder.Amount)

	assert.Equal(t, 5, buyOrder.CompletedAmount)
	assert.Equal(t, 5, sellOrder.CompletedAmount)
}

func TestEngine_FullMatch_OnBuy(t *testing.T) {
	engine := New()

	sellOrder := engine.AddLimitOrder(OrderTypeSell, 100, 5)
	buyOrder := engine.AddLimitOrder(OrderTypeBuy, 100, 5)

	// Should complete both orders
	assert.Len(t, engine.AskOrders(), 0)
	assert.Len(t, engine.BidOrders(), 0)

	assert.Equal(t, OrderStatusCompleted, buyOrder.Status)
	assert.Equal(t, OrderStatusCompleted, sellOrder.Status)

	assert.Equal(t, 5, buyOrder.Amount)
	assert.Equal(t, 5, sellOrder.Amount)

	assert.Equal(t, 5, buyOrder.CompletedAmount)
	assert.Equal(t, 5, sellOrder.CompletedAmount)
}

func TestEngine_PartialMatch_OnSell2(t *testing.T) {
	engine := New()

	buyOrder1 := engine.AddLimitOrder(OrderTypeBuy, 51, 2)
	buyOrder2 := engine.AddLimitOrder(OrderTypeBuy, 50, 1)
	buyOrder3 := engine.AddLimitOrder(OrderTypeBuy, 49, 10)
	sellOrder := engine.AddLimitOrder(OrderTypeSell, 50, 5)

	assert.Len(t, engine.AskOrders(), 1)
	assert.Len(t, engine.BidOrders(), 1)

	assert.Equal(t, OrderStatusCompleted, buyOrder1.Status)
	assert.Equal(t, OrderStatusCompleted, buyOrder2.Status)
	assert.Equal(t, OrderStatusPending, buyOrder3.Status)
	assert.Equal(t, OrderStatusPartiallyCompleted, sellOrder.Status)

	assert.Equal(t, 10, buyOrder3.RemainingAmount())
	assert.Equal(t, 2, sellOrder.RemainingAmount())

	assert.Equal(t, 2, buyOrder1.CompletedAmount)
	assert.Equal(t, 1, buyOrder2.CompletedAmount)
	assert.Equal(t, 0, buyOrder3.CompletedAmount)
	assert.Equal(t, 3, sellOrder.CompletedAmount)
}

func TestEngine_PartialMatch_OnSell(t *testing.T) {
	engine := New()

	buyOrder := engine.AddLimitOrder(OrderTypeBuy, 100, 5)
	sellOrder := engine.AddLimitOrder(OrderTypeSell, 100, 3)

	assert.Len(t, engine.AskOrders(), 0)
	assert.Len(t, engine.BidOrders(), 1)

	assert.Equal(t, OrderStatusPartiallyCompleted, buyOrder.Status)
	assert.Equal(t, OrderStatusCompleted, sellOrder.Status)

	assert.Equal(t, 5, buyOrder.Amount)
	assert.Equal(t, 3, sellOrder.Amount)

	assert.Equal(t, 3, buyOrder.CompletedAmount)
	assert.Equal(t, 3, sellOrder.CompletedAmount)
}

func TestEngine_Rollback_PartialMatch_OnSell(t *testing.T) {
	engine := New()

	buyOrder := engine.AddLimitOrder(OrderTypeBuy, 100, 5)
	engine.Bookmark()
	sellOrder := engine.AddLimitOrder(OrderTypeSell, 100, 3)

	assert.Nil(t, engine.Rollback())

	assert.Len(t, engine.AskOrders(), 0)
	assert.Len(t, engine.BidOrders(), 1)

	assert.Equal(t, OrderStatusDeleted, sellOrder.Status)
	assert.Equal(t, 0, buyOrder.ID)
	assert.Equal(t, OrderTypeBuy, buyOrder.Type)
	assert.Equal(t, 100, buyOrder.Price)
	assert.Equal(t, 5, buyOrder.Amount)
	assert.Equal(t, OrderStatusPending, buyOrder.Status)
}

func TestEngine_OverflowPartialMatch_OnSell(t *testing.T) {
	engine := New()

	buyOrder1 := engine.AddLimitOrder(OrderTypeBuy, 100, 5)
	buyOrder2 := engine.AddLimitOrder(OrderTypeBuy, 100, 2)
	sellOrder := engine.AddLimitOrder(OrderTypeSell, 100, 6)

	assert.Len(t, engine.AskOrders(), 0)
	assert.Len(t, engine.BidOrders(), 1)

	assert.Equal(t, OrderStatusCompleted, buyOrder1.Status)
	assert.Equal(t, OrderStatusPartiallyCompleted, buyOrder2.Status)
	assert.Equal(t, OrderStatusCompleted, sellOrder.Status)

	assert.Equal(t, 5, buyOrder1.CompletedAmount)
	assert.Equal(t, 1, buyOrder2.CompletedAmount)
	assert.Equal(t, 6, sellOrder.CompletedAmount)
}

func TestEngine_PartialSellFullBuy_OnSell(t *testing.T) {
	engine := New()

	buyOrder1 := engine.AddLimitOrder(OrderTypeBuy, 100, 5)
	buyOrder2 := engine.AddLimitOrder(OrderTypeBuy, 98, 2)
	sellOrder := engine.AddLimitOrder(OrderTypeSell, 100, 6)

	assert.Len(t, engine.AskOrders(), 1)
	assert.Len(t, engine.BidOrders(), 1)

	assert.Equal(t, OrderStatusCompleted, buyOrder1.Status)
	assert.Equal(t, OrderStatusPending, buyOrder2.Status)
	assert.Equal(t, OrderStatusPartiallyCompleted, sellOrder.Status)

	assert.Equal(t, 5, buyOrder1.CompletedAmount)
	assert.Equal(t, 0, buyOrder2.CompletedAmount)
	assert.Equal(t, 5, sellOrder.CompletedAmount)
}

func TestEngine_OverflowFullMatch_OnSell(t *testing.T) {
	engine := New()

	buyOrder1 := engine.AddLimitOrder(OrderTypeBuy, 100, 5)
	buyOrder2 := engine.AddLimitOrder(OrderTypeBuy, 100, 2)
	sellOrder := engine.AddLimitOrder(OrderTypeSell, 100, 10)

	assert.Len(t, engine.AskOrders(), 1)
	assert.Len(t, engine.BidOrders(), 0)

	assert.Equal(t, OrderStatusCompleted, buyOrder1.Status)
	assert.Equal(t, OrderStatusCompleted, buyOrder2.Status)
	assert.Equal(t, OrderStatusPartiallyCompleted, sellOrder.Status)

	assert.Equal(t, 5, buyOrder1.CompletedAmount)
	assert.Equal(t, 2, buyOrder2.CompletedAmount)
	assert.Equal(t, 7, sellOrder.CompletedAmount)
}
