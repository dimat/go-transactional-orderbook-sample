package matchingengine

import (
	rbt "github.com/emirpasic/gods/trees/redblacktree"
	"golang.org/x/exp/slices"
)

type orderbook struct {
	asks *rbt.Tree
	bids *rbt.Tree

	orders map[OrderID]*Order
}

func newOrderBook() *orderbook {
	return &orderbook{
		asks:   rbt.NewWithIntComparator(),
		bids:   rbt.NewWithIntComparator(),
		orders: make(map[OrderID]*Order),
	}
}

func (ob orderbook) addOrderToBook(order *Order) {
	orderList := ob.orderList(order.Type)
	queue, found := orderList.Get(order.Price)
	if !found {
		queue = []*Order{order}
	} else {
		queue = append(queue.([]*Order), order)
	}

	orderList.Put(order.Price, queue)
}

func (ob orderbook) deleteOrderFromBook(order *Order) {
	orderList := ob.orderList(order.Type)
	queue, found := orderList.Get(order.Price)
	if !found {
		panic("corrupted state of orderList")
	}
	ordersQueue := queue.([]*Order)
	idx := slices.Index(ordersQueue, order)
	if idx < 0 {
		panic("corrupted state of orderQueue in orderList")
	}
	ordersQueue = slices.Delete(ordersQueue, idx, idx+1)
	if len(ordersQueue) == 0 {
		orderList.Remove(order.Price)
	} else {
		orderList.Put(order.Price, ordersQueue)
	}
}

func (ob orderbook) orderList(orderType OrderType) *rbt.Tree {
	// assuming the input is valid
	if orderType == OrderTypeBuy {
		return ob.bids
	}
	return ob.asks
}

func (ob orderbook) AskOrders() []*Order {
	return unrollOrderTree(ob.asks)
}

func (ob orderbook) BidOrders() []*Order {
	return unrollOrderTree(ob.bids)
}

func unrollOrderTree(tree *rbt.Tree) []*Order {
	var result []*Order

	queues := tree.Values()
	for _, queue := range queues {
		orderQueue := queue.([]*Order)
		result = append(result, orderQueue...)
	}
	return result
}
