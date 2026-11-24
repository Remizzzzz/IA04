package world

import (
	"fmt"
	"soldier/cmd/agents"
	"soldier/cmd/geometry"
)

// SpawnUnit creates a solo unit with the given strategy, places it on a
// random free tile, registers it with the world and starts its goroutine.
func (w *World) SpawnUnitRandom(strategy agents.Strategy) *agents.Unit {
	return w.spawn(agents.NewUnitDefault(strategy), w.GetRandomFreePos())
}

// SpawnUnitAt is like SpawnUnit but places the unit at a specific tile,
// panicking if that tile isn't free.
func (w *World) SpawnUnitAt(strategy agents.Strategy, p geometry.Pos) *agents.Unit {
	if !w.IsFree(p) {
		panic(fmt.Sprintf("cannot spawn unit at position (%d, %d): tile is not free", p.X, p.Y))
	}
	return w.spawn(agents.NewUnitDefault(strategy), p)
}

// SpawnAllied creates a unit with the given strategy and allegiance, places
// it on a random free tile, and enlists it into the Army for that
// allegiance (forming one if this is the first unit to carry it).
func (w *World) SpawnAlliedRandom(strategy agents.Strategy, allegiance int, unitType agents.UnitType) *agents.Unit {
	return w.spawn(agents.NewUnit(strategy, allegiance, unitType), w.GetRandomFreePos())
}

// SpawnAlliedAt is like SpawnAllied but places the unit at a specific tile,
// panicking if that tile isn't free.
func (w *World) SpawnAlliedAt(strategy agents.Strategy, allegiance int, p geometry.Pos, unitType agents.UnitType) *agents.Unit {
	if !w.IsFree(p) {
		panic(fmt.Sprintf("cannot spawn unit at position (%d, %d): tile is not free", p.X, p.Y))
	}
	return w.spawn(agents.NewUnit(strategy, allegiance, unitType), p)
}

func (w *World) spawn(u *agents.Unit, p geometry.Pos) *agents.Unit {
	u.Position = p
	w.SetTile(p, occupied)
	w.Units = append(w.Units, u)
	w.enlist(u)
	u.Simulate()
	return u
}
