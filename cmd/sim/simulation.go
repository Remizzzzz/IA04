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
	entities := make([]*agents.Unit, 0)
	for i := 0; i < numberOfEntity; i++ {
		s := agents.NewUnit(w, agents.Random)
		entities = append(entities, s)
	}
	// Run the simulation loop in the background so the main goroutine can
	// block on fmt.Scanln() and stop everything when the user hits Enter.
	go func() {
		w.Simulate()
		for i := range entities {
			entities[i].Simulate()
		}
		for {
			// w.ReadyFlag becomes true once every unit has reported back for
			// the previous turn (see World.Simulate), signalling a new turn can start.
			if w.ReadyFlag {
				fmt.Print("\033[H\033[2J") //Refresh console
				w.PrintWorld()
				fmt.Println("Tap 'Enter' to end simulation")
				w.ReadyFlag = false
				for i := range entities {
					entities[i].Command <- "MOVE" // tell every unit to take its turn
				}
				time.Sleep(time.Duration(1) * time.Second)
			}
		}
	}()

	fmt.Scanln() // block until the user presses Enter, ending the program
}
