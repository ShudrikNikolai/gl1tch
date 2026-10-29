package boardpg

import (
	"log/slog"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/infrastructure/storage/pg"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/boards/repository"
)

var _ repository.Repository = (*Repo)(nil)

type Repo struct {
	db  *pg.Client
	log *slog.Logger
}

func New(db *pg.Client, log *slog.Logger) *Repo {
	return &Repo{db: db, log: log}
}
