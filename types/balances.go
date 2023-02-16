package types

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Balance struct {
	UserId    uuid.UUID `json:"user_id"`
	Timestamp time.Time `json:"time_stamp"`
	Amount    float64   `json:"amount"`
}

func (balance *Balance) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, balance)
}

func (balance *Balance) ToJSON() []byte {
	str, _ := json.Marshal(balance)
	return str
}

type UserBalances struct {
	Id         uuid.UUID
	Amount     float64
	AmountUSDT float64
}
