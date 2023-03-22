package types

import (
	"time"
)

type BroadcastedTx struct {
	Txid          string    `json:"txid"`
	Address       string    `json:"address"`
	Status        string    `json:"status"`
	Confirmations uint64    `json:"confirmations"`
	Timestamp     time.Time `json:"timestamp"`
}
