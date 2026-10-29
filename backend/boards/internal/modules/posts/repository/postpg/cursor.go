package postpg

const cursorKind = "posts"

// cursorData номер последнего отданного поста: по нему и идёт выдача (номера в теме уникальны и растут).
type cursorData struct {
	Number int `json:"n"`
}
