# Event Indexing / Aggregation (Go Backend)

## Abstract

![](./assets/dex-be-diagram1.png)

## DB:
- PostgreSQL - will require us to define a strict schema for upcoming events.
- MongoDB - not so strict, but also not very scalable. We will start to experience it's bottleneck after 2-3 milion records
- Cassandra - super scalable database, but may be an overkill

## Service (App)

Go Backend

1. Will poll for events on X smart contracts (via the mirror node rest api)
2. Will persist the data from those events in the DB
3. Will be responsible for creating and maintaining indexes on the database
4. Will be able to serve those events through a rest API


### Events

We must listen for the following events:

##### Pair Events
- `Swap` - for each pair; Will help for determining volume and the most-traded pairs.
- `Mint` - for each pair (add liquidity) - most liquid pairs
- `Burn` - for each pair (remove liquidity) ?
- `Sync` - for each pair. Represents update in the reserves of the pair.
##### Other events

- `PairCreated` - this event should start a new poller for the above events
- ...


```go
package hedera_dex_backend

type EventType int

const (
	EventTypeSwap EventType = iota
	EventTypeMint
	EventTypeBurn
	EventTypeSync
	EventTypePairCreated
	// other notable event types
)

type AbstractPersistableEvent struct {
	Type EventType // swap, mint, pairCreated, etc.
	Payload []byte // The data which the event contains. Byte array can be used for any event type.
	Created int64 // creation timestamp
}

// non-persistable event, as we cannot define serialization rule for interface{}
type TypedEvent struct {
    Payload interface{} // abstract enough to be able to contain different structures inside. 
}

// Example payload - from a Swap event
type SwapEventPayload struct {
	Sender      string
	Recipient   string
	
	Amount0In   float64
	Amount1In   float64
	
	Amount0Out  float64
	Amount1Out  float64
}
```

Example usage:

```go

	swpEvt := &SwapEventPayload{
		Sender:    "addr1",
		Recipient: "addr2",
		Amount0In: 1.2,
		Amount1In: 2.39,
		Amount0Out: 3.4,
		Amount1Out: 4.5,
	}
	evt := &Event{
		Type:    EventTypeSwap, // we can imagine type 1 is a swap event for now
		Payload: swpEvt.ToBytes(),
		Created: time.Now().Unix(),
	}

	f, _ := os.OpenFile("raw_event.out.json", os.O_RDWR|os.O_CREATE, 0755)
	defer f.Close()

	// write the raw event
	json.NewEncoder(f).Encode(evt)

	f2, _ := os.OpenFile("typed_event.out.json", os.O_RDWR|os.O_CREATE, 0755)
	defer f2.Close()

	// write the typed event with interface{} payload
	json.NewEncoder(f2).Encode(evt.ToTypedEvent())


```
-----------------------
raw_event.out.json:
```json
{"Type":0,"Payload":"cf+BAwEBEFN3YXBFdmVudFBheWxvYWQB/4IAAQYBBlNlbmRlcgEMAAEJUmVjaXBpZW50AQwAAQlBbW91bnQwSW4BCAABCUFtb3VudDFJbgEIAAEKQW1vdW50ME91dAEIAAEKQW1vdW50MU91dAEIAAAAM/+CAQVhZGRyMQEFYWRkcjIB+DMzMzMzM/M/AfgfhetRuB4DQAH4MzMzMzMzC0AB/hJAAA==","Created":1648914703,"Contract":"0xSomeContract"}
```

-----------------------

typed_event.out.json:
```json
{"Payload":{"Sender":"addr1","Recipient":"addr2","Amount0In":1.2,"Amount1In":2.39,"Amount0Out":3.4,"Amount1Out":4.5},"Contract":"0xSomeContract"}
```

SQL queries:

```sql
/* Assuming 4 is EventTypePairCreated */
/* This query will give us the latest pairs which were created */
SELECT * FROM events WHERE eventType = 4 ORDER BY created DESC;

/*  This query will give us the most liquid pairs */
SELECT COUNT(contractAddress) as cnt, contractAddress WHERE eventType = 1 GROUP BY contractAddress ORDER BY cnt DESC;
```

-----------------------

### Contract

~~Original idea~~ - posibility to add new contract event listeners dynamically.
Why was it rejected: overly complex solution, which would require significant development time.

**Current idea** - fixed amount of ContractEventListeners. They can be spawned when the app starts. They can utilize contract topics in order to filter events

`ContractEventListener` - the entity which will poll the contracts for events through the mirror node.
There should be 1 listener for `PairCreated`, and 3 new listeners(pollers) should be created for each new pair.


```go

package contract_event_listeners

type Event interface{}

// IContractEventListener is the entity responsible for polling events out of contracts
type IContractEventListener interface {
	ProcessEvent(e Event) error
	Poll()
}

type ContractEventListener struct {
	// http client for the mirror node
	// db connection for persisting events
	// contract address
	// event topic (optional)
}

```

### Queries

GET /pairs - should return all pairs (pools in uniswap)

GET /pairs/{symbol|address} - should return info for specific pair

GET /pairs/top - should return top pairs(pools), ordered by TVL/Volume




