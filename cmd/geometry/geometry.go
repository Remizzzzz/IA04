// Package geometry holds the coordinate type shared by world and agents.
// It exists so agents can describe a position without importing world,
// since world needs to import agents (to hold *agents.Unit references)
// and that would otherwise create an import cycle.
package geometry

import "math"

// Pos is a coordinate on the grid.
type Pos struct {
	X int
	Y int
}

// Made by Claude:
// ManhattanDistance returns the grid distance between a and b (the number of
// cardinal steps needed to go from one to the other).
func ManhattanDistance(a, b Pos) int {
	dx := a.X - b.X
	if dx < 0 {
		dx = -dx
	}
	dy := a.Y - b.Y
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

func EuclidianDistance(a, b Pos) int {
	dx := float64(b.X - a.X)
	dy := float64(b.Y - a.Y)
	return int(math.Sqrt(dx*dx + dy*dy))
}
