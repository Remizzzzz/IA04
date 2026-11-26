package sim

import (
	"fmt"
	"soldier/cmd/agents"
	"soldier/cmd/world"
	"time"
)

// SimulateArmyWorldWithComposition creates two armies with a fixed composition.
// Each army has one tank (the general), two infantry and one sniper. Armies
// spawn grouped around their general and mirrored by the map's symmetry (see
// World.SpawnArmies), so a symmetric cfg gives both sides the same start.
// TODO: Display live counts per army and unit type. The initial-composition
// label below stays unchanged after casualties; it is not a survivor count.
// TODO: Use distinct symbols for Tank, Infantry and Sniper, keeping team colors.
// TODO: Detect victory when only one army has surviving units.
// TODO: Separate army creation, display and the simulation loop; consider a
// stop channel to terminate the loop cleanly when returning from this function.
// TODO: Add an optional random composition with configurable min/max total (useful for the simulation later on)
// units and min/max counts per type. Validate that these bounds are compatible
// with each other and the available tiles; keep the fixed scenario available.
func SimulateArmyWorldWithComposition(cfg world.MapConfig) {
	w := world.NewWorld(cfg)
	composition := []agents.UnitType{
		agents.Tank,
		agents.Infantry,
		agents.Infantry,
		agents.Sniper,
	}

	w.SpawnArmies(2, composition, agents.Coordinated) // composition[0] (the tank) is each army's general

	go func() {
		for {
			fmt.Print("\033[H\033[2J")
			w.PrintWorld()
			fmt.Println("Initial composition per army: 1 Tank, 2 Infantry, 1 Sniper.")
			fmt.Println("Press Enter to end simulation.")
			w.Turn()
			time.Sleep(time.Second)
		}
	}()

	fmt.Scanln()
}

// SimulateRandomWorld builds a random world and a set of units, then runs the
// simulation loop until the user presses Enter.
func SimulateRandomWorld(cfg world.MapConfig, numberOfEntity int) {
	w := world.NewWorld(cfg)
	for i := 0; i < numberOfEntity; i++ {
		w.SpawnUnitRandom(agents.Random)
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

// SimulateArmyWorld builds a world from cfg with numberOfArmies armies of
// soldiersPerArmy Coordinated infantry each (the first one of each army is its
// general), spawned grouped and mirrored by the map's symmetry (see
// World.SpawnArmies), then runs the simulation loop until the user presses
// Enter - handy for watching the army strategy play out.
func SimulateArmyWorld(numberOfArmies int, soldiersPerArmy int, cfg world.MapConfig) {
	w := world.NewWorld(cfg)
	composition := make([]agents.UnitType, soldiersPerArmy)
	for i := range composition {
		composition[i] = agents.Infantry
	}
	w.SpawnArmies(numberOfArmies, composition, agents.Coordinated)

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
