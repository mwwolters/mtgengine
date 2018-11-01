package engine

import (
	"container/ring"
	"github.com/google/uuid"
	"math/rand"
)

const (
	Untap = iota
	Upkeep = iota
	Draw = iota
	Main = iota
	Combat = iota
	PostCombat = iota
	End = iota
	Cleanup = iota
)

type State struct {
	Players map[string]PlayerState
	PlayerOrder []string
	CurrentPlayer int
	Phase *ring.Ring
	Turn int
}

type PlayerState struct {
	Health int
	Hand []Card
	Deck []Card
	Graveyard []Card
	Exile []Card
}

func Shuffle(cards []Card) []Card {
	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}
	return cards
}

func (s State) Step() int {
	next := s.Phase.Next()
	// TODO Thoughts:
	// * switch for step and have each have a function for transition?
	// * do middle steps explicitly?
	return next.Value.(int)
}

// TODO simple state stepper for doing simulations

type SimpleState struct {
	Played, Hand, Deck []Card
	ID uuid.UUID
	Turn int
}

func NewState(cards []Card) SimpleState {
	deck := Shuffle(cards)
	played := []Card{}
	hand := deck[0:7]
	deck = deck [7:]
	id := uuid.New()
	return SimpleState{played, hand, deck, id, 0}
}

type TurnCalc interface {
	Calc(SimpleState) (SimpleState)
	End()
	Process()
}

// TODO make this more generic for players etc.
func SimulateOnePlayerTurns(turns int, draw bool, state SimpleState, tc TurnCalc){
	for i := 0; i < turns; i++ {
		state = SimulateTurn(draw, state, tc)
	}
	tc.End()
}

func SimulateTurn(draw bool, state SimpleState, tc TurnCalc) SimpleState {
	if len(state.Deck) != 0 {
		state.Hand = append(state.Hand, state.Deck[0])
		state.Deck = state.Deck[1:]
	}
	return tc.Calc(state)
}
