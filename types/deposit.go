package types

import (
	"encoding/json"
)

type DepositTx struct {
	Reference     string `json:"reference"`
	AmountInCents uint64 `json:"amount_in_cents"`
}

func (depositTx *DepositTx) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, depositTx)
}
