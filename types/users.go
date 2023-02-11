package types

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gocql/gocql"
)

type User struct {
	Id     uuid.UUID `json:"id"`
	Email     string `json:"email"`
}

func (user *User) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, user)
}

func (user *User) ToJSON() []byte {
	str, _ := json.Marshal(user)
	return str
}

type UserC struct {
	Id     gocql.UUID `json:"id"`
	Email     string `json:"email"`
}

func (user *UserC) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, user)
}

func (user *UserC) ToJSON() []byte {
	str, _ := json.Marshal(user)
	return str
}

