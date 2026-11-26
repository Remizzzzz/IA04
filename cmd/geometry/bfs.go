package geometry

// Made by Claude :
// BFS runs a breadth-first search on a size x size grid, moving one step at a
// time along the four cardinal directions (the same moves units make). Tiles
// are visited closest first (in number of steps from the nearest start), like
// a flood fill spreading outward ring by ring, and never twice.
//
// Parameters:
//   - size: side of the grid; the search never steps outside of it.
//   - starts: the tiles the search starts from (one or several). They are
//     neither checked by passable nor passed to visit: they are where the
//     search stands, not where it goes (e.g. a unit's own, occupied tile).
//   - passable: "may the search step onto this tile?" - called on every
//     in-bounds neighbour not visited yet; tiles it rejects are skipped.
//   - visit: "what to do with this newly reached tile, and should the search
//     go on?" - called once per tile p that passable accepted, with from the
//     tile it was reached from (one step closer to the starts; following from
//     back gives a shortest path). Return true to keep searching, false to
//     stop the whole search right away.
//
// BFS returns nothing: visit is where callers record what they need, in
// variables declared around the call (e.g. appending p to a slice).
//
// Example - the count tiles closest to start, walking only on free tiles:
//
//	tiles := []Pos{start}
//	BFS(size, []Pos{start}, isFree, func(p, _ Pos) bool {
//		tiles = append(tiles, p)
//		return len(tiles) < count
//	})
//
// The grid itself is never touched: everything BFS needs to know about it
// goes through passable, which is what lets one BFS serve every caller.
func BFS(size int, starts []Pos, passable func(Pos) bool, visit func(p, from Pos) bool) {
	directions := [...]Pos{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}} // right/left/down/up

	visited := make(map[Pos]bool, len(starts))
	queue := make([]Pos, 0, len(starts)) // tiles reached whose neighbours haven't been looked at yet
	for _, s := range starts {
		if !visited[s] {
			visited[s] = true
			queue = append(queue, s)
		}
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, d := range directions {
			next := Pos{X: current.X + d.X, Y: current.Y + d.Y}
			if next.X < 0 || next.X >= size || next.Y < 0 || next.Y >= size {
				continue // outside the grid
			}
			if visited[next] || !passable(next) {
				continue // already reached, or can't step there
			}
			visited[next] = true // mark now so it's never queued twice
			if !visit(next, current) {
				return
			}
			queue = append(queue, next)
		}
	}
}
