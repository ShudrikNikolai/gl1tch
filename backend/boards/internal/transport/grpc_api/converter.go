package grpc_api

import (
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/domain"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/boards"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/posts"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads"
	boardsv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/boards/v1"
	commonv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/common/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func timestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}

	return timestamppb.New(t)
}

func timestampPtr(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}

	return timestamppb.New(*t)
}

func authorToProto(a domain.Author) *boardsv1.Author {
	return &boardsv1.Author{UserId: a.UserID, Username: a.Username}
}

func authorPtrToProto(a *domain.Author) *boardsv1.Author {
	if a == nil {
		return nil
	}

	return authorToProto(*a)
}

func boardToProto(b boards.Board) *boardsv1.Board {
	return &boardsv1.Board{
		BoardId:        b.ID,
		Slug:           b.Slug,
		Title:          b.Title,
		Description:    b.Description,
		Visibility:     visibilityToProto(b.Visibility),
		ThreadCount:    int32(b.ThreadCount),
		PostCount:      int32(b.PostCount),
		LastPostAuthor: authorPtrToProto(b.LastPostAuthor),
		LastPostAt:     timestampPtr(b.LastPostAt),
		SortOrder:      int32(b.SortOrder),
	}
}

func threadToProto(t threads.Thread) *boardsv1.Thread {
	return &boardsv1.Thread{
		ThreadId:       t.ID,
		BoardId:        t.BoardID,
		Title:          t.Title,
		Slug:           t.Slug,
		Author:         authorToProto(t.Author),
		Status:         statusToProto(t.Status),
		Category:       categoryToProto(t.Category),
		Visibility:     visibilityToProto(t.Visibility),
		PostCount:      int32(t.PostCount),
		Pinned:         t.Pinned,
		LastPostAuthor: authorToProto(t.LastPostAuthor),
		LastPostAt:     timestamp(t.LastPostAt),
		CreatedAt:      timestamp(t.CreatedAt),
		UpdatedAt:      timestamp(t.UpdatedAt),
	}
}

// postToProto: у удалённого поста текст пустой (PublicBody), а в ответе стоит deleted=true.
func postToProto(p posts.Post) *boardsv1.Post {
	return &boardsv1.Post{
		PostId:       p.ID,
		ThreadId:     p.ThreadID,
		ParentPostId: p.ParentID,
		Author:       authorToProto(p.Author),
		Body:         p.PublicBody(),
		Number:       int32(p.Number),
		Deleted:      p.Deleted(),
		EditedAt:     timestampPtr(p.EditedAt),
		CreatedAt:    timestamp(p.CreatedAt),
		UpdatedAt:    timestamp(p.UpdatedAt),
	}
}

func pageRequest(p *commonv1.CursorPaginationRequest) domain.PageRequest {
	return domain.PageRequest{Size: int(p.GetPageSize()), Token: p.GetPageToken()}
}

func pageResponse(i domain.PageInfo) *commonv1.CursorPaginationResponse {
	return &commonv1.CursorPaginationResponse{NextPageToken: i.NextToken, HasMore: i.HasMore}
}
