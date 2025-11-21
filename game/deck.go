package game

import (
	"math/rand"
)

// Card represents a playing card with rank and suit.
type Card struct {
	Rank int
	Suit string
}

// Deck represents a standard deck of 52 playing cards.
type Deck struct {
	Cards []Card
}

// NewDeck creates and returns a new shuffled deck of cards.
func NewDeck() Deck {
	suits := []string{"Spades", "Hearts", "Diamonds", "Clubs"}
	ranks := []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11 /* J */, 12 /* Q */, 13 /* K */, 14 /* A */}
	var deck []Card

	for _, rank := range ranks {
		for _, suit := range suits {
			deck = append(deck, Card{rank, suit})
		}
	}
	return Deck{deck}
}

// Shuffle randomizes the order of cards in the deck.
func (d *Deck) Shuffle() {
	rand.Shuffle(len(d.Cards), func(i, j int) {
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	})
}

// DealBoard deals the flop, turn, and river cards from the deck.
func (d *Deck) DealBoard() (flop []Card, turn Card, river Card) {
	flop = d.Cards[:3]
	d.Cards = d.Cards[3:]
	turn = d.Cards[0]
	d.Cards = d.Cards[1:]
	river = d.Cards[0]
	d.Cards = d.Cards[1:]

	return flop, turn, river
}

// Deal deals two hole cards (to a player) from the deck.
func (d *Deck) DealHand() []Card {
	hand := d.Cards[:2]
	d.Cards = d.Cards[2:]
	return hand
}

/*
CardDifference returns a new slice containing the cards in 'deck' that are
not present
*/
func CardDifference(deck []Card, removeLists ...[]Card) []Card {
	// Build a lookup set of cards to remove
	rm := make(map[Card]struct{})
	for _, list := range removeLists {
		for _, c := range list {
			rm[c] = struct{}{}
		}
	}

	// Build output slice
	out := make([]Card, 0, len(deck))
	for _, c := range deck {
		if _, exists := rm[c]; !exists {
			out = append(out, c)
		}
	}
	return out
}
