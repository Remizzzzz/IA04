package utils

import (
	"math/rand"
	"time"
)

// rng is a package-level random source seeded once at startup.
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// RandInt returns a random integer in [start, end), or start if the range is empty or invalid.
func RandInt(start int, end int) int {
	if end <= start {
		return start
	}
	return rng.Intn(end-start) + start
}
