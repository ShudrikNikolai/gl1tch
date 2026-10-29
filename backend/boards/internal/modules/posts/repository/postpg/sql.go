package postpg

const columns = `post_id::text, thread_id::text, number, parent_post_id::text, author_id::text, author_name,
	body, edited_at, deleted_at, deleted_by::text, created_at, updated_at`

const (
	qCreate = `INSERT INTO posts (post_id, thread_id, number, parent_post_id, author_id, author_name, body, created_at, updated_at)
		VALUES ($1::uuid, $2::uuid, $3, $4::uuid, $5::uuid, $6, $7, $8, $8)
		RETURNING ` + columns

	qGet = `SELECT ` + columns + ` FROM posts WHERE post_id = $1::uuid`

	// По индексу UNIQUE (thread_id, number)
	qList = `SELECT ` + columns + ` FROM posts
		WHERE thread_id = $1::uuid AND number > $2
		ORDER BY number
		LIMIT $3`

	qUpdateBody = `UPDATE posts SET body = $2, edited_at = $3, updated_at = $3
		WHERE post_id = $1::uuid AND deleted_at IS NULL
		RETURNING ` + columns

	qSoftDelete = `UPDATE posts SET deleted_at = $2, deleted_by = $3::uuid, updated_at = $2
		WHERE post_id = $1::uuid AND deleted_at IS NULL
		RETURNING ` + columns
)
