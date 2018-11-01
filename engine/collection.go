package engine

import ()

type Collection interface {
        Cards() []Card
        Name() string
        Id() string
        AddCards([]Card)
        RemoveCards([]Card)
}

type Deck struct {
        Name, Id string
        Cards    []Card
}

func NewDeck(name, id string, cards []Card) Deck {
        return Deck{name, id, cards}
}

func (d Deck) AddCards(cards []Card) {
        d.Cards = append(d.Cards, cards...)
}

func (d Deck) RemoveCards(cards []Card) {
        keep := make(map[int]bool)
        for _, c := range cards {
                for i, dc := range d.Cards {
                        if c.Id == dc.Id {
                                keep[i] = false
                        }
                }
        }
        newDeck := []Card{}
        for i, dc := range d.Cards {
                if _, ok := keep[i]; !ok {
                        newDeck = append(newDeck, dc)
                }
        }
        d.Cards = newDeck
}
