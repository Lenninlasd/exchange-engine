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

type Auth0User struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

func (auth0User *Auth0User) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, auth0User)
}
