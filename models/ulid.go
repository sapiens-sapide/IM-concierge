package models

import (
	"crypto/rand"
	"github.com/oklog/ulid"
	"time"
)

func NewULID() ulid.ULID {
	return ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)
}
