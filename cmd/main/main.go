package main

import (
	"soldier/cmd/sim"
)

func main() {
	sizeOfMap := 10         // grid is sizeOfMap x sizeOfMap tiles
	obstacleFrequency := 30 // percentage chance for a tile to be an obstacle
	// Previous simulation modes: disable the composition call below before restoring these lines.
	// soldiersPerArmy := 3 // includes the general
	// numberOfArmies := 2
	// random := true
	// if random {
	// 	sim.SimulateRandomWorld(sizeOfMap, 8, obstacleFrequency)
	// } else {
	// 	sim.SimulateArmyWorld(sizeOfMap, numberOfArmies, soldiersPerArmy, obstacleFrequency)
	// }

	sim.SimulateArmyWorldWithComposition(sizeOfMap, obstacleFrequency)
}
