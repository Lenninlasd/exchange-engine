package engine

import (
	"encoding/json"
	"github.com/gocql/gocql"
)

type Order struct {
	Id     gocql.UUID	`json:"id"`
	UserId gocql.UUID	`json:"user_id"`
	Amount float64		`json:"amount"`
	Price  float64		`json:"price"`
	Side   int8			`json:"side"`
}

func (order *Order) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, order)
}

func (order *Order) ToJSON() []byte {
	str, _ := json.Marshal(order)
	return str
}
