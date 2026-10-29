package threads

import "strings"

const maxSlugBase = 60

// translit русская и украинская кириллица в латиницу. Пустая строка: знак просто опускается.
var translit = map[rune]string{
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "e", 'ж': "zh", 'з': "z", 'и': "i",
	'й': "y", 'к': "k", 'л': "l", 'м': "m", 'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t",
	'у': "u", 'ф': "f", 'х': "kh", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "shch", 'ъ': "", 'ы': "y", 'ь': "",
	'э': "e", 'ю': "yu", 'я': "ya", 'і': "i", 'ї': "yi", 'є': "ye", 'ґ': "g",
}

// MakeSlug строит читаемый slug из заголовка и добавляет хвост id, чтобы он был уникален внутри доски.
// Заголовок без латиницы и цифр после транслитерации даёт slug из одного хвоста.
func MakeSlug(title, id string) string {
	base, suffix := slugify(title), idSuffix(id)
	if base == "" {
		return suffix
	}

	return base + "-" + suffix
}

func slugify(s string) string {
	var b strings.Builder
	lastDash := true // не начинаем с дефиса

	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if t, ok := translit[r]; ok {
				if t != "" {
					b.WriteString(t)
					lastDash = false
				}

				continue
			}
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}

	out := strings.TrimRight(b.String(), "-")
	if len(out) > maxSlugBase {
		out = strings.TrimRight(out[:maxSlugBase], "-")
	}

	return out
}

// idSuffix последние 8 символов uuid без дефисов: у UUIDv7 это случайная часть.
func idSuffix(id string) string {
	id = strings.ToLower(strings.ReplaceAll(id, "-", ""))
	if len(id) > 8 {
		id = id[len(id)-8:]
	}

	return id
}
