package types

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

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

type BroadcastedTx struct {
	Txid          string    `json:"txid"`
	Address       string    `json:"address"`
	Status        string    `json:"status"`
	Confirmations uint64    `json:"confirmations"`
	BlockHeight   uint64    `json:"blockheight"`
	Timestamp     time.Time `json:"timestamp"`
}
