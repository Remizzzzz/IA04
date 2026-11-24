package utils

import (
	"math/rand"
	"slices"
	"time"
)

// rng is a package-level random source seeded once at startup.
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// Tuple allow to create a pair of values of any given types
type Tuple[T1 any, T2 any] struct { //Because it's always useful
	First  T1
	Second T2
}

// RandInt returns a random integer in [start, end), or start if the range is empty or invalid.
func RandInt(start int, end int) int {
	if end <= start {
		return start
	}
	return rng.Intn(end-start) + start
}

// Made by AI : Implements contains
func Contains[T comparable](tab []T, element T) bool {
	for _, value := range tab {
		if value == element {
			return true
		}
	}
	return false
}

// Util I made to delete an element in a slice dynamically
// Cannot be use in a loop for _,i :=range slice, as the
// slice will not be updated dynamically in the loop
func DeleteFromSlice[T any](slice *[]*T, element *T) {
	idx := slices.Index(*slice, element)
	if idx == -1 {
		return
	}
	*slice = slices.Delete(*slice, idx, idx+1)
}
