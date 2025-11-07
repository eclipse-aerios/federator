package models

import "time"

type AeriosShimToken struct {
	Token string `json:"token"`
}

type KeycloakAccessToken struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in,omitempty"`
	ExpiresAt        time.Time
	RefreshExpiresIn int    `json:"refresh_expires_in,omitempty"`
	TokenType        string `json:"token_type,omitempty"`
	NotBeforePolicy  int    `json:"not-before-policy,omitempty"`
	Scope            string `json:"scope,omitempty"`
}

func (token KeycloakAccessToken) IsTokenExpired() bool {
	return time.Now().After(token.ExpiresAt)
}
