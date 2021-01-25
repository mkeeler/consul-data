package generators

import (
	"math/rand"

	uuid "github.com/hashicorp/go-uuid"
)

type randReader struct{}

func (r *randReader) Read(p []byte) (n int, err error) {
	return rand.Read(p)
}

var (
	uuidRand = &randReader{}
)

func UUIDGen() (string, error) {
	return uuid.GenerateUUIDWithReader(uuidRand)
}
