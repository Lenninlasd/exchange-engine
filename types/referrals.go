package types

import "encoding/json"

type ReferrerIdBody struct {
	ReferrerId string `json:"referrer_id"`
}

func (referrerIdBody *ReferrerIdBody) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, referrerIdBody)
}
