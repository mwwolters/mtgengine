package engine

import (
	"fmt"
	"strconv"
	"strings"
)

type Set struct {
        Cards   []Card `json:"cards"`
        Name    string `json:"name"`
        Code    string `json:"code"`
        Release string `json:"magicCardsInfoCode"`
        Border  string `json:"border"`
}

type Card struct {
        Artist     string   `json:"artist"`
        Cmc        int      `json:"cmc"`
        Colors     []string `json:"colorIdentity"`
        Flavor     string   `json:"flavor"`
        Id         string   `json:"id"`
        Name       string   `json:"name"`
        Power      string   `json:"power"`
        Toughness  string   `json:"toughness"`
        Rarity     string   `json:"artist"`
        ManaCost   string   `json:"manaCost"`
        Subtypes   []string `json:"subtypes"`
        Supertypes []string `json:"supertypes"`
        Text       string   `json:"text"`
        CType      string   `json:"type"`
        Types      []string `json:"types"`
}

type Mana struct {
	Black int
	Blue int
	Generic int
	Green int
	Red int
	White int
}

func (c Card) Pow() int {
	p, _ := strconv.Atoi(c.Power)
	return p
}

func (c Card) IsLand() bool {
	for _, t := range c.Types {
		if t == "Land" {
			return true
		}
	}
	return false
}

// TODO handle colorless mana cost
func (c Card) GivesMana() Mana {
	switch c.Name {
	case "Mountain":
		return Mana{Red: 1}
	case "Forest":
		return Mana{Green: 1}
	case "Plains":
		return Mana{White: 1}
	case "Swamp":
		return Mana{Black: 1}
	case "Island":
		return Mana{Blue: 1}
	default:
		// TODO handle special mana
		return Mana{}
	}
}

func (c Card) Cost() Mana {
	x := strings.Replace(c.ManaCost, "{", "", -1)
	y := strings.Split(x, "}")
	cost := Mana{}
	for _, tok := range y {
		switch tok {
		case "B":
			cost.Black += 1
		case "U":
			cost.Blue += 1
		case "G":
			cost.Green += 1
		case "R":
			cost.Red += 1
		case "W":
			cost.White += 1
		default:
			c, _ := strconv.Atoi(tok)
			// TODO log err or something
			cost.Generic += c
		}
	}
	return cost
}

func (m Mana) Total() int {
	return m.Black + m.Blue + m.Generic + m.Green + m.Red + m.White
}

func (m Mana) Add(other Mana) Mana {
	return Mana{
		m.Black + other.Black,
		m.Blue + other.Blue,
		m.Generic + other.Generic,
		m.Green + other.Green,
		m.Red + other.Red,
		m.White + other.White,
	}
}

func (m Mana) Sub(other Mana) Mana {
	return Mana{
		m.Black - other.Black,
		m.Blue - other.Blue,
		m.Generic - other.Generic,
		m.Green - other.Green,
		m.Red - other.Red,
		m.White - other.White,
	}
}

// TODO clean up
func (m Mana) GT(other Mana) bool {
	total := m.Total()
	if m.Black < other.Black {
		return false	
	}
	total -= other.Black

	if m.Blue < other.Blue {
		return false	
	}
	total -= other.Blue

	if m.Generic < other.Generic {
		return false	
	}
	total -= other.Generic

	if m.Green < other.Green {
		return false	
	}
	total -= other.Green

	if m.Red < other.Red {
		return false	
	}
	total -= other.Red

	if m.White < other.White {
		return false	
	}
	total -= other.White
	return total > 0
}

func (m Mana) GTE(other Mana) bool {
	total := m.Total()
	if m.Black < other.Black {
		return false	
	}
	total -= other.Black

	if m.Blue < other.Blue {
		return false	
	}
	total -= other.Blue

	if m.Generic < other.Generic {
		return false	
	}
	total -= other.Generic

	if m.Green < other.Green {
		return false	
	}
	total -= other.Green

	if m.Red < other.Red {
		return false	
	}
	total -= other.Red

	if m.White < other.White {
		return false	
	}
	total -= other.White
	return total >= 0
}

func PrintCards(cards []Card) {
	for _, c := range cards {
		fmt.Printf("%v\n", c.Name)
	}
}
