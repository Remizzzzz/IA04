package main

import (
	"soldier/cmd/geometry"
	"soldier/cmd/sim"
	"soldier/cmd/world"
)

func main() {
	sizeOfMap := 10                       // grid is sizeOfMap x sizeOfMap tiles
	obstacleFrequency := 30               // percentage chance for a tile to be an obstacle
	symmetry := geometry.VerticalSymmetry //No symmetry in the map
	cfg := world.MapConfig{Size: sizeOfMap, ObstacleFrequency: obstacleFrequency, Symmetry: symmetry}
	// Previous simulation modes: disable the composition call below before restoring these lines.
	// soldiersPerArmy := 3 // includes the general
	// numberOfArmies := 2
	// random := true
	// if random {
	// 	sim.SimulateRandomWorld(cfg, 8)
	// } else {
	// 	sim.SimulateArmyWorld(numberOfArmies, soldiersPerArmy, cfg)
	// }

	sim.SimulateArmyWorldWithComposition(cfg)
}
