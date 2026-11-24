package world

import "soldier/cmd/agents"

const MaxTurns = 200 //à choisir

type Result struct {
	Finished    bool
	Winner      *agents.Army
	Description string
}

func (w *World) CheckResult() Result {
	switch {
	case len(w.Armies) == 0:
		return Result{Finished: true, Description: "match nul : les deux armées sont éliminées"}
	case len(w.Armies) == 1:
		return Result{Finished: true, Winner: w.Armies[0], Description: "victoire par élimination"}
	case w.TurnCount >= MaxTurns:
		return Result{Finished: true, Description: "match nul : limite de tours atteinte"}
	}
	return Result{}
}

func (w *World) Stop() {
	for _, u := range w.Units {
		close(u.Orders)
	}
}
