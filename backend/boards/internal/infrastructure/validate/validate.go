// Package validation проверяет и нормализует пользовательский ввод. Ошибки это apperr.InvalidArgument.
package validation

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/posts"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
)

const (
	maxSlugLen          = 64
	maxTitleRunes       = 200
	maxDescriptionRunes = 1000
	maxSearchRunes      = 100
	maxReasonRunes      = 500
)

// slugRe совпадает с CHECK в миграции.
var slugRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func badRequest(code, msg string) error {
	return apperr.InvalidArgument(code, msg)
}

// Slug slug доски: строчная латиница, цифры и дефисы.
func Slug(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > maxSlugLen || !slugRe.MatchString(s) {
		return "", badRequest("INVALID_SLUG", "slug must contain lowercase latin letters, digits and single hyphens, up to 64 characters")
	}

	return s, nil
}

// Title заголовок доски или темы: одна строка, 1-200 символов.
func Title(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" || utf8.RuneCountInString(s) > maxTitleRunes || strings.IndexFunc(s, unicode.IsControl) >= 0 {
		return "", badRequest("INVALID_TITLE", "title must be a single line of 1 to 200 characters")
	}

	return s, nil
}

func Description(s string) (string, error) {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) > maxDescriptionRunes {
		return "", badRequest("INVALID_DESCRIPTION", "description is too long (max 1000 characters)")
	}

	return s, nil
}

// Body текст поста: markdown, перевод строк разрешён. Пробелы по краям обрезаются.
func Body(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" || utf8.RuneCountInString(s) > posts.MaxBodyLength {
		return "", badRequest("INVALID_BODY", "post body must contain 1 to 20000 characters")
	}

	return s, nil
}

// Search подстрока для поиска по заголовкам. Пустая строка означает «без поиска».
func Search(s string) (string, error) {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) > maxSearchRunes {
		return "", badRequest("INVALID_SEARCH", "search query is too long (max 100 characters)")
	}

	return s, nil
}

// Reason причина модераторского действия, необязательна.
func Reason(s string) (string, error) {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) > maxReasonRunes {
		return "", badRequest("INVALID_REASON", "reason is too long (max 500 characters)")
	}

	return s, nil
}
