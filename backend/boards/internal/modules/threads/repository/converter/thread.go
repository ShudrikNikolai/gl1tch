package converter

import (
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads/repository/model"
)

func ThreadSrvToModel(t threads.Thread) model.Thread {
	return model.Thread{
		ID:                 t.ID,
		BoardID:            t.BoardID,
		Title:              t.Title,
		Slug:               t.Slug,
		AuthorID:           t.Author.UserID,
		AuthorName:         t.Author.Username,
		Status:             string(t.Status),
		Category:           string(t.Category),
		Visibility:         string(t.Visibility),
		Pinned:             t.Pinned,
		PostCount:          t.PostCount,
		LastPostAuthorID:   t.LastPostAuthor.UserID,
		LastPostAuthorName: t.LastPostAuthor.Username,
		LastPostAt:         t.LastPostAt,
		DeletedAt:          t.DeletedAt,
		CreatedAt:          t.CreatedAt,
		UpdatedAt:          t.UpdatedAt,
	}
}

func ThreadModelToSrv(m model.Thread) threads.Thread {
	return threads.Thread{
		ID:             m.ID,
		BoardID:        m.BoardID,
		Title:          m.Title,
		Slug:           m.Slug,
		Author:         domain.Author{UserID: m.AuthorID, Username: m.AuthorName},
		Status:         threads.Status(m.Status),
		Category:       threads.Category(m.Category),
		Visibility:     domain.Visibility(m.Visibility),
		Pinned:         m.Pinned,
		PostCount:      m.PostCount,
		LastPostAuthor: domain.Author{UserID: m.LastPostAuthorID, Username: m.LastPostAuthorName},
		LastPostAt:     m.LastPostAt,
		DeletedAt:      m.DeletedAt,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}
