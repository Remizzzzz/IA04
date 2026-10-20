package sim

import (
	"fmt"
	"soldier/cmd/agents"
	"soldier/cmd/world"
	"time"
)

func SimulateRandomWorld(sizeOfMap int, numberOfEntity int, obstacleFrequency int) {
	w := world.NewWorld(sizeOfMap, obstacleFrequency)
	entities := make([]*agents.Soldier, 0)
	for i := 0; i < numberOfEntity; i++ {
		s := agents.NewSoldier(w, agents.Random)
		entities = append(entities, s)
	}
	go func() {
		w.Simulate()
		for i := range entities {
			entities[i].Simulate()
		}
		for {
			if w.ReadyFlag {
				fmt.Print("\033[H\033[2J") //Refresh console
				w.PrintWorld()
				fmt.Println("Tap 'Enter' to end simulation")
				w.ReadyFlag = false
				for i := range entities {
					entities[i].Command <- "MOVE"
				}
				time.Sleep(time.Duration(1) * time.Second)
			}
		}
	}()

	fmt.Scanln()
}
