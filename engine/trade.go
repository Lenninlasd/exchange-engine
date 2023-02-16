package engine

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Trade struct {
	TakerUserId  uuid.UUID `json:"taker_user_id"`
	TakerOrderId uuid.UUID `json:"taker_order_id"`
	MakerOrderId uuid.UUID `json:"maker_order_id"`
	MakerUserId  uuid.UUID `json:"maker_user_id"`
	Amount       float64   `json:"amount"`
	Price        float64   `json:"price"`
}

func (trade *Trade) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, trade)
}

func (trade *Trade) ToJSON() []byte {
	str, _ := json.Marshal(trade)
	return str
}
