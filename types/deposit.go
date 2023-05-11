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

type FiatDepositTx struct {
	Reference     string `json:"reference"`
	Id            string `json:"id"`
	AmountInCents uint64 `json:"amount_in_cents"`
	Status        string `json:"status"`
	UserId        string `json:"user_id"`
}

type WompiTx struct {
	ID                string `json:"id"`
	AmountInCents     uint64 `json:"amount_in_cents"`
	Reference         string `json:"reference"`
	CustomerEmail     string `json:"customer_email"`
	Currency          string `json:"currency"`
	PaymentMethodType string `json:"payment_method_type"`
	Status            string `json:"status"`
}

type WompiTransactionUpdatedEvent struct {
	Event string `json:"event"`
	Data  struct {
		Transaction WompiTx `json:"transaction"`
	} `json:"data"`
	Environment string `json:"environment"`
	Signature   struct {
		Properties []string `json:"properties"`
		Checksum   string   `json:"checksum"`
	} `json:"signature"`
	Timestamp int64  `json:"timestamp"`
	SentAt    string `json:"sent_at"`
}

func (w *WompiTransactionUpdatedEvent) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, w)
}
