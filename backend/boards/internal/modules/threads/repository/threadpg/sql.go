package threadpg

// slugConstraint имя уникального ограничения из миграции.
const slugConstraint = "threads_board_slug_key"

const columns = `thread_id::text, board_id::text, title, slug, author_id::text, author_name,
	status, category, visibility, pinned, post_count,
	last_post_author_id::text, last_post_author_name, last_post_at,
	deleted_at, created_at, updated_at`

// Все изменяющие запросы работают только с неудалёнными темами.
const (
	qCreate = `INSERT INTO threads (thread_id, board_id, title, slug, author_id, author_name,
			status, category, visibility, pinned,
			last_post_author_id, last_post_author_name, last_post_at, created_at, updated_at)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5::uuid, $6, $7, $8, $9, $10, $5::uuid, $6, $11, $11, $11)
		RETURNING ` + columns

	qGet = `SELECT ` + columns + ` FROM threads WHERE thread_id = $1::uuid AND deleted_at IS NULL`

	// COALESCE: nil-параметр оставляет прежнее значение
	qUpdate = `UPDATE threads SET
			title = COALESCE($2::text, title),
			category = COALESCE($3::text, category),
			visibility = COALESCE($4::text, visibility),
			updated_at = $5
		WHERE thread_id = $1::uuid AND deleted_at IS NULL
		RETURNING ` + columns

	qSetStatus = `UPDATE threads SET status = $2, updated_at = $3
		WHERE thread_id = $1::uuid AND deleted_at IS NULL
		RETURNING ` + columns

	qSetPinned = `UPDATE threads SET pinned = $2, updated_at = $3
		WHERE thread_id = $1::uuid AND deleted_at IS NULL
		RETURNING ` + columns

	qSoftDelete = `UPDATE threads SET deleted_at = $2, deleted_by = $3::uuid, updated_at = $2
		WHERE thread_id = $1::uuid AND deleted_at IS NULL`

	qNextPostNumber = `UPDATE threads SET
			post_seq = post_seq + 1,
			post_count = post_count + 1,
			last_post_author_id = $2::uuid,
			last_post_author_name = $3,
			last_post_at = $4,
			updated_at = $4
		WHERE thread_id = $1::uuid AND deleted_at IS NULL
		RETURNING post_seq`

	qReleasePost = `UPDATE threads SET post_count = post_count - 1, updated_at = $2
		WHERE thread_id = $1::uuid AND deleted_at IS NULL`
)
