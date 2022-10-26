package main

import (
	"encoding/json"
)

type User struct {
	ID     string `json:"id"`
	Email     string `json:"email"`
}

func (user *User) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, user)
}

func (user *User) ToJSON() []byte {
	str, _ := json.Marshal(user)
	return str
}
