package types

import (
	"encoding/json"
)

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
}

func (user *User) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, user)
}

func (user *User) ToJSON() []byte {
	str, _ := json.Marshal(user)
	return str
}
