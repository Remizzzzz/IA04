package agents

import (
	"math"
	"soldier/cmd/geometry"
)

// ArmyPhase identifies which stage of the default army strategy an army is
// in: it first regroups around its general, then advances together once
// every soldier has caught up.
type ArmyPhase int

const (
	Regrouping ArmyPhase = iota // soldiers converge on the general (zero value: armies start here)
	Advancing                   // the army moves together on its target, or explores to find one
)

func (p ArmyPhase) String() string {
	switch p {
	case Regrouping:
		return "Regrouping"
	case Advancing:
		return "Advancing"
	default:
		return "Unknown"
	}
}

// RegroupRadius is how close every soldier must be to the general before the
// army considers itself regrouped.
const RegroupRadius = 2

// StagnationThreshold is how many consecutive turns a soldier can go without
// getting any closer to the general before it's considered stuck (blocked by
// terrain, or simply unable to find a free tile close enough - e.g. an army
// too big for RegroupRadius to hold all at once) and stops counting against
// the regroup check.
const StagnationThreshold = 2

// ArrivalRadius is how close a soldier must get to the target's last known
// position for the army to consider it has checked that spot.
const ArrivalRadius = 1

// regroupProgress tracks one soldier's progress toward the general, used to
// detect soldiers that are stuck rather than just slow.
type regroupProgress struct {
	lastDistance int
	stuckTurns   int
}

// Army groups every unit sharing an allegiance under a single general. It is
// formed the moment the first unit of an allegiance is spawned (see
// World.enlist); IsActive reports whether it actually has anyone to
// coordinate with yet.
type Army struct {
	Allegiance int
	Soldiers   []*Unit
	General    *Unit
	Phase      ArmyPhase
	Enemy      *Unit    // the enemy the army is hunting, nil while it has none and is exploring
	Known      KnownMap // terrain pooled from every Coordinated soldier's sightings
	progress   map[*Unit]regroupProgress

	clock        int                  // turns elapsed, used to timestamp sightings
	lastSeen     map[geometry.Pos]int // turn each known tile was last seen by a soldier
	enemyInSight bool                 // whether any soldier sees Enemy this turn
	lastKnown    geometry.Pos         // where Enemy was last seen
	goal         geometry.Pos         // where the army heads this turn while advancing
	hasGoal      bool
}

// NewArmy creates an army led by general, who is also its first soldier.
func NewArmy(general *Unit) *Army {
	soldiers := make([]*Unit, 0)
	soldiers = append(soldiers, general)
	return &Army{
		Allegiance: general.Allegiance,
		Soldiers:   soldiers,
		General:    general,
		Known:      make(KnownMap),
		progress:   make(map[*Unit]regroupProgress),
		lastSeen:   make(map[geometry.Pos]int),
	}
}

// BeginTurn starts a new turn of the army's shared awareness: it advances
// the clock used to timestamp sightings and forgets whether the target was
// in sight, since that has to be confirmed again this turn. Call it before
// any Perceive of the turn.
func (a *Army) BeginTurn() {
	a.clock++
	a.enemyInSight = false
}

// IsActive reports whether the army has at least one soldier besides its
// general, i.e. whether there's actually anyone to coordinate.
func (a *Army) IsActive() bool {
	return len(a.Soldiers) >= 2
}

// UpdatePhase recomputes whether the army has finished regrouping around its
// general. A soldier stuck for StagnationThreshold turns (see
// regroupProgress) no longer counts against the check, so an oversized or
// boxed-in army doesn't wait on it forever. The transition itself is
// one-way: once the army starts advancing it keeps advancing, even if
// soldiers later scatter again. Call this once per army per turn, before any
// soldier decides its action.
func (a *Army) UpdatePhase() {
	if a.Phase == Advancing {
		return
	}
	regrouped := true
	for _, s := range a.Soldiers {
		d := geometry.ManhattanDistance(s.Position, a.General.Position)

		progress, tracked := a.progress[s]
		if tracked && d < progress.lastDistance {
			progress.stuckTurns = 0
		} else if tracked {
			progress.stuckTurns++
		}
		progress.lastDistance = d
		a.progress[s] = progress

		if d > RegroupRadius && progress.stuckTurns < StagnationThreshold {
			regrouped = false
		}
	}
	if regrouped {
		a.Phase = Advancing
	}
}

// Perceive lets one soldier's sightings feed the army's shared awareness:
// if the army has no target yet, the first enemy any soldier spots becomes
// it; and whenever a soldier sees the target, the army knows it's in sight
// and records where it is. Call this once per soldier per turn, after
// BeginTurn and before any soldier decides its action - that ordering is
// what makes it safe for every soldier's concurrent Decide call to read the
// army's state afterwards without a race.
func (a *Army) Perceive(targets []*Unit) {
	for _, t := range targets {
		if t.Allegiance == a.Allegiance {
			continue
		}
		if a.Enemy == nil {
			a.Enemy = t
		}
		if t == a.Enemy {
			a.enemyInSight = true
			a.lastKnown = t.Position
		}
	}
}

// ResolveGoal decides where the army heads this turn once every soldier has
// perceived: the target itself while it's in sight, otherwise the spot where
// it was last seen. If a soldier reaches that spot and the target still
// isn't in sight, the army gives up on it and explores until it spots a new
// one. size is the width of the (square) battlefield, needed to know where
// unexplored ground can be. Call it after every Perceive of the turn.
func (a *Army) ResolveGoal(size int) {
	if a.Enemy != nil && !a.enemyInSight && a.anySoldierWithin(a.lastKnown, ArrivalRadius) {
		a.Enemy = nil // reached where it was last seen and it's gone: lost track of it
	}
	if a.Enemy != nil {
		a.goal, a.hasGoal = a.lastKnown, true
		return
	}
	a.goal, a.hasGoal = a.explorationGoal(size)
}

