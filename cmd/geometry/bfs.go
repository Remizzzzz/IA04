package geometry

// Made by Claude :
// BFS runs a breadth-first search on a size x size grid, moving one step at a
// time along the four cardinal directions (the same moves units make).
//
// It starts from every tile of starts at once, then visits the tiles next to
// them, then the tiles next to those, and so on, like a flood fill spreading
// outward ring by ring - so tiles are always visited closest first (in number
// of steps from the nearest start). It only steps onto in-bounds tiles that
// passable accepts, and never visits a tile twice.
//
// visit is called once for every newly reached tile p, with from the tile it
// was reached from (one step closer to the starts): following from back
// gives a shortest path to the starts. The search stops as soon as visit
// returns false. Starts are not passed to visit, nor checked by passable:
// they are where the search stands, not where it goes (e.g. a unit's own,
// occupied tile).
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
