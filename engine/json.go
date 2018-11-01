package engine

import (
	"bufio"
        "encoding/json"
        "fmt"
        "io/ioutil"
        "os"
	"strings"
	"strconv"
)

func LoadSets(file string) (map[string]Set, error) {
        raw, err := ioutil.ReadFile(file)
        if err != nil {
                fmt.Printf("Couldn't read: %v\n", err)
		return nil, err
        }

        var set map[string]Set
        json.Unmarshal(raw, &set)
        return set, nil
}

func FlattenSet(sets map[string]Set) map[string]*Card {
	cards := make(map[string]*Card)
	for _, v := range sets {
		for _, c := range v.Cards {
			if _, ok := cards[c.Name]; !ok {
				card := Card{
					Cmc: c.Cmc,
					Colors: c.Colors,
					Name: c.Name,
					Power: c.Power,
					Toughness: c.Toughness,
					ManaCost: c.ManaCost,
					Text: c.Text,
					Types: c.Types,
					Subtypes: c.Subtypes,
				}
				cards[strings.ToUpper(card.Name)] = &card
			}
		}
	}
	return cards
}

func LoadCards(file string) []Card {
        raw, err := ioutil.ReadFile(file)
        if err != nil {
                fmt.Printf("Couldn't read: %v\n", err)
                os.Exit(1)
        }

        var cards []Card
        json.Unmarshal(raw, &cards)
        return cards
}

func LoadTextCards(path string, cards map[string]Card) []Card {
	in, _ := os.Open(path)
	defer in.Close()
	scanner := bufio.NewScanner(in)
	scanner.Split(bufio.ScanLines)
	deck := []Card{}
	for scanner.Scan() {
		line := scanner.Text()
		spl := strings.SplitN(line, " ", 2)
		if len(spl) < 2 {
			return deck
		}
		count, _ := strconv.Atoi(spl[0])
		// TODO error handle
		name := strings.ToUpper(spl[1])
		card, ok := cards[name]
		if ok {
			for i := 0; i < count; i++ {
				deck = append(deck, card)
			}
		}
	}
	return deck
}

