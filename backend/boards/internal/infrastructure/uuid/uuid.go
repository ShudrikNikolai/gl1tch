package uuid

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

func NewUUIDv7(now time.Time) string {
	var b [16]byte
	ms := uint64(now.UnixMilli())
	b[0], b[1], b[2], b[3], b[4], b[5] = byte(ms>>40), byte(ms>>32), byte(ms>>24), byte(ms>>16), byte(ms>>8), byte(ms)
	_, _ = rand.Read(b[6:])
	b[6] = b[6]&0x0f | 0x70 // версия 7
	b[8] = b[8]&0x3f | 0x80 // вариант RFC 4122

	h := hex.EncodeToString(b[:])

	return h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:]
}
