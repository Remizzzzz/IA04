// Package geometry holds the coordinate type shared by world and agents.
// It exists so agents can describe a position without importing world,
// since world needs to import agents (to hold *agents.Unit references)
// and that would otherwise create an import cycle.
package geometry

// Pos is a coordinate on the grid.
type Pos struct {
	X int
	Y int
}