func (a *Army) anySoldierWithin(p geometry.Pos, radius int) bool {
	for _, s := range a.Soldiers {
		if geometry.ManhattanDistance(s.Position, p) <= radius {
			return true
		}
	}
	return false
}

// explorationGoal picks where to look for a new target: the unexplored tile
// bordering known ground that's closest to the general, or once everything
// reachable has been seen, the known tile seen longest ago (a patrol, since
// enemies move around and a tile seen long ago may hide one now). Returns
// false if every known tile is in sight right now, leaving nothing to check.
func (a *Army) explorationGoal(size int) (geometry.Pos, bool) {
	from := a.General.Position
	var best geometry.Pos
	bestDist := 0
	found := false

	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			p := geometry.Pos{X: x, Y: y}
			if _, seen := a.Known[p]; seen || !a.bordersKnownGround(p) {
				continue
			}
			if d := geometry.ManhattanDistance(from, p); !found || d < bestDist {
				best, bestDist, found = p, d, true
			}
		}
	}
	if found {
		return best, true
	}

	oldest := 0
	for p, passable := range a.Known {
		if !passable {
			continue
		}
		seen, d := a.lastSeen[p], geometry.ManhattanDistance(from, p)
		if !found || seen < oldest || (seen == oldest && d < bestDist) {
			best, bestDist, oldest, found = p, d, seen, true
		}
	}
	if !found || oldest == a.clock {
		return geometry.Pos{}, false
	}
	return best, true
}

// bordersKnownGround reports whether p has a known passable neighbor, i.e.
// whether the army could walk up to it to see what's there.
func (a *Army) bordersKnownGround(p geometry.Pos) bool {
	for _, d := range [...]geometry.Pos{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}} {
		if a.Known[geometry.Pos{X: p.X + d.X, Y: p.Y + d.Y}] {
			return true
		}
	}
	return false
}

// Decide returns the Action u should take under this army's default
// strategy: regroup around the general, swerving away from any enemy a
// soldier personally spots along the way, then once regrouped advance
// together on the army's goal for the turn (see ResolveGoal): the target
// while it's in sight, its last known position once lost, or unexplored
// ground while the army has no target. Deliberately simple for now - it
// will grow once units carry types with their own behaviors.
func (a *Army) Decide(u *Unit, reachable []geometry.Pos, targets []*Unit, distance DistanceFunc) Action {
	if enemy := nearestEnemy(u, targets); enemy != nil && a.Phase == Regrouping {
		return moveAwayFrom(u, reachable, enemy.Position, distance)
	}

	switch a.Phase {
	case Regrouping:
		if u == a.General {
			return Action{Kind: Stay} // the general anchors the rally point
		}
		return moveToward(u, reachable, a.General.Position, distance)
	case Advancing:
		if !a.hasGoal {
			return Action{Kind: Stay} // everything known is in sight and there's no target: hold
		}
		return moveToward(u, reachable, a.goal, distance)
	}
	return Action{Kind: Stay}
}

// nearestEnemy returns the closest of targets that isn't u's own allegiance,
// or nil if none of them are. This is just picking which visible enemy to
// react to, not plotting a route, so plain straight-line distance is enough.
func nearestEnemy(u *Unit, targets []*Unit) *Unit {
	var nearest *Unit
	best := 0
	for _, t := range targets {
		if t.Allegiance == u.Allegiance {
			continue
		}
		if d := geometry.ManhattanDistance(u.Position, t.Position); nearest == nil || d < best {
			nearest, best = t, d
		}
	}
	return nearest
}

// effectiveDistance treats an unreachable path (-1, no route exists at all)
// as farther than any real path on the grid: worse than any finite distance
// when trying to approach a target, and the safest possible outcome when
// trying to flee one.
func effectiveDistance(d int) int {
	if d < 0 {
		return math.MaxInt
	}
	return d
}

// moveToward picks the reachable tile with the shortest walking distance to
// target as far as the unit knows, using distance (an A* search over its
// terrain memory, see World.knownPathLength) rather than straight-line
// distance - that's what lets it route around an obstacle or another unit
// even when the first step of that route doesn't itself get any closer.
func moveToward(u *Unit, reachable []geometry.Pos, target geometry.Pos, distance DistanceFunc) Action {
	best, bestDist := u.Position, effectiveDistance(distance(u.Position, target))
	for _, p := range reachable {
		if d := effectiveDistance(distance(p, target)); d < bestDist {
			best, bestDist = p, d
		}
	}
	if best == u.Position {
		return Action{Kind: Stay}
	}
	return Action{Kind: Move, Target: best}
}

// moveAwayFrom picks the reachable tile with the longest true walking
// distance from threat, using the same A* distance oracle as moveToward but
// maximizing instead of minimizing.
func moveAwayFrom(u *Unit, reachable []geometry.Pos, threat geometry.Pos, distance DistanceFunc) Action {
	best, bestDist := u.Position, effectiveDistance(distance(u.Position, threat))
	for _, p := range reachable {
		if d := effectiveDistance(distance(p, threat)); d > bestDist {
			best, bestDist = p, d
		}
	}
	if best == u.Position {
		return Action{Kind: Stay}
	}
	return Action{Kind: Move, Target: best}
}
