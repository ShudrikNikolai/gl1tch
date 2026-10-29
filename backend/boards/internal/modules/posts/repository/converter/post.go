package converter

import (
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/posts"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/posts/repository/model"
)

func PostSrvToModel(p posts.Post) model.Post {
	m := model.Post{
		ID:         p.ID,
		ThreadID:   p.ThreadID,
		Number:     p.Number,
		AuthorID:   p.Author.UserID,
		AuthorName: p.Author.Username,
		Body:       p.Body,
		EditedAt:   p.EditedAt,
		DeletedAt:  p.DeletedAt,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}
	if p.ParentID != "" {
		m.ParentID = &p.ParentID
	}
	if p.DeletedBy != "" {
		m.DeletedBy = &p.DeletedBy
	}

	return m
}

func PostModelToSrv(m model.Post) posts.Post {
	p := posts.Post{
		ID:        m.ID,
		ThreadID:  m.ThreadID,
		Number:    m.Number,
		Author:    domain.Author{UserID: m.AuthorID, Username: m.AuthorName},
		Body:      m.Body,
		EditedAt:  m.EditedAt,
		DeletedAt: m.DeletedAt,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
	if m.ParentID != nil {
		p.ParentID = *m.ParentID
	}
	if m.DeletedBy != nil {
		p.DeletedBy = *m.DeletedBy
	}

	return p
}
