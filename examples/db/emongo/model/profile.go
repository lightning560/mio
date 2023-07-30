package model

import (
	"encoding/hex"
)

// User .
type Profile struct {
	Mid    int64  `json:"mid"`
	Name   string `json:"name,omitempty"`
	Sex    int8   `json:"sex"`
	Status int8   `json:"status"`
	Phone  string `json:"phone,omitempty"`
	Face   string `json:"face,omitempty"`
	Level  int    `json:"level"`
	Cid    string `json:"cid,omitempty"`
	Email  []byte `json:"email,omitempty"`
	Sign   string `json:"sign,omitempty"`
}

// DecodeUser .
type DecodeProfile struct {
	Mid    int64  `json:"mid"`
	Name   string `json:"name"`
	Sex    int8   `json:"sex"`
	Status int8   `json:"status"`
	Phone  string `json:"Phone"`
	Face   string `json:"face"`
	Level  int    `json:"level"`
	Cid    string `json:"cid"`
	Email  string `json:"email"`
}

// Decode decode user
func (d *Profile) Decode() *DecodeProfile {
	return &DecodeProfile{
		Mid:    d.Mid,
		Name:   d.Name,
		Sex:    d.Sex,
		Status: d.Status,
		Phone:  d.Phone,
		Face:   d.Face,
		Level:  d.Level,
		Cid:    d.Cid,
		Email:  hex.EncodeToString(d.Email),
	}
}
