package token

import ()

type JWTPayload struct {
	Audience string `json:"aud,omitempty"`
	Id       string `json:"jti,omitempty"`
	Issuer   string `json:"iss,omitempty"`
	Subject  string `json:"sub,omitempty"`
	Product  string `json:"product"`
}
