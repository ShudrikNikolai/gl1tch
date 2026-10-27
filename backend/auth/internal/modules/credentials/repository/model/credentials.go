package model

import "time"

type Credential struct {
	UserID       string
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}
