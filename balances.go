package main

import (
	"encoding/json"
)

type Banlace struct {
	UserId		string `json:"id"`
	CoinType	string `json:"coin_type"`
	Amount		uint64 `json:"amount"`
	Unix		int64  `json:"timestamp"`
}

func (banlace *Banlace) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, banlace)
}

func (banlace *Banlace) ToJSON() []byte {
	str, _ := json.Marshal(banlace)
	return str
}
