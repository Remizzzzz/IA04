package main

import (
	"soldier/cmd/sim"
)

func main() {
	sizeOfMap := 10         // grid is sizeOfMap x sizeOfMap tiles
	numberOfEntity := 5     // number of units spawned in the world
	obstacleFrequency := 30 // percentage chance for a tile to be an obstacle
	sim.SimulateRandomWorld(sizeOfMap, numberOfEntity, obstacleFrequency)
}
