package engine

import ()

type Worker struct {
	Cards []Card
	Calc func() TurnCalc
	Turns int
	Draw bool
}

func (w Worker) Start(sims int) {
	for i := 0; i < sims; i++ {
		w.Simulate()
	}
}

func (w Worker) Simulate() {
	state := NewState(w.Cards)
	calc := w.Calc()
	go SimulateOnePlayerTurns(w.Turns, w.Draw, state, calc)
	calc.Process()
}
