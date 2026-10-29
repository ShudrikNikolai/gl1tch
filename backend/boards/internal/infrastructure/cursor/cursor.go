// Package cursor кодирует состояние курсорной пагинации в непрозрачный токен страницы.
package cursor

import (
	"encoding/base64"
	"encoding/json"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
)

var ErrInvalid = apperr.InvalidArgument("INVALID_PAGE_TOKEN", "invalid page token")

type envelope struct {
	Kind string          `json:"k"`
	Data json.RawMessage `json:"d"`
}

// Encode упаковывает v в токен. kind защищает от подстановки токена из другого списка.
func Encode(kind string, v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	raw, err := json.Marshal(envelope{Kind: kind, Data: data})
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// Decode разбирает токен. Любая проблема с токеном клиента это ErrInvalid, а не внутренняя ошибка.
func Decode(token, kind string, v any) error {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return ErrInvalid
	}

	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil || env.Kind != kind {
		return ErrInvalid
	}
	if err := json.Unmarshal(env.Data, v); err != nil {
		return ErrInvalid
	}

	return nil
}
