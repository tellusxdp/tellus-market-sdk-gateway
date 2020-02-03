package token

import ()

type JWTPayload struct {
	Audience string `json:"aud,omitempty"`
	ID       string `json:"jti,omitempty"`
	Issuer   string `json:"iss,omitempty"`
	Subject  string `json:"sub,omitempty"`
	ToolID   string `json:"tool_id"`
}
