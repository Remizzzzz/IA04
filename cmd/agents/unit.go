package agents

import (
	"soldier/cmd/utils"
	"soldier/cmd/world"
)

// Unit represents a single agent moving on the world grid.
type Unit struct {
	Map      *world.World
	Speed    int // max number of tiles the unit can travel in one move (BFS depth)
	Position world.Pos
	Strategy Strategy
	Command  chan string // channel used by the simulation loop to send orders (e.g. "MOVE") to this unit
	//WG       *sync.WaitGroup
}

// NewUnit creates a unit on the given world, spawning it on a random free tile.
func NewUnit(w *world.World, strategy Strategy) *Unit {
	c := make(chan string)
	return &Unit{Map: w, Speed: 1, Position: w.SpawnUnit(), Strategy: strategy, Command: c}
}
func NewUnitPos(w *world.World, strategy Strategy, spawnPosition world.Pos) *Unit {
	c := make(chan string)
	return &Unit{Map: w, Speed: 1, Position: w.SpawnUnitAt(spawnPosition), Strategy: strategy, Command: c}
}

// Move relocates the unit to p if p is within range and reachable,
// updating the grid so the old tile becomes free and the new tile occupied.
func (s *Unit) Move(p world.Pos) {
	if s.Map.OutOfRange(p) || !s.Map.IsFree(p) {
		return
	}
	if s.Reachable(p) {
		s.Map.SetTile(s.Position, world.Tile(0)) // Tile(0) is the "free" tile value
		s.Position = p
		s.Map.SetTile(s.Position, world.Tile(1)) // Tile(1) is the "occupied" tile value
	}
}

// GetReachableTiles returns every tile within the unit's speed range (bounding box)
// that can actually be reached without crossing obstacles or occupied tiles.
func (s *Unit) GetReachableTiles() []world.Pos {
	reachableTiles := make([]world.Pos, 0)
	for i := s.Position.X - s.Speed; i <= s.Position.X+s.Speed; i++ {
		for j := s.Position.Y - s.Speed; j <= s.Position.Y+s.Speed; j++ {
			if s.Reachable(world.Pos{X: i, Y: j}) {
				reachableTiles = append(reachableTiles, world.Pos{X: i, Y: j})
			}
		}
	}
	return reachableTiles
}

// region strategy

// Strategy identifies which movement policy a unit follows.
type Strategy int

const (
	Random Strategy = iota // pick a random reachable tile each turn
)

// strategyRandom picks a random tile among those currently reachable,
// excluding the unit's own position.
func (s *Unit) strategyRandom() world.Pos {
	reachable := s.GetReachableTiles()
	others := make([]world.Pos, 0, len(reachable))
	for _, p := range reachable {
		if p != s.Position {
			others = append(others, p)
		}
	}
	if len(others) == 0 {
		return s.Position // no free tile to move to, stay put
	}
	return others[utils.RandInt(0, len(others))]
}

// UseStrategy applies the unit's configured movement strategy for this turn.
func (s *Unit) UseStrategy() {
	switch s.Strategy {
	case Random:
		s.Move(s.strategyRandom())
	}
}

// region simulation

// Simulate starts the unit's event loop in its own goroutine: it waits for a
// "MOVE" order on Command, acts, then reports back on the world's Info channel
// so the simulation can tell when every unit has finished its turn.
func (s *Unit) Simulate() {
	go func() {
		for {
			c := <-s.Command
			if c == "MOVE" {
				s.UseStrategy()
			}
			s.Map.Info <- 0
		}
	}()
}

//region test

// Reachable reports whether p can be reached from the unit's current position
// within its Speed, using a breadth-first search restricted to free tiles and
// moving one step at a time along the four cardinal directions.
func (s *Unit) Reachable(p world.Pos) bool {
	if p == s.Position {
		return true
	}
	if s.Speed <= 0 || s.Map.OutOfRange(p) || !s.Map.IsFree(p) {
		return false
	}

	type node struct {
		pos  world.Pos
		step int
	}
	queue := []node{{pos: s.Position}}
	visited := map[world.Pos]bool{s.Position: true}
	directions := [...]world.Pos{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}} // up/down/left/right

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current.step >= s.Speed {
			continue // reached the speed limit along this path, don't expand further
		}

		for _, direction := range directions {
			next := world.Pos{X: current.pos.X + direction.X, Y: current.pos.Y + direction.Y}
			if visited[next] || s.Map.OutOfRange(next) || !s.Map.IsFree(next) {
				continue
			}
			if next == p {
				return true
			}
			visited[next] = true
			queue = append(queue, node{pos: next, step: current.step + 1})
		}
	}

	return false
}
