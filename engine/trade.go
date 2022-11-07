package engine

import (
	"encoding/json"
	"github.com/gocql/gocql"
)

type Trade struct {
	TakerOrderID gocql.UUID `json:"taker_order_id"`
	MakerOrderID gocql.UUID `json:"maker_order_id"`
	Amount       float64 	`json:"amount"`
	Price        float64 	`json:"price"`
}

func (trade *Trade) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, trade)
}

func (trade *Trade) ToJSON() []byte {
	str, _ := json.Marshal(trade)
	return str
}