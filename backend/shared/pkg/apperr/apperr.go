package apperr

import (
	"errors"
	"fmt"
	"net/http"
)

type Kind int

const (
	KindInternal Kind = iota //всё необработанное
	KindInvalidArgument
	KindUnauthenticated
	KindPermissionDenied
	KindNotFound
	KindAlreadyExists
	KindRateLimited
	KindUnavailable
)

const (
	CodeBadRequest      = "BAD_REQUEST"
	CodeUnauthenticated = "UNAUTHENTICATED"
	CodeAccessDenied    = "ACCESS_DENIED"
	CodeNotFound        = "NOT_FOUND"
	CodeConflict        = "CONFLICT"
	CodeRateLimited     = "RATE_LIMITED"
	CodeUnavailable     = "UNAVAILABLE"
	CodeInternal        = "INTERNAL"
)

// Error прикладная ошибка. Message предназначен клиенту, Err (причина) наружу не отдаётся.
type Error struct {
	Kind    Kind
	Code    string
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}

	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.Err }

// New создаёт ошибку. Пустой code заменяется кодом по умолчанию для вида.
func New(kind Kind, code, message string) *Error {
	if code == "" {
		code = DefaultCode(kind)
	}

	return &Error{Kind: kind, Code: code, Message: message}
}

// Wrap то же, что New.
func Wrap(kind Kind, code, message string, err error) *Error {
	e := New(kind, code, message)
	e.Err = err

	return e
}

func InvalidArgument(code, message string) *Error  { return New(KindInvalidArgument, code, message) }
func Unauthenticated(code, message string) *Error  { return New(KindUnauthenticated, code, message) }
func PermissionDenied(code, message string) *Error { return New(KindPermissionDenied, code, message) }
func NotFound(code, message string) *Error         { return New(KindNotFound, code, message) }
func AlreadyExists(code, message string) *Error    { return New(KindAlreadyExists, code, message) }

// From возвращает *Error из цепочки err
func From(err error) *Error {
	if err == nil {
		return nil
	}
	var e *Error
	if errors.As(err, &e) {
		return e
	}

	return Wrap(KindInternal, CodeInternal, "internal error", err)
}

func KindOf(err error) Kind { return From(err).Kind }

func DefaultCode(k Kind) string {
	switch k {
	case KindInvalidArgument:
		return CodeBadRequest
	case KindUnauthenticated:
		return CodeUnauthenticated
	case KindPermissionDenied:
		return CodeAccessDenied
	case KindNotFound:
		return CodeNotFound
	case KindAlreadyExists:
		return CodeConflict
	case KindRateLimited:
		return CodeRateLimited
	case KindUnavailable:
		return CodeUnavailable
	default:
		return CodeInternal
	}
}

func HTTPStatus(k Kind) int {
	switch k {
	case KindInvalidArgument:
		return http.StatusBadRequest
	case KindUnauthenticated:
		return http.StatusUnauthorized
	case KindPermissionDenied:
		return http.StatusForbidden
	case KindNotFound:
		return http.StatusNotFound
	case KindAlreadyExists:
		return http.StatusConflict
	case KindRateLimited:
		return http.StatusTooManyRequests
	case KindUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
