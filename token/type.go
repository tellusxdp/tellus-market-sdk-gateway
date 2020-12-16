package token

type JWTPayload struct {
	Audience  string `json:"aud,omitempty"`
	ID        string `json:"jti,omitempty"`
	Issuer    string `json:"iss,omitempty"`
	Subject   string `json:"sub,omitempty"`
	ProductID string `json:"product_id"`
	AuthType  string `json:"auth_type"`
}
