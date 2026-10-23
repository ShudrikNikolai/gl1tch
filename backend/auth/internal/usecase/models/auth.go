package models

import (
	"time"
)

type BaseAuthRequest struct {
	Username string
	Password string
}

type BaseAuthResponse struct {
	UserId string
	Tokens TokenPair
}

type RefreshResponse struct {
	Tokens TokenPair
}

type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  *time.Time
	RefreshExpiresAt *time.Time
}
