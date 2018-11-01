package engine

import ()

type ManaStats struct {
	CMC int
	Black int
	Blue int
	Green int
	Red int
	White int
}

type TypeStats struct {
	Creatures int
	Instant int
	Sorcery int
	Enchantment int
	Lands int	
}

type CMCStats struct {
	CMC map[int]int
	Min int
	Max int
}

func DeckCMCStats(deck Deck) CMCStats {
	min := 20
	max := -1
	cmc := make(map[int]int)
	for _, card := range deck.Cards {
		c := card.Cmc
		cmc[c] += 1
		if c < min {
			min = c
		}
		if c > max {
			max = c
		}
	}
	return CMCStats{cmc, min, max}
}
