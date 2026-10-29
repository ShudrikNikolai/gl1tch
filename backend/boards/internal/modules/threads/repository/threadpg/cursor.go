package threadpg

import (
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads"
)

const cursorKind = "threads"

// cursorData позиция последней отданной темы. Порядок входит в токен, чтобы его нельзя было применить к другой сортировке.
type cursorData struct {
	Order         threads.Order `json:"o"`
	Pinned        bool          `json:"p"`
	LastPostMicro int64         `json:"t,omitempty"` // last_post_at в микросекундах: точность timestamptz
	ID            string        `json:"i"`
}

func cursorFrom(order threads.Order, t threads.Thread) cursorData {
	return cursorData{Order: order, Pinned: t.Pinned, LastPostMicro: t.LastPostAt.UnixMicro(), ID: t.ID}
}
