package world

import (
	"container/heap"
	"fmt"
	"slices"
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
// the grid directly, they only receive Orders and reply with Actions (see
// Turn), which is what lets World hold *agents.Unit references without
// agents importing world back.
type World struct {
	Grid      [][]Tile
	Size      int
	Printable map[Tile]string // text representation of each Tile, used by PrintWorld
	FreeTiles []geometry.Pos  // tiles that were free at world creation time (used to spawn units)
	Units     []*agents.Unit  // every unit currently in the world
	Armies    []*agents.Army  // every army formed so far, one per distinct allegiance seen
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

// GetUnitAt returns the unit occupying pos, or nil if there isn't one. This
// is an O(n) scan over Units - fine for the small simulations this runs at,
// but worth revisiting if the unit count grows large.
func (w *World) GetUnitAt(pos geometry.Pos) *agents.Unit {
	for _, u := range w.Units {
		if u.Position == pos {
			return u
		}
	}
	return nil
}

// Made by Claude : (idk the color code)
// unitColors is the palette PrintWorld cycles through for occupied tiles, by
// the order each army first appeared in Armies (wrapping if there are more
// armies than colors). ANSI SGR codes: 31/32/34/35/36 = red/green/blue/
// magenta/cyan.
var unitColors = []string{"\033[31m", "\033[32m", "\033[34m", "\033[35m", "\033[36m"}

const colorReset = "\033[0m"

// PrintWorld renders the grid to stdout using the Printable representation,
// coloring each occupied tile by the army of the unit standing on it.
func (w *World) PrintWorld() {
	for i := range w.Grid {
		for j := range w.Grid[i] {
			pos := geometry.Pos{X: i, Y: j}
			tile := w.GetTile(pos)
			toPrint := w.Printable[tile]
			if tile == occupied {
				if u := w.GetUnitAt(pos); u != nil && u.Army != nil {
					toPrint = w.colorFor(u.Army) + toPrint + colorReset
				}
			}
			fmt.Print(toPrint)
			fmt.Printf(" ")
		}
		fmt.Printf("\n")
	}
}

// colorFor returns the ANSI color code for army, picked from unitColors by
// the order army first appeared in Armies.
func (w *World) colorFor(army *agents.Army) string {
	for i, a := range w.Armies {
		if a == army {
			return unitColors[i%len(unitColors)]
		}
	}
	return ""
}

// Made by Claude :
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

// SpawnUnit creates a solo unit with the given strategy, places it on a
// random free tile, registers it with the world and starts its goroutine.
func (w *World) SpawnUnit(strategy agents.Strategy) *agents.Unit {
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
func (w *World) SpawnAllied(strategy agents.Strategy, allegiance int) *agents.Unit {
	return w.spawn(agents.NewUnit(strategy, allegiance, agents.Infantry), w.GetRandomFreePos())
}

// SpawnAlliedAt is like SpawnAllied but places the unit at a specific tile,
// panicking if that tile isn't free.
func (w *World) SpawnAlliedAt(strategy agents.Strategy, allegiance int, p geometry.Pos) *agents.Unit {
	if !w.IsFree(p) {
		panic(fmt.Sprintf("cannot spawn unit at position (%d, %d): tile is not free", p.X, p.Y))
	}
	return w.spawn(agents.NewUnit(strategy, allegiance, agents.Infantry), p)
}

func (w *World) spawn(u *agents.Unit, p geometry.Pos) *agents.Unit {
	u.Position = p
	w.SetTile(p, occupied)
	w.Units = append(w.Units, u)
	w.enlist(u)
	u.Simulate()
	return u
}

// enlist finds the Army matching u's allegiance, or forms a new one with u
// as general if none exists yet, and wires u.Army so the unit's own strategy
// can consult it directly.
func (w *World) enlist(u *agents.Unit) {
	for _, army := range w.Armies {
		if army.Allegiance == u.Allegiance {
			army.Soldiers = append(army.Soldiers, u)
			u.Army = army
			return
		}
	}
	army := agents.NewArmy(u)
	u.Army = army
	w.Armies = append(w.Armies, army)
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
	if !w.reachable(u.Position, p, u.Speed, w.IsFree) {
		return
	}
	w.SetTile(u.Position, free)
	u.Position = p
	w.SetTile(u.Position, occupied)
}

// GetReachableTiles returns every tile within speed steps of pos that a unit
// standing there can see it could reach: the destination and every tile on
// the way must be in visible and free right now, so a unit only commits to
// moves through ground it's actually looking at. Anything it returns is
// also really reachable, so Move's own check on the real grid accepts it.
func (w *World) GetReachableTiles(pos geometry.Pos, speed int, visible []geometry.Pos) []geometry.Pos {
	visibleSet := make(map[geometry.Pos]bool, len(visible))
	for _, p := range visible {
		visibleSet[p] = true
	}
	passable := func(p geometry.Pos) bool { return visibleSet[p] && w.IsFree(p) }

	reachableTiles := make([]geometry.Pos, 0)
	for i := pos.X - speed; i <= pos.X+speed; i++ {
		for j := pos.Y - speed; j <= pos.Y+speed; j++ {
			candidate := geometry.Pos{X: i, Y: j}
			if w.reachable(pos, candidate, speed, passable) {
				reachableTiles = append(reachableTiles, candidate)
			}
		}
	}
	return reachableTiles
}

// Made by Claude:
// reachable reports whether target can be reached from pos within speed
// steps, using a breadth-first search through tiles accepted by passable
// and moving one step at a time along the four cardinal directions.
func (w *World) reachable(pos geometry.Pos, target geometry.Pos, speed int, passable func(geometry.Pos) bool) bool {
	if target == pos {
		return true
	}
	if speed <= 0 || w.OutOfRange(target) || !passable(target) {
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
			if visited[next] || w.OutOfRange(next) || !passable(next) {
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

// region pathfinding
// Made by Claude

// pathNode is one entry in the A* open set: a candidate tile, the cost to
// reach it so far (g), and its estimated total cost to the goal (f = g +
// heuristic).
type pathNode struct {
	pos geometry.Pos
	g   int
	f   int
}

// pathQueue is a min-heap of pathNode ordered by f, used as A*'s open set.
type pathQueue []pathNode

func (q pathQueue) Len() int            { return len(q) }
func (q pathQueue) Less(i, j int) bool  { return q[i].f < q[j].f }
func (q pathQueue) Swap(i, j int)       { q[i], q[j] = q[j], q[i] }
func (q *pathQueue) Push(x interface{}) { *q = append(*q, x.(pathNode)) }
func (q *pathQueue) Pop() interface{} {
	old := *q
	n := len(old)
	node := old[n-1]
	*q = old[:n-1]
	return node
}

// knownPathLength returns the walking distance from `from` to `to` as far as
// a unit can tell from its terrain memory: A* (Manhattan heuristic, which is
// admissible and consistent on a uniform-cost grid, so the result is the
// shortest route) restricted to tiles in known that are passable and not
// blocked by a unit it currently sees. Tiles never seen are treated as
// impassable - a unit won't plan a route through terrain it knows nothing
// about. `to` itself is always a valid destination, seen or not, occupied
// or not: this measures distance to a tile, not whether you could stand on it.
//
// When no known route reaches `to`, it falls back to exploring: the cost of
// the best known route to a frontier tile (a known tile bordering unseen
// terrain) plus that tile's straight-line distance to `to`, offset by a
// penalty larger than any real route on the grid so a known route always
// beats a guess. Walking to the frontier reveals more terrain, and the unit
// replans next turn. Returns -1 only if `to` is unreachable and there's no
// frontier left to explore either.
func (w *World) knownPathLength(from, to geometry.Pos, known agents.KnownMap, blocked map[geometry.Pos]bool) int {
	if from == to {
		return 0
	}
	directions := [...]geometry.Pos{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}}

	open := &pathQueue{{pos: from, g: 0, f: geometry.ManhattanDistance(from, to)}}
	heap.Init(open)
	best := map[geometry.Pos]int{from: 0}
	bestFrontier := -1

	for open.Len() > 0 {
		current := heap.Pop(open).(pathNode)
		if current.pos == to {
			return current.g
		}
		if current.g > best[current.pos] {
			continue // a cheaper path to this tile was already found
		}

		for _, d := range directions {
			next := geometry.Pos{X: current.pos.X + d.X, Y: current.pos.Y + d.Y}
			if w.OutOfRange(next) {
				continue
			}
			if _, seen := known[next]; !seen && next != to {
				// current borders unseen terrain: it's a frontier tile
				if est := current.g + geometry.ManhattanDistance(current.pos, to); bestFrontier < 0 || est < bestFrontier {
					bestFrontier = est
				}
				continue
			}
			if next != to && (!known[next] || blocked[next]) {
				continue
			}
			g := current.g + 1
			if existing, seen := best[next]; seen && existing <= g {
				continue
			}
			best[next] = g
			heap.Push(open, pathNode{pos: next, g: g, f: g + geometry.ManhattanDistance(next, to)})
		}
	}

	if bestFrontier < 0 {
		return -1
	}
	return w.Size*w.Size + bestFrontier
}

// observe updates u's terrain memory (and its army's, if it coordinates)
// with every tile it currently sees.
func (w *World) observe(u *agents.Unit, vision []geometry.Pos) {
	for _, p := range vision {
		u.Remember(p, w.GetTile(p) != obstacle)
	}
}

// distanceFor builds u's DistanceFunc for this turn: pathfinding over the
// terrain u plans with (see Unit.Knowledge), where only the units u
// currently sees count as obstacles - it doesn't know where unseen units are.
func (w *World) distanceFor(u *agents.Unit, vision []geometry.Pos) agents.DistanceFunc {
	blocked := make(map[geometry.Pos]bool)
	for _, p := range vision {
		if p != u.Position && w.GetTile(p) == occupied {
			blocked[p] = true
		}
	}
	known := u.Knowledge()
	return func(a, b geometry.Pos) int { return w.knownPathLength(a, b, known, blocked) }
}

// region vision
// Made by Claude

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

// GetVisibleUnits returns every other unit currently visible to self, i.e.
// whose position is among the tiles self sees. It hands back the units
// themselves (not just their positions) so a strategy can read whatever
// else a unit carries - allegiance today, a unit type later - without World
// having to pick out fields on its behalf.
func (w *World) GetVisibleUnits(self *agents.Unit, visible []geometry.Pos) []*agents.Unit {
	visibleSet := make(map[geometry.Pos]bool, len(visible))
	for _, p := range visible {
		visibleSet[p] = true
	}
	targets := make([]*agents.Unit, 0)
	for _, other := range w.Units {
		if other == self {
			continue
		}
		if visibleSet[other.Position] {
			targets = append(targets, other)
		}
	}
	return targets
}

//region CombatResolution

func (w *World) ResolveDamage(target *agents.Unit, damage int) {
	target.LifePoints -= damage
	if target.LifePoints <= 0 {
		//remove the reference to the unit in the world
		idx := slices.Index(w.Units, target)
		w.Units = slices.Delete(w.Units, idx, idx+1)
		w.SetTile(target.Position, free)
		//Let the unit manage its destruction further
		target.Alive = false
	}
}

// region Simulate

// Turn runs one simulation step. It first updates every army's shared state
// (phase, terrain memory, whether its target is in sight, where it heads
// this turn) - that happens here, sequentially, in this single goroutine,
// before any unit starts deciding, so it's safe for every soldier's
// concurrent decision to read that state afterwards without a race. It then
// sends every unit this turn's vision, reachable tiles and shootable
// targets, and resolves the Action it replies with. Units decide
// concurrently (each in its own goroutine, started by Simulate), while the
// final two sequential loops keep every Grid mutation serialized in World,
// avoiding data races on the shared grid.
func (w *World) Turn() {
	for _, army := range w.Armies {
		army.BeginTurn()
		army.UpdatePhase()
	}

	// Every memory write (observe, Perceive, ResolveGoal) happens before any
	// unit decides. Each Distance closure only reads that memory lazily while
	// its unit decides, so by then its army's shared map already holds every
	// soldier's sightings from this turn, and nothing writes to it concurrently.
	orders := make(map[*agents.Unit]agents.Order, len(w.Units))
	for _, u := range w.Units {
		vision := w.GetVisibleTiles(u.Position, u.VisionRange)
		w.observe(u, vision)
		targets := w.GetVisibleUnits(u, vision)
		if u.Army != nil {
			u.Army.Perceive(targets)
		}
		orders[u] = agents.Order{
			Vision:    vision,
			Reachable: w.GetReachableTiles(u.Position, u.Speed, vision),
			Targets:   targets,
			Distance:  w.distanceFor(u, vision),
		}
	}
	for _, army := range w.Armies {
		army.ResolveGoal(w.Size)
	}

	for _, u := range w.Units {
		u.Orders <- orders[u]
	}
	//The order of resolution is shooting -> moving
	movingActions := make([]utils.Tuple[*agents.Unit, geometry.Pos], 0)
	shootingActions := make([]utils.Tuple[*agents.Unit, int], 0)
	for _, u := range w.Units {
		switch action := <-u.Actions; action.Kind {
		case agents.Move:
			movingActions = append(movingActions, utils.Tuple[*agents.Unit, geometry.Pos]{First: u, Second: action.Target})
		case agents.Shoot:
			shootingActions = append(shootingActions, utils.Tuple[*agents.Unit, int]{First: w.GetUnitAt(action.Target), Second: u.ShootingDamage})
		case agents.Stay:
			// deliberately do nothing: the unit holds its position
		}
	}
	for _, a := range shootingActions {
		if a.First.Alive { //It's possible that the unit is already dead (someone shoot it already)
			w.ResolveDamage(a.First, a.Second)
		}
	}
	for _, a := range movingActions {
		if a.First.Alive { //If the unit's still alive after the shooting phase, it can move
			w.Move(a.First, a.Second)
		}
	}
}
