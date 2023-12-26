package types

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Withdraw struct {
	Amount  uint64 `json:"amount"`
	Address string `json:"address"`
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
	Amount        uint64    `json:"amount"`
	Status        string    `json:"status"`
	Confirmations uint64    `json:"confirmations"`
	BlockHeight   uint64    `json:"blockheight"`
	Timestamp     time.Time `json:"timestamp"`
}

type WithdrawalRecord struct {
	Id        uuid.UUID `json:"id"`
	UserId    string    `json:"userid"`
	Email     string    `json:"email"`
	Amount    uint64    `json:"amount"`
	Address   string    `json:"address"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type UserRole struct {
	Role string `json:"role"`
}
