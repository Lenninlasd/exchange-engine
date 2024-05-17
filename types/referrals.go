package types

import (
	"encoding/json"
	"time"
)

type ReferrerIdBody struct {
	ReferrerId string `json:"referrer_id"`
}

func (referrerIdBody *ReferrerIdBody) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, referrerIdBody)
}

type ReferralHistory struct {
	Id                string    `json:"id"`
	UserId            string    `json:"userId"`
	ReferredUserId    string    `json:"referredUserId"`
	ReferredUserEmail string    `json:"referredUserEmail"`
	PurchasedSats     uint64    `json:"purchasedSats"`
	AmountInCents     uint64    `json:"amountInCents"`
	Timestamp         time.Time `json:"timestamp"`
}
