package converter

import (
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/refresh_tokens"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/refresh_tokens/repository/model"
)

func SessionSrvToModel(s refresh_tokens.Session) model.Session {
	return model.Session{UserID: s.UserID, IssuedAtMs: s.IssuedAt.UnixMilli()}
}

func SessionModelToSrv(m model.Session) refresh_tokens.Session {
	return refresh_tokens.Session{UserID: m.UserID, IssuedAt: time.UnixMilli(m.IssuedAtMs)}
}
