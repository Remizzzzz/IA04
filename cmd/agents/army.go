package agents

import "soldier/cmd/geometry"

// ArmyPhase identifies which stage of the default army strategy an army is
// in: it first regroups around its general, then advances together once
// every soldier has caught up.
type ArmyPhase int

const (
	Regrouping ArmyPhase = iota // soldiers converge on the general (zero value: armies start here)
	Advancing                   // the army moves together on Enemy
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
	Enemy      *Unit // first enemy spotted by any soldier once regrouped; the whole army then advances on it
	progress   map[*Unit]regroupProgress
}

// NewArmy creates an army led by general, who is also its first soldier.
func NewArmy(general *Unit) *Army {
	soldiers := make([]*Unit, 0)
	soldiers = append(soldiers, general)
	return &Army{Allegiance: general.Allegiance, Soldiers: soldiers, General: general, progress: make(map[*Unit]regroupProgress)}
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
// the first enemy spotted by any soldier becomes the whole army's target
// for the advancing phase. Call this once per soldier per turn, before any
// soldier decides its action - that ordering is what makes it safe for every
// soldier's concurrent Decide call to read Enemy afterwards without a race.
func (a *Army) Perceive(targets []*Unit) {
	if a.Enemy != nil {
		return
	}
	for _, t := range targets {
		if t.Allegiance != a.Allegiance {
			a.Enemy = t
			return
		}
	}
}

// Decide returns the Action u should take under this army's default
// strategy: regroup around the general, swerving away from any enemy a
// soldier personally spots along the way, then once regrouped advance
// together on the first enemy the army spotted. Deliberately simple for
// now - it will grow once units carry types with their own behaviors.
func (a *Army) Decide(u *Unit, reachable []geometry.Pos, targets []*Unit) Action {
	if enemy := nearestEnemy(u, targets); enemy != nil && a.Phase == Regrouping {
		return moveAwayFrom(u, reachable, enemy.Position)
	}

	switch a.Phase {
	case Regrouping:
		if u == a.General {
			return Action{Kind: Stay} // the general anchors the rally point
		}
		return moveToward(u, reachable, a.General.Position)
	case Advancing:
		if a.Enemy == nil {
			return Action{Kind: Stay} // nothing spotted yet to advance on
		}
		return moveToward(u, reachable, a.Enemy.Position)
	}
	return Action{Kind: Stay}
}

// nearestEnemy returns the closest of targets that isn't u's own allegiance,
// or nil if none of them are.
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

// moveToward picks the reachable tile that gets u closest to target.
func moveToward(u *Unit, reachable []geometry.Pos, target geometry.Pos) Action {
	best, bestDist := u.Position, geometry.ManhattanDistance(u.Position, target)
	for _, p := range reachable {
		if d := geometry.ManhattanDistance(p, target); d < bestDist {
			best, bestDist = p, d
		}
	}
	if best == u.Position {
		return Action{Kind: Stay}
	}
	return Action{Kind: Move, Target: best}
}

// moveAwayFrom picks the reachable tile that puts the most distance between
// u and threat.
func moveAwayFrom(u *Unit, reachable []geometry.Pos, threat geometry.Pos) Action {
	best, bestDist := u.Position, geometry.ManhattanDistance(u.Position, threat)
	for _, p := range reachable {
		if d := geometry.ManhattanDistance(p, threat); d > bestDist {
			best, bestDist = p, d
		}
	}
	if best == u.Position {
		return Action{Kind: Stay}
	}
	return Action{Kind: Move, Target: best}
}
