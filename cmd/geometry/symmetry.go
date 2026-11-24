package geometry

type Symmetry int //Type of symmetry possible in map generation
const (
	NoSymmetry         Symmetry = iota
	VerticalSymmetry            //On Y-axis ->left/right
	HorizontalSymmetry          //On X-axis ->Top/down
	DoubleSymmetry              //Map is divided in mirrored quarters
)

func (s Symmetry) Transform(p Pos, size int) []Pos {
	mirrorX := Pos{X: p.X, Y: size - 1 - p.Y}         // Top/down
	mirrorY := Pos{X: size - 1 - p.X, Y: p.Y}         // left/right
	both := Pos{X: size - 1 - p.X, Y: size - 1 - p.Y} // both
	switch s {
	case VerticalSymmetry:
		return []Pos{p, mirrorY}
	case HorizontalSymmetry:
		return []Pos{p, mirrorX}
	case DoubleSymmetry:
		return []Pos{p, mirrorY, mirrorX, both}
	default:
		return []Pos{p}
	}
}

// This function verifies that the position is in the original domain
// of the symmetry (in Vertical symmetry, left part of the map is the domain)
// This function will likely be used for army spawn logic
func (s Symmetry) InDomain(p Pos, size int) bool {
	switch s {
	case VerticalSymmetry:
		return p.X < size-1-p.X //in vertical symmetry, original x is always lower than its image
	case HorizontalSymmetry:
		return p.Y < size-1-p.Y //same
	case DoubleSymmetry:
		return (p.X < size-1-p.X) && (p.Y < size-1-p.Y) //must respect both obviously
	case NoSymmetry:
		return true //every position is in the domain
	default:
		panic("InDomain : unknown symmetry")
	}
}
