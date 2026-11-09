package validation

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
)

const (
	minPasswordRunes = 8
	maxPasswordBytes = 128
	minNickRunes     = 3
	maxNickRunes     = 32
)

func badRequest(msg string) error { return apperr.InvalidArgument("", msg) }

func validatePassword(p string) error {
	if utf8.RuneCountInString(p) < minPasswordRunes {
		return badRequest("password must be at least 8 characters")
	}
	if len(p) > maxPasswordBytes {
		return badRequest("password is too long")
	}

	return nil
}

// validateNickname: буквы, цифры, '_', '.', '-' длиной 3-32 символа; начинаться должен с буквы или цифры.
func validateNickname(n string) (string, error) {
	n = strings.TrimSpace(n)
	count := utf8.RuneCountInString(n)
	if count < minNickRunes || count > maxNickRunes {
		return "", badRequest("nickname must be 3-32 characters")
	}
	for i, r := range n {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
		case i > 0 && (r == '_' || r == '.' || r == '-'):
		default:
			return "", badRequest("nickname may contain letters, digits, '_', '.', '-' and must start with a letter or digit")
		}
	}

	return n, nil
}

// Registration проверяет логин и пароль при регистрации и возвращает нормализованный логин.
func Registration(username, password string) (string, error) {
	nick, err := validateNickname(username)
	if err != nil {
		return "", err
	}
	if err := validatePassword(password); err != nil {
		return "", err
	}

	return nick, nil
}

// Login не применяет правила регистрации (они могли измениться), а только отсекает пустой и чрезмерно длинный ввод,
// чтобы не отдавать Argon2 гигантские пароли.
func Login(username, password string) (string, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return "", badRequest("username and password are required")
	}
	if utf8.RuneCountInString(username) > maxNickRunes || len(password) > maxPasswordBytes {
		return "", badRequest("invalid username or password")
	}

	return username, nil
}
