package sim

import (
	"fmt"
	"soldier/cmd/agents"
	"soldier/cmd/world"
	"time"
)

// SimulateRandomWorld builds a random world and a set of units, then runs the
// simulation loop until the user presses Enter.
func SimulateRandomWorld(sizeOfMap int, numberOfEntity int, obstacleFrequency int) {
	w := world.NewWorld(sizeOfMap, obstacleFrequency)
	for i := 0; i < numberOfEntity; i++ {
		w.SpawnUnit(agents.Random)
	}

	// Run the simulation loop in the background so the main goroutine can
	// block on fmt.Scanln() and stop everything when the user hits Enter.
	go func() {
		for {
			fmt.Print("\033[H\033[2J") //Refresh console
			w.PrintWorld()
			fmt.Println("Tap 'Enter' to end simulation")
			w.Turn() // sends vision/reachable tiles to every unit and applies their chosen moves
			time.Sleep(time.Duration(1) * time.Second)
		}
	}()

	fmt.Scanln() // block until the user presses Enter, ending the program
}
