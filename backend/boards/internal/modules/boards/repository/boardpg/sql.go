package boardpg

// slugConstraint имя уникального индекса из миграции.
const slugConstraint = "boards_slug_lower_idx"

const columns = `board_id::text, slug, title, description, visibility, sort_order,
	thread_count, post_count, last_post_author_id::text, last_post_author_name, last_post_at,
	created_by::text, created_at, updated_at`

const (
	qCreate = `INSERT INTO boards (board_id, slug, title, description, visibility, sort_order, created_by)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7::uuid)
		RETURNING ` + columns

	qGet = `SELECT ` + columns + ` FROM boards WHERE board_id = $1::uuid`

	qList = `SELECT ` + columns + ` FROM boards
		WHERE ($1::boolean OR visibility = 'public')
		ORDER BY sort_order, board_id`

	// COALESCE: nil-параметр оставляет прежнее значение превью
	qApply = `UPDATE boards SET
			thread_count = thread_count + $2,
			post_count = post_count + $3,
			last_post_author_id = COALESCE($4::uuid, last_post_author_id),
			last_post_author_name = COALESCE($5::text, last_post_author_name),
			last_post_at = COALESCE($6::timestamptz, last_post_at),
			updated_at = now()
		WHERE board_id = $1::uuid`
)
