package threadpg

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads"
)

// anonymousID подставляется вместо id читателя без аккаунта: под него не подходит ни одна тема.
const anonymousID = "00000000-0000-0000-0000-000000000000"

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// buildList собирает запрос списка тем и его аргументы. Порядок ORDER BY повторяет индексы из миграции,
// а условие курсора построено как сравнение кортежей, чтобы Postgres брал границу прямо из индекса.
func buildList(f threads.ListFilter, limit int, cur *cursorData) (string, []any) {
	var args []any
	arg := func(v any) string {
		args = append(args, v)

		return "$" + strconv.Itoa(len(args))
	}

	where := []string{"board_id = " + arg(f.BoardID) + "::uuid", "deleted_at IS NULL"}

	if f.Category != "" {
		where = append(where, "category = "+arg(string(f.Category)))
	}
	if f.Status != "" {
		where = append(where, "status = "+arg(string(f.Status)))
	}
	if f.AuthorID != "" {
		where = append(where, "author_id = "+arg(f.AuthorID)+"::uuid")
	}
	if f.Search != "" {
		where = append(where, "title ILIKE "+arg("%"+likeEscaper.Replace(f.Search)+"%")+` ESCAPE '\'`)
	}
	if !f.Viewer.Moderator {
		viewer := f.Viewer.UserID
		if viewer == "" {
			viewer = anonymousID
		}
		where = append(where, "(visibility = 'public' OR (visibility = 'author' AND author_id = "+arg(viewer)+"::uuid))")
	}

	var order string
	switch f.Order {
	case threads.OrderCreated:
		order = "pinned DESC, thread_id DESC"
		if cur != nil {
			where = append(where, fmt.Sprintf("(pinned, thread_id) < (%s, %s::uuid)", arg(cur.Pinned), arg(cur.ID)))
		}
	default:
		order = "pinned DESC, last_post_at DESC, thread_id DESC"
		if cur != nil {
			where = append(where, fmt.Sprintf("(pinned, last_post_at, thread_id) < (%s, %s::timestamptz, %s::uuid)",
				arg(cur.Pinned), arg(time.UnixMicro(cur.LastPostMicro).UTC()), arg(cur.ID)))
		}
	}

	q := "SELECT " + columns + " FROM threads WHERE " + strings.Join(where, " AND ") +
		" ORDER BY " + order + " LIMIT " + arg(limit)

	return q, args
}
