package converter

import (
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/credentials"
	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/modules/credentials/repository/model"
)

func CredentialSrvToModel(c credentials.Credential) model.Credential {
	return model.Credential{
		UserID:       c.UserID,
		Username:     c.Username,
		PasswordHash: c.PasswordHash,
		CreatedAt:    c.CreatedAt,
	}
}

func CredentialModelToSrv(m model.Credential) credentials.Credential {
	return credentials.Credential{
		UserID:       m.UserID,
		Username:     m.Username,
		PasswordHash: m.PasswordHash,
		CreatedAt:    m.CreatedAt,
	}
}
