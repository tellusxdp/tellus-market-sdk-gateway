package token

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var publicKeys map[string]string = map[string]string{}

type TokenError struct {
	Status  int
	Message string
}

func (e TokenError) Error() string {
	return e.Message
}

func NewTokenError(status int, message string) TokenError {
	return TokenError{status, message}
}

func updatePublicKeys(url string) error {
	log.Debugf("Download public keys from %s", url)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Invalid respons status %d", resp.StatusCode)
	}

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	keys := map[string]string{}
	err = json.Unmarshal(byteArray, &keys)
	if err != nil {
		return err
	}

	log.Infof("Download %d public keys from %s", len(keys), url)
	publicKeys = keys
	return nil
}

func ValidateToken(tokenString string, publicKeysURL string) (*JWTPayload, error) {
	updatePublicKeys(publicKeysURL)

	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// check signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("Unexpected signing method")
		}

		kid, ok := token.Header["kid"]
		if !ok {
			return nil, errors.New("kid not found")
		}

		kidStr, ok := kid.(string)
		if !ok {
			return nil, errors.New("Invalid kid")
		}
		publicKeyStr, ok := publicKeys[kidStr]
		if !ok {
			return nil, fmt.Errorf("Unknown kid %s", kidStr)
		}

		publicKeyBlock, _ := pem.Decode([]byte(publicKeyStr))
		if publicKeyBlock == nil {
			return nil, errors.New("Public key cannot decode")
		}
		if publicKeyBlock.Type != "PUBLIC KEY" {
			return nil, errors.New("Public key type is invalid")
		}

		publicKey, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
		if err != nil {
			return nil, errors.New("Failed to parse public key")
		}

		return publicKey, nil
	})

	if err != nil {
		err = errors.Wrap(err, "Token is invalid")
		return nil, err
	}
	if !parsedToken.Valid {
		return nil, errors.New("Token is invalid")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("Invalid claims")
	}

	t := &JWTPayload{
		Audience: claims["aud"].(string),
		ID:       claims["jti"].(string),
		Issuer:   claims["iss"].(string),
		Subject:  claims["sub"].(string),
		ToolID:   claims["product"].(string),
	}
	return t, nil
}
