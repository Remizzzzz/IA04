package world

import (
	"fmt"
	"soldier/cmd/agents"
	"soldier/cmd/geometry"
	"soldier/cmd/utils"
)

// Tile is the state of a single grid cell.
type Tile int

const (
	free Tile = iota
	occupied
	obstacle
)

// RandCase returns obstacle with roughly probabilitY% chance, free otherwise.
// probabilitY must be between 0 and 100 inclusive.
func RandCase(probabilitY int) Tile {
	if probabilitY > 100 {
		panic("Invalid probabilitY (over 100%)")
	} else if probabilitY < 0 {
		panic("Invalid probabilitY (under 0%)")
	}
	choice := utils.RandInt(0, 100)
	result := choice >= probabilitY
	if result {
		return free
	} else {
		return obstacle
	}
}

// World holds the grid and every unit in it. World is the sole owner of
// spatial logic (movement, reachability, line of sight): units never touch
// the grid directly, they only receive Orders and reply with Moves (see
// Turn), which is what lets World hold *agents.Unit references without
// agents importing world back.
type World struct {
	Grid      [][]Tile
	Size      int
	Printable map[Tile]string // text representation of each Tile, used by PrintWorld
	FreeTiles []geometry.Pos  // tiles that were free at world creation time (used to spawn units)
	Units     []*agents.Unit  // every unit currently in the world
}

// NewWorld creates an N x N grid where each tile is randomly an obstacle with
// probabilitY% chance, and free otherwise.
func NewWorld(N int, probabilitY int) *World {
	grid := make([][]Tile, N)
	FreePos := make([]geometry.Pos, 0)
	for i := range grid {
		grid[i] = make([]Tile, N)
		for j := range grid[i] {
			grid[i][j] = RandCase(probabilitY)
			if grid[i][j] == free {
				FreePos = append(FreePos, geometry.Pos{X: i, Y: j})
			}
		}
	}
	print := make(map[Tile]string)
	print[free] = "[ ]"
	print[occupied] = "[X]"
	print[obstacle] = "[O]"
	return &World{Grid: grid, Size: N, Printable: print, FreeTiles: FreePos}
}

// SetTile sets the tile value at pos.
func (w *World) SetTile(pos geometry.Pos, c Tile) {
	w.Grid[pos.X][pos.Y] = c
}

// GetTile returns the tile value at pos.
func (w *World) GetTile(pos geometry.Pos) Tile {
	return w.Grid[pos.X][pos.Y]
}

// PrintWorld renders the grid to stdout using the Printable representation.
func (w *World) PrintWorld() {
	for i := range w.Grid {
		for j := range w.Grid[i] {
			to_print := w.Printable[w.GetTile(geometry.Pos{X: i, Y: j})]
			fmt.Print(to_print)
			fmt.Printf(" ")
		}
		fmt.Printf("\n")
	}
}

// GetNeighbours returns the up to 8 tiles surrounding pos (within grid bounds).
func (w *World) GetNeighbours(pos geometry.Pos) []geometry.Pos {
	neighbours := make([]geometry.Pos, 0)
	for dX := -1; dX <= 1; dX++ {
		for dY := -1; dY <= 1; dY++ {
			if dX == 0 && dY == 0 {
				continue
			}
			nextX := pos.X + dX
			nextY := pos.Y + dY
			if nextX < 0 || nextX >= w.Size || nextY < 0 || nextY >= w.Size {
				continue
			}
			neighbours = append(neighbours, geometry.Pos{X: nextX, Y: nextY})
		}
	}
	return neighbours
}

// GetRandomPos returns a uniformly random position within the grid bounds
// (may or may not be free).
func (w *World) GetRandomPos() geometry.Pos {
	return geometry.Pos{X: utils.RandInt(0, w.Size), Y: utils.RandInt(0, w.Size)}
}

// GetRandomFreePos returns a random position among the tiles that were free
// at world creation time.
func (w *World) GetRandomFreePos() geometry.Pos {
	return w.FreeTiles[utils.RandInt(0, len(w.FreeTiles))]
}

// SpawnUnit creates a unit with the given strategy, places it on a random
// free tile, registers it with the world and starts its goroutine.
func (w *World) SpawnUnit(strategy agents.Strategy) *agents.Unit {
	return w.spawn(strategy, w.GetRandomFreePos())
}

// SpawnUnitAt is like SpawnUnit but places the unit at a specific tile,
// panicking if that tile isn't free.
func (w *World) SpawnUnitAt(strategy agents.Strategy, p geometry.Pos) *agents.Unit {
	if !w.IsFree(p) {
		panic(fmt.Sprintf("cannot spawn unit at position (%d, %d): tile is not free", p.X, p.Y))
	}
	return w.spawn(strategy, p)
}

func (w *World) spawn(strategy agents.Strategy, p geometry.Pos) *agents.Unit {
	u := agents.NewUnit(strategy)
	u.Position = p
	w.SetTile(p, occupied)
	w.Units = append(w.Units, u)
	u.Simulate()
	return u
}

//region Test

