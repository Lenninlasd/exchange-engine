package types

import (
	"time"
	"encoding/json"
	"github.com/gocql/gocql"
)

type Balance struct {
	UserId		gocql.UUID 	`json:"user_id"`
	Timestamp	time.Time  	`json:"time_stamp"`
	Amount		float64 	`json:"amount"`
}

func (balance *Balance) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, balance)
}

func (balance *Balance) ToJSON() []byte {
	str, _ := json.Marshal(balance)
	return str
}


type UserBalances struct {
	Id		gocql.UUID
	Amount		float64
	AmountUSDT	float64
}
