package info

import (
	"math/rand"
	"time"
)

const (
	UUIDLength = 16
	charset    = "abcdefghijklmnopqrstuvwxyz0123456789"
)

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec // this is just for random string, no security issue here

func GenerateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
