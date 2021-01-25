package generators

import (
	"encoding/base64"
	"fmt"
	"math/rand"
)

func RandomB64Generator(minSize int, maxSize int) StringGenerator {
	return func() (string, error) {
		size := minSize + rand.Intn(maxSize-minSize)
		raw := make([]byte, size)
		_, err := rand.Read(raw)

		// Technically math/rand.Read is guaranteed to always return a nil error but
		// we are checking anyways just in case we switch over to something else like
		// crypto/rand where the same guarantees might not be in place.
		if err != nil {
			return "", fmt.Errorf("Failed to generate random KV value: %w", err)
		}

		encoded := make([]byte, base64.StdEncoding.EncodedLen(len(raw)))
		base64.StdEncoding.Encode(encoded, raw)

		return string(encoded), nil
	}
}
