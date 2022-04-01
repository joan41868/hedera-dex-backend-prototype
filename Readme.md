# Event Indexing / Aggregation (Go Backend)

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

```go
package hedera_dex_backend

// persistable event, where the payload can be bytes or base64
type Event struct {
	Type int // swap, mint, pairCreated, etc.
	Payload []byte // The data which the event contains
	Created int64 // creation timestamp
}

// non-persistable event, as we cannot define serialization rule for interface{}
type TypedEvent struct {
    Payload interface{} // abstract enough to be able to contain different structures inside. 
}

// Example payload
type SwapEventPayload struct {
	From string
	To string
	TokenA string
	TokenB string
	AmountA float64 // or big.Int
	AmountB float64 // or big.Int
}

func (sep *SwapEventPayload) ToBytes() []byte {
	// conversion to bytes logic here..
	return []byte{}
}

func NewSwapEventPayloadFromBytes(b []byte) *SwapEventPayload {
    // conversion from bytes logic here...
    return &SwapEventPayload{}
}

func (e Event) ToTypedEvent() TypedEvent {
    return TypedEvent{
		// for the sake of the example, the event type is only 1 - swap
        Payload: NewSwapEventPayloadFromBytes(e.Payload),
    }
}
```

Example usage:

```go

	swpEvt := &SwapEventPayload{
		From:    "addr1",
		To:      "addr2",
		TokenA:  "tokenA",
		TokenB:  "tokenB",
		AmountA: 1.2,
		AmountB: 2.39,
	}
	evt := &Event{
		Type:    1, // we can imagine type 1 is a swap event for now
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
{"Type":1,"Payload":"XP+BAwEBEFN3YXBFdmVudFBheWxvYWQB/4IAAQYBBEZyb20BDAABAlRvAQwAAQZUb2tlbkEBDAABBlRva2VuQgEMAAEHQW1vdW50QQEIAAEHQW1vdW50QgEIAAAANf+CAQVhZGRyMQEFYWRkcjIBBnRva2VuQQEGdG9rZW5CAfgzMzMzMzPzPwH4H4XrUbgeA0AA","Created":1648822712}
```

-----------------------

typed_event.out.json:
```json
{"Payload":{"From":"addr1","To":"addr2","TokenA":"tokenA","TokenB":"tokenB","AmountA":1.2,"AmountB":2.39}}
```

-----------------------

### Contract


`ContractEventListener` - the entity which will poll the contracts for events through the mirror node.
There can be 1 listener per contract, or 1 per event type. (TODO: pros/cons)

~~Original idea~~ - posibility to add new contract event listeners dynamically.
Why was it rejected: overly complex solution, which would require significant development time.

**Current idea** - fixed amount of ContractEventListeners. They can be spawned when the app starts. They can utilize contract topics in order to filter events

```go

package contract_event_listeners

// ContractEventListener is the entity responsible for polling events out of contracts
type IContractEventListener interface {
	ProcessEvent(e Event) error
	Poll()
}

type ContractEventListener struct {
	// http client for the mirror node
	// db connection for persisting events
	// contract address
}

```

