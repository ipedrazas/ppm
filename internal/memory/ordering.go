package memory

import (
	"time"

	"github.com/google/uuid"
)

// Today returns the current UTC date as YYYY-MM-DD.
func Today() string {
	return time.Now().UTC().Format("2006-01-02")
}

// OrderingKey returns a monotonic, lexically sortable key for the frontmatter
// "ts" field. UUIDv7 is time-ordered, so its canonical string sorts
// chronologically — and unlike the TS in-process counter it stays monotonic
// across separate CLI invocations. Falls back to a nanosecond timestamp if the
// UUID source fails.
func OrderingKey() string {
	if id, err := uuid.NewV7(); err == nil {
		return id.String()
	}
	return time.Now().UTC().Format(time.RFC3339Nano)
}
