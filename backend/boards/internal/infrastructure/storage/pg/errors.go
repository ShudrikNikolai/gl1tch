package pg

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	codeUniqueViolation = "23505"
	codeInvalidText     = "22P02" // например, строка не похожа на uuid
)

// UniqueViolation сообщает, что запись нарушила уникальное ограничение, и возвращает его имя.
func UniqueViolation(err error) (constraint string, ok bool) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == codeUniqueViolation {
		return pgErr.ConstraintName, true
	}

	return "", false
}

// InvalidInput сообщает, что Postgres не смог разобрать значение параметра (чаще всего кривой uuid от клиента).
// Для репозиториев это «такой записи нет», а не внутренняя ошибка.
func InvalidInput(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == codeInvalidText
}
