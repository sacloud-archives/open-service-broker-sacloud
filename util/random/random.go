package random

import (
	"math/rand"
	"time"
)

var r *rand.Rand

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// String generates random strings with specified length
// Only letters in [a-zA-z] are used as strings
func String(strlen int) string {
	result := make([]byte, strlen)
	for i := range result {
		result[i] = chars[r.Intn(len(chars))]
	}
	return string(result)
}
