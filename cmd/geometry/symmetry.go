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
