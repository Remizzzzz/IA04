package world

import (
	"fmt"
	"soldier/cmd/utils"
)

// Tile is the state of a single grid cell.
type Tile int

const (
	free Tile = iota
	occupied
	obstacle
)

// Pos is a coordinate on the grid.
type Pos struct {
	X int
	Y int
}

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

// World holds the grid and the shared state used to synchronize a turn
// across all units.
type World struct {
	Grid      [][]Tile
	Size      int
	Printable map[Tile]string // text representation of each Tile, used by PrintWorld
	FreeTiles []Pos           // tiles that were free at world creation time (used to spawn units)
	Info      chan int        // units report back on this channel once they've acted on a turn
	nEntities int             // number of units spawned so far, used to know when a turn is complete
	ReadyFlag bool            // true when the world is ready to start a new turn
}

// NewWorld creates an N x N grid where each tile is randomly an obstacle with
// probabilitY% chance, and free otherwise.
func NewWorld(N int, probabilitY int) *World {
	grid := make([][]Tile, N)
	FreePos := make([]Pos, 0)
	for i := range grid {
		grid[i] = make([]Tile, N)
		for j := range grid[i] {
			grid[i][j] = RandCase(probabilitY)
			if grid[i][j] == free {
				FreePos = append(FreePos, Pos{i, j})
			}
		}
	}
	print := make(map[Tile]string)
	print[free] = "[ ]"
	print[occupied] = "[X]"
	print[obstacle] = "[O]"
	i := make(chan int)
	return &World{Grid: grid, Size: N, Printable: print, FreeTiles: FreePos, nEntities: 0, ReadyFlag: true, Info: i}
}

// SetTile sets the tile value at pos.
func (w *World) SetTile(pos Pos, c Tile) {
	w.Grid[pos.X][pos.Y] = c
}

// GetTile returns the tile value at pos.
func (w *World) GetTile(pos Pos) Tile {
	return w.Grid[pos.X][pos.Y]
}

// PrintWorld renders the grid to stdout using the Printable representation.
func (w *World) PrintWorld() {
	for i := range w.Grid {
		for j := range w.Grid[i] {
			to_print := w.Printable[w.GetTile(Pos{i, j})]
			fmt.Print(to_print)
			fmt.Printf(" ")
		}
		fmt.Printf("\n")
	}
}

// GetNeighbours returns the up to 8 tiles surrounding pos (within grid bounds).
func (w *World) GetNeighbours(pos Pos) []Pos {
	neighbours := make([]Pos, 0)
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
			neighbours = append(neighbours, Pos{nextX, nextY})
		}
	}
	return neighbours
}

// GetRandomPos returns a uniformly random position within the grid bounds
// (may or may not be free).
func (w *World) GetRandomPos() Pos {
	return Pos{utils.RandInt(0, w.Size), utils.RandInt(0, w.Size)}
}

// GetRandomFreePos returns a random position among the tiles that were free
// at world creation time.
func (w *World) GetRandomFreePos() Pos {
	return w.FreeTiles[utils.RandInt(0, len(w.FreeTiles))]
}

// SpawnUnit places a new unit on a random free tile, marks that tile as
// occupied, and increments the entity count used to track turn completion.
func (w *World) SpawnUnit() Pos {
	p := w.GetRandomFreePos()
	w.SetTile(p, occupied)
	w.nEntities++
	return p
}

// SpawnUnitAt spawn a new unit on a given tile if it's free
func (w *World) SpawnUnitAt(p Pos) Pos {
	if w.IsFree(p) {
		w.SetTile(p, occupied)
		w.nEntities++
	} else {
		panic(fmt.Sprintf("cannot spawn unit at position (%d, %d): tile is not free", p.X, p.Y))
	}
	return p
}

//region Test

// OutOfRange reports whether p falls outside the grid bounds.
func (w *World) OutOfRange(p Pos) bool {
	return p.X < 0 || p.X >= w.Size || p.Y < 0 || p.Y >= w.Size
}

// IsFree reports whether p is within bounds and not occupied/an obstacle.
func (w *World) IsFree(p Pos) bool {
	if w.OutOfRange(p) {
		return false
	}
	return w.GetTile(p) == free
}

// region Simulate

// Simulate starts a goroutine that counts how many units have reported back
// on Info after a turn was started; once every spawned unit (nEntities) has
// reported, it flips ReadyFlag so the simulation loop can start the next turn.
func (w *World) Simulate() {
	go func() {
		ready := 0
		for {
			if !w.ReadyFlag {
				n := <-w.Info
				if n == 0 {
					ready++
					if ready == w.nEntities {
						w.ReadyFlag = true
						ready = 0
					}
				}
			}
		}
	}()
}
