package agents

import (
	"soldier/cmd/utils"
	"soldier/cmd/world"
)

type Soldier struct {
	Map      *world.World
	Speed    int
	Position world.Pos
	Strategy Strategy
	Command  chan string
	//WG       *sync.WaitGroup
}

func NewSoldier(w *world.World, Strategy Strategy) *Soldier {
	c := make(chan string)
	return &Soldier{Map: w, Speed: 1, Position: w.SpawnUnit(), Strategy: Strategy, Command: c}
}

func (s *Soldier) Move(p world.Pos) {
	if s.Map.OutOfRange(p) || !s.Map.IsFree(p) {
		return
	}
	if s.Reachable(p) {
		s.Map.SetTile(s.Position, world.Tile(0)) // free est la valeur initiale du type Tile
		s.Position = p
		s.Map.SetTile(s.Position, world.Tile(1))
	}
}

func (s *Soldier) GetReachableTiles() []world.Pos {
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
type Strategy int

const (
	Random Strategy = iota
)

func (s *Soldier) strategyRandom() world.Pos {
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

func (s *Soldier) UseStrategy() {
	switch s.Strategy {
	case Random:
		s.Move(s.strategyRandom())
	}
}

// region simulation
func (s *Soldier) Simulate() {
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

func (s *Soldier) Reachable(p world.Pos) bool {
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
	directions := [...]world.Pos{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current.step >= s.Speed {
			continue
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
