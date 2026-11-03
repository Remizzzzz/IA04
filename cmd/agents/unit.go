package agents

import (
	"soldier/cmd/geometry"
	"soldier/cmd/utils"
)

// nArmies hands out a unique allegiance to every default (solo) unit, so it
// never coincidentally matches another unit's and never forms an Army.
var nArmies int

// Strategy identifies which decision policy a unit follows.
type Strategy int

const (
	Random      Strategy = iota // move to a random reachable tile each turn
	Coordinated                 // Every units from the same allegiance works together and are coordinated
	Squad                       // Every units from the same allegiance works by group of X units. They coordinate multiple squads (to implement later maybe)
)

// ActionKind identifies what a unit chose to do this turn.
type ActionKind int

const (
	Stay  ActionKind = iota // hold the current position (also the zero value)
	Move                    // move to Target
	Shoot                   // shoot whatever is on Target
)

// Action is a unit's reply to an Order: what it wants to do, and the tile
// that action applies to (a destination for Move, a target's tile for
// Shoot; unused for Stay).
type Action struct {
	Kind   ActionKind
	Target geometry.Pos
}

// Order is what World sends a unit at the start of each turn: what tiles it
// can currently see, which tiles it could move to, and which other units it
// could shoot at. Bundling everything a unit needs for the turn into one
// message (instead of separate channels) avoids having to coordinate the
// order multiple sends and receives happen in.
type Order struct {
	Vision    []geometry.Pos
	Reachable []geometry.Pos //Reachable tiles where the units can move (even if out of sight for now)
	Targets   []*Unit        // other units currently visible, in shooting range - read-only: only World may mutate a Unit's fields
}

// Unit represents a single agent. It holds no reference to the world: World
// owns every bit of spatial logic (movement, vision, shooting) and drives
// units purely by sending Orders and reading back Actions, which keeps world
// free to hold *Unit references without agents importing world back (an
// import cycle).
type Unit struct {
	Speed       int
	Position    geometry.Pos
	Strategy    Strategy
	VisionRange int
	Vision      []geometry.Pos // tiles seen as of the last Order received
	Allegiance  int            // units with the same allegiance recognize each other as allies
	Army        *Army          // the army this unit belongs to, set by World on spawn
	Alive       bool
	Orders      chan Order  // World -> Unit, sent at the start of each turn
	Actions     chan Action // Unit -> World, the action chosen for this turn
}

// NewUnit creates a unit with the given strategy and allegiance. Several
// units can share an allegiance so they end up in the same Army together
// (see World.SpawnAllied). It still needs to be placed in a World before it
// can take part in a simulation.
func NewUnit(strategy Strategy, allegiance int) *Unit {
	return &Unit{
		Speed:       1,
		Strategy:    strategy,
		VisionRange: 3,
		Allegiance:  allegiance,
		Alive:       true,
		Orders:      make(chan Order),
		Actions:     make(chan Action),
	}
}

// NewUnitDefault creates a solo unit with a guaranteed-unique allegiance, so
// it never joins or forms an active Army. It still needs to be placed in a
// World (see World.SpawnUnit) before it can take part in a simulation.
func NewUnitDefault(strategy Strategy) *Unit {
	nArmies++
	return NewUnit(strategy, nArmies)
}

// region strategy

// UseStrategy picks this turn's Action out of the reachable tiles and
// targets carried by the last Order, according to the unit's Strategy.
func (u *Unit) UseStrategy(reachable []geometry.Pos, targets []*Unit) Action {
	switch u.Strategy {
	case Random:
		return u.strategyRandom(reachable)
	case Coordinated:
		return u.strategyCoordinated(reachable, targets)
	}
	return Action{Kind: Stay}
}

// strategyRandom picks a random tile among the reachable ones, excluding the
// unit's own position, and moves there; it stays put if there is nowhere to go.
func (u *Unit) strategyRandom(reachable []geometry.Pos) Action {
	others := make([]geometry.Pos, 0, len(reachable))
	for _, p := range reachable {
		if p != u.Position {
			others = append(others, p)
		}
	}
	if len(others) == 0 {
		return Action{Kind: Stay}
	}
	return Action{Kind: Move, Target: others[utils.RandInt(0, len(others))]}
}

// strategyCoordinated asks the unit's Army what to do when it actually has
// soldiers to coordinate and a living general to issue the plan; otherwise
// it falls back to deciding on its own.
func (u *Unit) strategyCoordinated(reachable []geometry.Pos, targets []*Unit) Action {
	if u.Army != nil && u.Army.IsActive() && u.Army.General.Alive {
		return u.Army.Decide(u, reachable, targets)
	}
	return u.strategyRandom(reachable)
}

// region simulation

// Simulate starts the unit's event loop in its own goroutine: for every Order
// received it records its current vision and replies on Actions with the
// Action it chose. World reads Actions and resolves them, since it alone
// owns the grid (and, eventually, combat state).
func (u *Unit) Simulate() {
	go func() {
		for order := range u.Orders {
			u.Vision = order.Vision
			u.Actions <- u.UseStrategy(order.Reachable, order.Targets)
		}
	}()
}