// OutOfRange reports whether p falls outside the grid bounds.
func (w *World) OutOfRange(p geometry.Pos) bool {
	return p.X < 0 || p.X >= w.Size || p.Y < 0 || p.Y >= w.Size
}

// IsFree reports whether p is within bounds and not occupied/an obstacle.
func (w *World) IsFree(p geometry.Pos) bool {
	if w.OutOfRange(p) {
		return false
	}
	return w.GetTile(p) == free
}

// region movement

// Move relocates unit u to p if p is reachable from its current position,
// freeing its old tile and marking the new one occupied.
func (w *World) Move(u *agents.Unit, p geometry.Pos) {
	if w.OutOfRange(p) || !w.IsFree(p) {
		return
	}
	if !w.reachable(u.Position, p, u.Speed) {
		return
	}
	w.SetTile(u.Position, free)
	u.Position = p
	w.SetTile(u.Position, occupied)
}

// GetReachableTiles returns every tile within speed steps of pos that can
// actually be reached without crossing obstacles or occupied tiles.
func (w *World) GetReachableTiles(pos geometry.Pos, speed int) []geometry.Pos {
	reachableTiles := make([]geometry.Pos, 0)
	for i := pos.X - speed; i <= pos.X+speed; i++ {
		for j := pos.Y - speed; j <= pos.Y+speed; j++ {
			candidate := geometry.Pos{X: i, Y: j}
			if w.reachable(pos, candidate, speed) {
				reachableTiles = append(reachableTiles, candidate)
			}
		}
	}
	return reachableTiles
}

// reachable reports whether target can be reached from pos within speed
// steps, using a breadth-first search restricted to free tiles and moving
// one step at a time along the four cardinal directions.
func (w *World) reachable(pos geometry.Pos, target geometry.Pos, speed int) bool {
	if target == pos {
		return true
	}
	if speed <= 0 || w.OutOfRange(target) || !w.IsFree(target) {
		return false
	}

	type node struct {
		pos  geometry.Pos
		step int
	}
	queue := []node{{pos: pos}}
	visited := map[geometry.Pos]bool{pos: true}
	directions := [...]geometry.Pos{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}} // up/down/left/right

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current.step >= speed {
			continue // reached the speed limit along this path, don't expand further
		}

		for _, direction := range directions {
			next := geometry.Pos{X: current.pos.X + direction.X, Y: current.pos.Y + direction.Y}
			if visited[next] || w.OutOfRange(next) || !w.IsFree(next) {
				continue
			}
			if next == target {
				return true
			}
			visited[next] = true
			queue = append(queue, node{pos: next, step: current.step + 1})
		}
	}

	return false
}

// region vision

// GetVisibleTiles returns every in-bounds tile within visionRange of pos
// that is in direct line of sight of pos. This is where a future height map
// would plug in too: a taller obstacle would block more lines of sight than
// a shorter one.
func (w *World) GetVisibleTiles(pos geometry.Pos, visionRange int) []geometry.Pos {
	visible := make([]geometry.Pos, 0)
	for i := pos.X - visionRange; i <= pos.X+visionRange; i++ {
		for j := pos.Y - visionRange; j <= pos.Y+visionRange; j++ {
			target := geometry.Pos{X: i, Y: j}
			if w.OutOfRange(target) {
				continue
			}
			if w.hasLineOfSight(pos, target) {
				visible = append(visible, target)
			}
		}
	}
	return visible
}

// hasLineOfSight reports whether target is visible from pos: it walks the
// straight line between the two tiles (Bresenham's line algorithm) and fails
// as soon as it crosses an obstacle before reaching target. The target tile
// itself is always visible even if it is an obstacle (you can see a wall,
// just not through it).
func (w *World) hasLineOfSight(pos geometry.Pos, target geometry.Pos) bool {
	x0, y0 := pos.X, pos.Y
	x1, y1 := target.X, target.Y
	dx, dy := abs(x1-x0), abs(y1-y0)
	sx, sy := sign(x1-x0), sign(y1-y0)
	x, y := x0, y0
	err := dx - dy

	for {
		if x == x1 && y == y1 {
			return true
		}
		if (x != x0 || y != y0) && w.GetTile(geometry.Pos{X: x, Y: y}) == obstacle {
			return false // blocked before reaching target
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func sign(n int) int {
	switch {
	case n > 0:
		return 1
	case n < 0:
		return -1
	default:
		return 0
	}
}

// region Simulate

// Turn runs one simulation step: every unit is sent this turn's vision and
// reachable tiles and replies with where it wants to move; World then
// applies every move. Units decide concurrently (each in its own goroutine,
// started by Simulate), while the two sequential loops below keep every
// Grid mutation serialized in World, avoiding data races on the shared grid.
func (w *World) Turn() {
	for _, u := range w.Units {
		u.Orders <- agents.Order{
			Vision:    w.GetVisibleTiles(u.Position, u.VisionRange),
			Reachable: w.GetReachableTiles(u.Position, u.Speed),
			Cmd:       "MOVE",
		}
	}
	for _, u := range w.Units {
		w.Move(u, <-u.Moves)
	}
}
