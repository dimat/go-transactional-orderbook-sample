# Transactional Order Matching Engine Sample
## Simplifications
1. Engine works for the single market
2. All prices are in `int`, however, in real life it would be a decimal
3. No thread safety. Assumed that all the calls to the engine are serialised.
4. Rollback is assumed to be for internal purposes. It doesn't take into account 
states of orders.
5. There is no support of multiple concurrent transactions that can be committed or reverted. 
Not like in a relational DB.
6. I skipped "market order" and implemented only limit order
7. I kept the structure of packages flat to make it easier to navigate. However, it adds
verbosity to names.
8. The engine returns references to order structs without making copies.  It would be unsafe if it was used by
3rd parties, because the order could be modified. However, for purposes of this demo, I assumed that 
it would not be the case. 
9. I didn't add additional abstraction layers for storage layer, however, it would be easy to refactor having the tests.
10. When event insertion is rolled back, it is still kept in the memory with the "cancelled" status. 
However, for simplicity I implemented order IDs by auto-incremental integers, and after subsequent adding of a new order,
the cancelled one is overwritten in the cache.
11. Although the task said to implement the solution in two files, I split functionality
across multiple files for better readability and comprehension.
12. I used slices which wouldn't be efficient of constant insertions/deletions.
It would be better to use linked lists for an in-memory solution, or in real life
most probably it would be in an external storage.

## Design
Layers of abstraction:
- Engine events (public methods): place order, cancel order
- Internal events: match order
- Data events: add order, remove order, update order, delete order from the orderbook

## Order book
Keeps two lists of orders: asks and bids.
Each list is stored as a red-black tree to keep orders sorted by price.
The tree keeps price as a key and value is a FIFO queue of orders. 
Therefore, when there are two orders with the same price, the priority is given to the one that is placed first.

Additionally, orders are stored in a map to be able to find them by ID. 

## Event sourcing
All modifications the orderbook and orders are done by storing the sequence of data events.
Each data event has two methods: execute and revert.

Rollback is implemented on the data layer by reverting data events to a bookmarked position.
The history can be only linear, however, it would be possible to extend the functionality depending on the needs.

## Running
From the directory `matchingengine`, run the following commands:
```sh
go mod download
```

And to run the tests:
```sh
go test . -v
```
