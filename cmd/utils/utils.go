package utils

import (
	"math/rand"
	"time"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandInt(start int, end int) int {
	if end <= start {
		return start
	}
	return rng.Intn(end-start) + start
}
