package world

import "soldier/cmd/geometry"

type MapConfig struct {
	Size              int
	ObstacleFrequency int
	Symmetry          geometry.Symmetry
}

// NewMapFromConfig generates a Size x Size grid where each tile is an
// obstacle with ObstacleFrequency% chance, constrained by cfg.Symmetry: a
// tile and all its images under the symmetry always share the same value.
func NewMapFromConfig(cfg MapConfig) ([][]Tile, []geometry.Pos) {
	grid := make([][]Tile, cfg.Size)
	filled := make([][]bool, cfg.Size) // free is Tile's zero value, so we track filled tiles separately
	for y := range grid {
		grid[y] = make([]Tile, cfg.Size)
		filled[y] = make([]bool, cfg.Size)
	}
	freePos := make([]geometry.Pos, 0)

	// each tile not yet filled draws a value and gives it to its whole orbit.
	for y := range grid {
		for x := range grid[y] {
			if filled[y][x] {
				continue // already set as the image of an earlier tile
			}
			//If the tile is not filled, then we fill it with a random case then we fill its symmetric images
			tile := RandCase(cfg.ObstacleFrequency)
			for _, p := range cfg.Symmetry.Transform(geometry.Pos{X: x, Y: y}, cfg.Size) {
				if filled[p.Y][p.X] {
					continue //When we are on the axis, all returned positions are the same
				}
				grid[p.Y][p.X] = tile
				filled[p.Y][p.X] = true
				if tile == free {
					freePos = append(freePos, geometry.Pos{X: p.X, Y: p.Y})
				}
			}
		}
	}

	return grid, freePos
}
