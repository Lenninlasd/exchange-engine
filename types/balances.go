package types

import (
	"time"
	"encoding/json"
	"github.com/gocql/gocql"
)

type Balance struct {
	Id			gocql.UUID 	`json:"id"`
	UserId		gocql.UUID 	`json:"user_id"`
	CoinType	string 		`json:"coin_type"`
	Amount		int64 		`json:"amount"`
	Timestamp	time.Time  	`json:"time_stamp"`
}

func (balance *Balance) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, balance)
}

func (balance *Balance) ToJSON() []byte {
	str, _ := json.Marshal(balance)
	return str
}
