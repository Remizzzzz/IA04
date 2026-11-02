package agents

import (
	"soldier/cmd/geometry"
	"soldier/cmd/utils"
)

// Strategy identifies which movement policy a unit follows.
type Strategy int

const (
	Random Strategy = iota // pick a random reachable tile each turn
)

// Order is what World sends a unit at the start of each turn: what tiles it
// can currently see, which tiles it could move to, and what it's asked to do.
// Bundling everything a unit needs for the turn into one message (instead of
// separate channels) avoids having to coordinate the order multiple sends
// and receives happen in.
type Order struct {
	Vision    []geometry.Pos
	Reachable []geometry.Pos
	Cmd       string
}

// Unit represents a single agent. It holds no reference to the world: World
// owns every bit of spatial logic (movement, vision) and drives units purely
// by sending Orders and reading back Moves, which keeps world free to hold
// *Unit references without agents importing world back (an import cycle).
type Unit struct {
	Speed       int
	Position    geometry.Pos
	Strategy    Strategy
	VisionRange int
	Vision      []geometry.Pos    // tiles seen as of the last Order received
	Orders      chan Order        // World -> Unit, sent at the start of each turn
	Moves       chan geometry.Pos // Unit -> World, the destination chosen for this turn
}

// NewUnit creates a unit with default Speed/VisionRange for the given
// strategy. It still needs to be placed in a World (see World.SpawnUnit)
// before it can take part in a simulation.
func NewUnit(strategy Strategy) *Unit {
	return &Unit{
		Speed:       1,
		Strategy:    strategy,
		VisionRange: 3,
		Orders:      make(chan Order),
		Moves:       make(chan geometry.Pos),
	}
}

// region strategy

// UseStrategy picks a destination among the reachable tiles carried by the
// last Order, according to the unit's Strategy.
func (u *Unit) UseStrategy(reachable []geometry.Pos) geometry.Pos {
	switch u.Strategy {
	case Random:
		return u.strategyRandom(reachable)
	}
	return u.Position
}

// strategyRandom picks a random tile among the reachable ones, excluding the
// unit's own position.
func (u *Unit) strategyRandom(reachable []geometry.Pos) geometry.Pos {
	others := make([]geometry.Pos, 0, len(reachable))
	for _, p := range reachable {
		if p != u.Position {
			others = append(others, p)
		}
	}
	if len(others) == 0 {
		return u.Position // no free tile to move to, stay put
	}
	return others[utils.RandInt(0, len(others))]
}

// region simulation

// Simulate starts the unit's event loop in its own goroutine: for every Order
// received it records its current vision and, on a "MOVE" command, replies
// on Moves with the tile it wants to move to. World reads Moves and performs
// the actual move, since it alone owns the grid.
func (u *Unit) Simulate() {
	go func() {
		for order := range u.Orders {
			u.Vision = order.Vision
			switch order.Cmd {
			case "MOVE":
				u.Moves <- u.UseStrategy(order.Reachable)
			}
		}
	}()
}
