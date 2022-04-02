package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"time"
)

type EventType int

const (
	EventTypeSwap EventType = iota
	EventTypeTransfer
	EventTypePairCreated
	EventTypeMint
	EventTypeBurn
)

// persistable event, where the payload can be bytes or base64
type Event struct {
	*gorm.Model
	Type     EventType // swap, mint, pairCreated, etc.
	Payload  []byte    // The data which the event contains
	Created  int64     // creation timestamp
	Contract string
}

// non-persistable event, as we cannot define serialization rule for interface{}
type TypedEvent struct {
	Payload  interface{} // abstract enough to be able to contain different structures inside.
	Contract string
}

// Example payload
type SwapEventPayload struct {
	Sender    string
	Recipient string

	Amount0In float64
	Amount1In float64

	Amount0Out float64
	Amount1Out float64
}

type TransferEventPayload struct {
	From  string
	To    string
	Token string
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
	var p interface{}
	if e.Type == EventTypeSwap {
		p, _ = NewSwapEventFromBytes(e.Payload)
	} else if e.Type == EventTypeTransfer {
		p, _ = NewTransferEventFromBytes(e.Payload)
	}
	return TypedEvent{
		Payload:  p,
		Contract: e.Contract,
	}
}

func NewTransferEventFromBytes(payload []byte) (TransferEventPayload, error) {
	buf := bytes.NewBuffer(payload)
	dec := gob.NewDecoder(buf)
	var tep TransferEventPayload
	err := dec.Decode(&tep)
	if err != nil {
		return TransferEventPayload{}, err
	}
	return tep, nil
}

func (tep *TransferEventPayload) ToBytes() []byte {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	enc.Encode(tep)
	return buf.Bytes()
}

func main() {

	swpEvt := &SwapEventPayload{
		Sender:     "addr1",
		Recipient:  "addr2",
		Amount0In:  1.2,
		Amount1In:  2.39,
		Amount0Out: 3.4,
		Amount1Out: 4.5,
	}
	evt := &Event{
		Type:     EventTypeSwap, // we can imagine type 1 is a swap event for now
		Payload:  swpEvt.ToBytes(),
		Created:  time.Now().Unix(),
		Contract: "0xSomeContract",
	}

	f, _ := os.OpenFile("raw_event.out.json", os.O_RDWR|os.O_CREATE, 0755)
	defer f.Close()

	// write the raw event
	json.NewEncoder(f).Encode(evt)

	f2, _ := os.OpenFile("typed_event.out.json", os.O_RDWR|os.O_CREATE, 0755)
	defer f2.Close()

	// write the typed event with interface{} payload
	json.NewEncoder(f2).Encode(evt.ToTypedEvent())

	//db := getDB()
	//swpEvt := &SwapEventPayload{
	//	From:    "0x_addr1",
	//	To:      "0x_addr2",
	//	TokenA:  "0x_tokenA",
	//	TokenB:  "0x_tokenB",
	//	AmountA: 1.2,
	//	AmountB: 2.39,
	//}
	//transferEvt := &TransferEventPayload{
	//	From:  "0x_addr1",
	//	To:    "0x_addr2",
	//	Token: "0x_tokenA",
	//}
	//evt := &Event{
	//	Type:     EventTypeSwap,
	//	Payload:  swpEvt.ToBytes(),
	//	Created:  time.Now().Unix(),
	//	Contract: "0x_contract1",
	//}
	//evt2 := &Event{
	//	Type:     EventTypeTransfer,
	//	Payload:  transferEvt.ToBytes(),
	//	Created:  time.Now().Unix(),
	//	Contract: "0x_contract2",
	//}
	//
	//err := db.Create(evt).Error
	//if err != nil {
	//	panic(err)
	//}
	//
	//err = db.Create(evt2).Error
	//
	//var events []Event
	//db.Find(&events)
	//for _, e := range events {
	//	json.NewEncoder(os.Stdout).Encode(e.ToTypedEvent())
	//}
}

func getDB() *gorm.DB {
	pg := "postgres"
	str := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s",
		"localhost", "5432", pg, pg, pg)
	db, err := gorm.Open(postgres.Open(str), &gorm.Config{})
	if err != nil {
		db, err = gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")))
		if err != nil {
			panic(err)
		}
	}
	db.AutoMigrate(&Event{})
	return db
}
