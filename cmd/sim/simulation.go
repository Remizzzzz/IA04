package sim

import (
	"fmt"
	"soldier/cmd/agents"
	"soldier/cmd/world"
	"time"
)

// SimulateArmyWorldWithComposition creates two armies with a fixed composition.
// Each army has one tank (the general), two infantry and one sniper.
// TODO: Display live counts per army and unit type. The initial-composition
// label below stays unchanged after casualties; it is not a survivor count.
// TODO: Use distinct symbols for Tank, Infantry and Sniper, keeping team colors.
// TODO: Separate army creation, display and the simulation loop; consider a
// stop channel to terminate the loop cleanly when returning from this function.
// TODO: Add an optional random composition with configurable min/max total (useful for the simulation later on)
// units and min/max counts per type. Validate that these bounds are compatible
// with each other and the available tiles; keep the fixed scenario available.
func SimulateArmyWorldWithComposition(sizeOfMap int, obstacleFrequency int) {
	w := world.NewWorld(sizeOfMap, obstacleFrequency)
	composition := []agents.UnitType{
		agents.Tank,
		agents.Infantry,
		agents.Infantry,
		agents.Sniper,
	}

	for allegiance := 1; allegiance <= 2; allegiance++ {
		for _, unitType := range composition {
			w.SpawnAllied(agents.Random, allegiance, unitType)
		}
	}

	done := make(chan world.Result)
	go func() {
		for {
			fmt.Print("\033[H\033[2J")
			w.PrintWorld()
			fmt.Println("Initial composition per army: 1 Tank, 2 Infantry, 1 Sniper.")
			fmt.Printf("Tour %d / %d\n", w.TurnCount, world.MaxTurns)
			fmt.Println("Press Enter to end simulation.")
			w.Turn()
			if result := w.CheckResult(); result.Finished {
				w.Stop()
				done <- result
				return
			}
			time.Sleep(time.Second)
		}
	}()

	// quit permet toujours d'arrêter la simulation avec Entrée.
	quit := make(chan struct{})
	go func() {
		fmt.Scanln()
		close(quit)
	}()

	select {
	case result := <-done:
		fmt.Print("\033[H\033[2J")
		w.PrintWorld()
		printResult(result, w.TurnCount)
	case <-quit:
		fmt.Println("Simulation interrompue.")
	}
}

func printResult(result world.Result, turns int) {
	if result.Winner != nil {
		fmt.Printf("L'armée %d gagne (%s) en %d tours, avec %d unité(s) survivante(s) !\n",
			result.Winner.Allegiance, result.Description, turns, len(result.Winner.Soldiers))
	} else {
		fmt.Printf("Fin de partie après %d tours : %s\n", turns, result.Description)
	}
}

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

// SimulateArmyWorld builds a random world with numberOfArmies armies of
// soldiersPerArmy Coordinated units each (one per army becomes its general),
// then runs the simulation loop until the user presses Enter - handy for
// watching the regroup-then-advance army strategy play out.
func SimulateArmyWorld(sizeOfMap int, numberOfArmies int, soldiersPerArmy int, obstacleFrequency int) {
	w := world.NewWorld(sizeOfMap, obstacleFrequency)
	for allegiance := 1; allegiance <= numberOfArmies; allegiance++ {
		for i := 0; i < soldiersPerArmy; i++ {
			w.SpawnAllied(agents.Coordinated, allegiance, agents.Infantry)
		}
	}

	go func() {
		for {
			fmt.Print("\033[H\033[2J") //Refresh console
			w.PrintWorld()
			for _, army := range w.Armies {
				fmt.Printf("Army (allegiance %d): %d soldiers, phase=%s, general at %v\n",
					army.Allegiance, len(army.Soldiers), army.Phase, army.General.Position)
			}
			fmt.Println("Tap 'Enter' to end simulation")
			w.Turn()
			time.Sleep(time.Duration(1) * time.Second)
		}
	}()

	fmt.Scanln()
}
