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
// Corridors are then carved through obstacles so the map is connected: every
// free tile is reachable from every other (see connectComponents).
func NewMapFromConfig(cfg MapConfig) ([][]Tile, []geometry.Pos) {
	grid := make([][]Tile, cfg.Size)
	filled := make([][]bool, cfg.Size) // free is Tile's zero value, so we track filled tiles separately
	for y := range grid {
		grid[y] = make([]Tile, cfg.Size)
		filled[y] = make([]bool, cfg.Size)
	}

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
			}
		}
	}

	connectComponents(grid, cfg.Symmetry)

	// The map is now a single connected zone, which holds every free tile.
	comps := freeComponents(grid)
	if len(comps) == 0 {
		panic("cannot generate map: every tile is an obstacle (check ObstacleFrequency)")
	}
	freePos := comps[0]
	return grid, freePos
}

// Made by Claude :
// connectComponents carves corridors through obstacles until grid's free
// tiles form a single connected zone, keeping grid symmetric under sym.
//
// While there is more than one zone, it runs a BFS that starts from every
// tile of the largest zone at once and, unlike freeComponents, is allowed to
// walk through obstacles. It stops at the first free tile outside the largest
// zone: the tiles it crossed to get there are all obstacles, and since a BFS
// finds the shortest route, they are the fewest obstacles that need removing
// to join that zone to the largest one. Those tiles, along with all their
// images under sym, are made free - carving whole orbits keeps the map
// symmetric, and also joins the mirror images of the zone in the same pass.
//
// It always terminates: each pass joins at least one zone to the largest one,
// and removing obstacles never splits a zone, so the number of zones strictly
// decreases. A grid with no free tile at all is left untouched.
func connectComponents(grid [][]Tile, sym geometry.Symmetry) {
	size := len(grid)
	directions := [...]geometry.Pos{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}} // right/left/down/up

	for {
		comps := freeComponents(grid)
		if len(comps) <= 1 {
			return // already connected (or no free tile at all)
		}
		largest := comps[0]
		for _, c := range comps[1:] {
			if len(c) > len(largest) {
				largest = c
			}
		}

		// BFS from every tile of the largest zone, through obstacles too.
		// parent[p] is the tile the BFS came from to reach p; tiles of the
		// largest zone are the starting points, so they have no parent.
		visited := make([][]bool, size)
		for y := range visited {
			visited[y] = make([]bool, size)
		}
		queue := make([]geometry.Pos, 0, len(largest))
		for _, p := range largest {
			visited[p.Y][p.X] = true
			queue = append(queue, p)
		}
		parent := make(map[geometry.Pos]geometry.Pos)

		var reached geometry.Pos // first free tile found outside the largest zone
		found := false
		for len(queue) > 0 && !found {
			current := queue[0]
			queue = queue[1:]
			for _, d := range directions {
				next := geometry.Pos{X: current.X + d.X, Y: current.Y + d.Y}
				if next.X < 0 || next.X >= size || next.Y < 0 || next.Y >= size || visited[next.Y][next.X] {
					continue // outside the grid, or already reached
				}
				visited[next.Y][next.X] = true
				parent[next] = current
				if grid[next.Y][next.X] == free {
					// every tile of the largest zone is already visited, so a
					// free tile reached now belongs to another zone
					reached, found = next, true
					break
				}
				queue = append(queue, next) // an obstacle: keep digging from it
			}
		}
		if !found {
			return // can't happen (the BFS covers the whole grid), but never loop forever
		}

		// Walk back from the reached tile to the largest zone, carving every
		// obstacle on the way, along with its symmetric images.
		for p, ok := parent[reached]; ok; p, ok = parent[p] {
			for _, img := range sym.Transform(p, size) {
				grid[img.Y][img.X] = free
			}
		}
	}
}

// Made by Claude :
// freeComponents splits the free tiles of grid into connected components: two
// free tiles are in the same component if a unit could walk from one to the
// other through free tiles, moving one step at a time along the four cardinal
// directions (the same moves units make, see World.reachable). Components are
// returned in discovery order (row by row), each one as the list of its tiles.
//
// It runs one breadth-first search (BFS) per component: starting from a free
// tile nobody has visited yet, it visits every free neighbour of that tile,
// then every free neighbour of those, and so on, like a flood fill spreading
// outward ring by ring, until it can't reach any new tile - everything it
// visited along the way is one component.
func freeComponents(grid [][]Tile) [][]geometry.Pos {
	size := len(grid)
	visited := make([][]bool, size) // visited[y][x]: tile already assigned to a component
	for y := range visited {
		visited[y] = make([]bool, size)
	}
	directions := [...]geometry.Pos{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}} // right/left/down/up

	components := make([][]geometry.Pos, 0)
	for y := range grid {
		for x := range grid[y] {
			if grid[y][x] != free || visited[y][x] {
				continue // an obstacle, or a tile an earlier BFS already reached
			}

			// New component: BFS from (x, y). queue holds the tiles reached but
			// whose neighbours haven't been looked at yet; component collects
			// every tile reached.
			start := geometry.Pos{X: x, Y: y}
			visited[y][x] = true
			queue := []geometry.Pos{start}
			component := []geometry.Pos{start}

			for len(queue) > 0 {
				current := queue[0]
				queue = queue[1:]
				for _, d := range directions {
					next := geometry.Pos{X: current.X + d.X, Y: current.Y + d.Y}
					if next.X < 0 || next.X >= size || next.Y < 0 || next.Y >= size {
						continue // outside the grid
					}
					if grid[next.Y][next.X] != free || visited[next.Y][next.X] {
						continue // can't walk there, or already reached
					}
					visited[next.Y][next.X] = true // mark now so it's never queued twice
					queue = append(queue, next)
					component = append(component, next)
				}
			}
			components = append(components, component)
		}
	}
	return components
}
