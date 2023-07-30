package model

import (
	"encoding/hex"
)

// User .
type User struct {
	Mid      int64  `json:"mid"`
	UserName string `json:"username"`
	Pwd      []byte `json:"pwd"`
	// Pwd    string `json:"pwd"`
	Salt   string `json:"salt"`
	Status int8   `json:"status"`
	Tel    []byte `json:"tel"`
	Cid    string `json:"cid"`
	Email  []byte `json:"email"`
}

// DecodeUser .
type DecodeUser struct {
	Mid      int64  `json:"mid"`
	UserName string `json:"username"`
	Pwd      string `json:"pwd"`
	Salt     string `json:"salt"`
	Status   int8   `json:"status"`
	Tel      string `json:"tel"`
	Cid      string `json:"cid"`
	Email    string `json:"email"`
}

// Decode decode user
func (d *User) Decode() *DecodeUser {
	return &DecodeUser{
		Mid:      d.Mid,
		UserName: d.UserName,
		Pwd:      hex.EncodeToString(d.Pwd),
		Salt:     d.Salt,
		Status:   d.Status,
		Tel:      hex.EncodeToString(d.Tel),
		Cid:      d.Cid,
		Email:    hex.EncodeToString(d.Email),
	}
}
