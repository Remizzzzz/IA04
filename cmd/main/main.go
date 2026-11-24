package main

import (
	"soldier/cmd/sim"
)

func main() {
	sizeOfMap := 10         // grid is sizeOfMap x sizeOfMap tiles
	numberOfArmies := 2     // number of armies to spawn
	soldiersPerArmy := 3    // number of soldiers (including the general) per army
	obstacleFrequency := 30 // percentage chance for a tile to be an obstacle
	random := true
	if random {
		sim.SimulateRandomWorld(sizeOfMap, 8, obstacleFrequency)
	} else {
		sim.SimulateArmyWorld(sizeOfMap, numberOfArmies, soldiersPerArmy, obstacleFrequency)
	}
}
