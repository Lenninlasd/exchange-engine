package types

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Balance struct {
	UserId    uuid.UUID `json:"user_id"`
	Timestamp time.Time `json:"time_stamp"`
	Amount    uint64    `json:"amount"`
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
	Amount     uint64
	AmountUSDT float64
}

type Withdraw struct {
	UserId  uuid.UUID `json:"user_id"`
	Amount  uint64    `json:"amount"`
	Address string    `json:"address"`
}

func (withdraw *Withdraw) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, withdraw)
}
func (withdraw *Withdraw) ToJSON() []byte {
	str, _ := json.Marshal(withdraw)
	return str
}

type Transaction struct {
	Txid     string `json:"txid"`
	Received uint64 `json:"received"`
	Sent     uint64 `json:"sent"`
	Fee      uint64 `json:"fee"`
}

func (transaction *Transaction) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, transaction)
}
