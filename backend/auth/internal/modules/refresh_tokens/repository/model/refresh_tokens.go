package model

// Session значение ключа rt:<hash> в Redis (JSON).
type Session struct {
	UserID     string `json:"user_id"`
	IssuedAtMs int64  `json:"issued_at_ms"`
}
