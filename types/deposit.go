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

type DepositEventTx struct {
	Reference     string `json:"reference"`
	AmountInCents uint64 `json:"amount_in_cents"`
}

type WompiTransactionUpdatedEvent struct {
	Event string `json:"event"`
	Data  struct {
		Transaction struct {
			ID                string      `json:"id"`
			AmountInCents     uint64      `json:"amount_in_cents"`
			Reference         string      `json:"reference"`
			CustomerEmail     string      `json:"customer_email"`
			Currency          string      `json:"currency"`
			PaymentMethodType string      `json:"payment_method_type"`
			RedirectURL       string      `json:"redirect_url"` // ❌
			Status            string      `json:"status"`
			ShippingAddress   interface{} `json:"shipping_address"`  // ❌
			PaymentLinkID     interface{} `json:"payment_link_id"`   // ❌
			PaymentSourceID   interface{} `json:"payment_source_id"` // ❌
		} `json:"transaction"`
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
