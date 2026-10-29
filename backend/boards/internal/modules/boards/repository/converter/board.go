package converter

import (
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/boards"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/boards/repository/model"
)

func BoardSrvToModel(b boards.Board) model.Board {
	m := model.Board{
		ID:          b.ID,
		Slug:        b.Slug,
		Title:       b.Title,
		Description: b.Description,
		Visibility:  string(b.Visibility),
		SortOrder:   b.SortOrder,
		ThreadCount: b.ThreadCount,
		PostCount:   b.PostCount,
		LastPostAt:  b.LastPostAt,
		CreatedBy:   b.CreatedBy,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
	if b.LastPostAuthor != nil {
		m.LastPostAuthorID = &b.LastPostAuthor.UserID
		m.LastPostAuthorName = &b.LastPostAuthor.Username
	}

	return m
}

func BoardModelToSrv(m model.Board) boards.Board {
	b := boards.Board{
		ID:          m.ID,
		Slug:        m.Slug,
		Title:       m.Title,
		Description: m.Description,
		Visibility:  domain.Visibility(m.Visibility),
		SortOrder:   m.SortOrder,
		ThreadCount: m.ThreadCount,
		PostCount:   m.PostCount,
		LastPostAt:  m.LastPostAt,
		CreatedBy:   m.CreatedBy,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
	if m.LastPostAuthorID != nil && m.LastPostAuthorName != nil {
		b.LastPostAuthor = &domain.Author{UserID: *m.LastPostAuthorID, Username: *m.LastPostAuthorName}
	}

	return b
}
