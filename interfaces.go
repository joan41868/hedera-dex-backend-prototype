package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"os"
	"time"
)

// persistable event, where the payload can be bytes or base64
type Event struct {
	Type    int    // swap, mint, pairCreated, etc.
	Payload []byte // The data which the event contains
	Created int64  // creation timestamp
}

// non-persistable event, as we cannot define serialization rule for interface{}
type TypedEvent struct {
	Payload interface{} // abstract enough to be able to contain different structures inside.
}

// Example payload
type SwapEventPayload struct {
	From    string
	To      string
	TokenA  string
	TokenB  string
	AmountA float64 // or big.Int
	AmountB float64 // or big.Int
}

func (sep *SwapEventPayload) ToBytes() []byte {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	enc.Encode(sep)
	return buf.Bytes()
}

func NewSwapEventFromBytes(data []byte) (*SwapEventPayload, error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	var sep SwapEventPayload
	err := dec.Decode(&sep)
	if err != nil {
		return nil, err
	}
	return &sep, nil
}

func (e Event) ToTypedEvent() TypedEvent {
	p, _ := NewSwapEventFromBytes(e.Payload)
	return TypedEvent{
		Payload: p,
	}
}

func main() {
	swpEvt := &SwapEventPayload{
		From:    "addr1",
		To:      "addr2",
		TokenA:  "tokenA",
		TokenB:  "tokenB",
		AmountA: 1.2,
		AmountB: 2.39,
	}
	evt := &Event{
		Type:    1,
		Payload: swpEvt.ToBytes(),
		Created: time.Now().Unix(),
	}

	f, _ := os.OpenFile("raw_event.out.json", os.O_RDWR|os.O_CREATE, 0755)
	defer f.Close()

	json.NewEncoder(f).Encode(evt)

	f2, _ := os.OpenFile("typed_event.out.json", os.O_RDWR|os.O_CREATE, 0755)
	defer f2.Close()

	json.NewEncoder(f2).Encode(evt.ToTypedEvent())
}
