package types

import (
	"encoding/json"
)

type User struct {
	Id         string          `json:"id"`
	Email      string          `json:"email"`
	ReferralId string          `json:"referralid"`
	Kyc        json.RawMessage `json:"kyc"`
}

func (user *User) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, user)
}

func (user *User) ToJSON() []byte {
	str, _ := json.Marshal(user)
	return str
}

type Kyc struct {
	Identity struct {
		Number         string `json:"number"`
		IssuingCountry string `json:"issuingCountry"`
	} `json:"identity"`
	Residence struct {
		CountryOfResidence string `json:"countryOfResidence"`
		Departamento       string `json:"departamento"`
		City               string `json:"city"`
		Address            string `json:"address"`
	} `json:"residence"`
	Phone string `json:"phone"`
}

func (kyc *Kyc) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, kyc)
}

func (kyc *Kyc) ToJSON() ([]byte, error) {
	return json.Marshal(kyc)
}
