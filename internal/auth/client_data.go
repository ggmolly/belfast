package auth

import (
	"encoding/base64"
	"encoding/json"
)

func ExtractChallenge(clientDataBase64 string) (string, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(clientDataBase64)
	if err != nil {
		return "", err
	}
	var payload struct {
		Challenge string `json:"challenge"`
	}
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return "", err
	}
	return payload.Challenge, nil
}
