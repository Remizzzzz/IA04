package main

import (
	"soldier/cmd/sim"
)

func main() {
	sizeOfMap := 10
	numberOfEntity := 2
	obstacleFrequency := 30
	sim.SimulateRandomWorld(sizeOfMap, numberOfEntity, obstacleFrequency)
}
