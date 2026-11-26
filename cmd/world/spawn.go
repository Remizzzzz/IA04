package world

import (
	"fmt"
	"soldier/cmd/agents"
	"soldier/cmd/geometry"
	"soldier/cmd/utils"
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

// anchor return the starting position of the general of army 1,
// that will be replicated by symmetry for all other armies
func (w *World) anchor() geometry.Pos {
	targetPos := anchorTarget(w.MapCfg.Size, w.MapCfg.Symmetry)
	var selectedPos geometry.Pos = geometry.Pos{X: -1, Y: -1}
	var currentPos geometry.Pos
	found := false
	if w.IsFree(targetPos) && w.MapCfg.Symmetry.InDomain(targetPos, w.MapCfg.Size) {
		return targetPos
	}
	for y, row := range w.Grid {
		for x := range row {
			currentPos = geometry.Pos{X: x, Y: y}
			if !w.MapCfg.Symmetry.InDomain(currentPos, w.MapCfg.Size) || !w.IsFree(currentPos) {
				continue
			}
			found = true
			if (selectedPos == geometry.Pos{X: -1, Y: -1}) {
				selectedPos = currentPos
			}

			if geometry.ManhattanDistance(currentPos, targetPos) < geometry.ManhattanDistance(selectedPos, targetPos) {
				selectedPos = currentPos
			}
		}
	}
	if found {
		return selectedPos
	}
	panic("Map has no free tiles in the spawn domain")
}

// used by anchor to know where the army should start ideally, based on the
// symmetry chosen
func anchorTarget(worldSize int, sym geometry.Symmetry) geometry.Pos {
	switch sym {
	case geometry.VerticalSymmetry:
		return geometry.Pos{X: 0, Y: worldSize / 2}
	case geometry.HorizontalSymmetry:
		return geometry.Pos{X: worldSize / 2, Y: 0}
	case geometry.DoubleSymmetry:
		return geometry.Pos{X: 0, Y: 0}
	case geometry.NoSymmetry: //NoSymmetry
		return geometry.Pos{X: 0, Y: 0}
	default:
		panic("Symmetry type not recognized")
	}
}

// This function will give free tiles around the given general
// to spawning its army
func (w *World) Cluster(general geometry.Pos, count int) []geometry.Pos {
	if count <= 0 {
		return []geometry.Pos{}
	}
	results := make([]geometry.Pos, 0)
	results = append(results, general)
	if count == 1 {
		return results
	}
	//passable only if it's free and in domain/ or no symmetry constraints
	passable := func(p geometry.Pos) bool {
		return w.IsFree(p) && (w.MapCfg.Symmetry.InDomain(p, w.MapCfg.Size) || w.MapCfg.Symmetry == geometry.NoSymmetry)
	}
	visit := func(p geometry.Pos, _ geometry.Pos) bool {
		results = append(results, p)
		return len(results) < count
	}
	geometry.BFS(w.MapCfg.Size, []geometry.Pos{general}, passable, visit)
	if len(results) < count {
		panic(fmt.Sprintf("cannot spawn %d units around (%d, %d): only %d reachable free tiles in the spawn domain",
			count, general.X, general.Y, len(results)))
	}
	return results
}

func (w *World) SpawnArmies(nArmies int, composition []agents.UnitType, strategy agents.Strategy) {
	if nArmies < 2 {
		panic("Not enough armies")
	}
	var armiesIdx []int
	switch w.MapCfg.Symmetry {
	case geometry.NoSymmetry:
		generalsPos := make([]geometry.Pos, 0)
		for i := 0; i < nArmies; i++ {
			generalPos := w.GetRandomFreePos()
			for utils.Contains(generalsPos, generalPos) {
				generalPos = w.GetRandomFreePos()
			}
			generalsPos = append(generalsPos, generalPos)
		}
		for allegiance, general := range generalsPos {
			armyPos := w.Cluster(general, len(composition))
			for i, unitType := range composition {
				w.SpawnAlliedAt(strategy, allegiance+1, armyPos[i], unitType)
			}
		}
		return
	case geometry.VerticalSymmetry:
		armiesIdx = []int{0, 1} //cap to 2
	case geometry.HorizontalSymmetry:
		armiesIdx = []int{0, 1} //cap to 2
	case geometry.DoubleSymmetry:
		switch nArmies {
		case 4:
			armiesIdx = []int{0, 1, 2, 3}
		case 3:
			armiesIdx = []int{0, 1, 2}
		case 2:
			armiesIdx = []int{0, 3}
		default: //5 or higher, we cap it to 4
			armiesIdx = []int{0, 1, 2, 3}
		}
	}
	generalPos := w.anchor()
	armyPos := w.Cluster(generalPos, len(composition))
	for k, a := range armiesIdx {
		for i, unitType := range composition {
			w.SpawnAlliedAt(strategy, k+1, w.MapCfg.Symmetry.Transform(armyPos[i], w.MapCfg.Size)[a], unitType)
		}
	}
}
