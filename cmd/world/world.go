package world

import (
	"fmt"
	"soldier/cmd/utils"
)

type Tile int

const (
	free Tile = iota
	occupied
	obstacle
)

type Pos struct {
	X int
	Y int
}

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

type World struct {
	Grid      [][]Tile
	Size      int
	Printable map[Tile]string
	FreeTiles []Pos
	Info      chan int
	nEntities int
	ReadyFlag bool
}

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

func (w *World) SetTile(pos Pos, c Tile) {
	w.Grid[pos.X][pos.Y] = c
}

func (w *World) GetTile(pos Pos) Tile {
	return w.Grid[pos.X][pos.Y]
}

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

func (w *World) GetRandomPos() Pos {
	return Pos{utils.RandInt(0, w.Size), utils.RandInt(0, w.Size)}
}

func (w *World) GetRandomFreePos() Pos {
	return w.FreeTiles[utils.RandInt(0, len(w.FreeTiles))]
}

func (w *World) SpawnUnit() Pos {
	p := w.GetRandomFreePos()
	w.SetTile(p, occupied)
	w.nEntities++
	return p
}

//region Test

func (w *World) OutOfRange(p Pos) bool {
	return p.X < 0 || p.X >= w.Size || p.Y < 0 || p.Y >= w.Size
}

func (w *World) IsFree(p Pos) bool {
	if w.OutOfRange(p) {
		return false
	}
	return w.GetTile(p) == free
}

// region Simulate
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
