package game

import (
	"math/rand"
	"time"
)

type Card struct {
	Rank int
	Suit string
}

type Deck struct {
	Cards []Card
}

type Player struct {
	Hand  []Card
	Stack int
}

func (p *Player) addStack(amount int) {
	p.Stack += amount
}

func (p *Player) removeStack(amount int) {
	p.Stack -= amount
}

func createDeck() Deck {
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

func (d *Deck) Shuffle() {
	rand.Seed(time.Now().UnixNano())

	rand.Shuffle(len(d.Cards), func(i, j int) {
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	})
}

func (d *Deck) dealBoard() (flop []Card, turn Card, river Card) {
	flop = d.Cards[:3]
	d.Cards = d.Cards[3:]
	turn = d.Cards[0]
	d.Cards = d.Cards[1:]
	river = d.Cards[0]
	d.Cards = d.Cards[1:]

	return flop, turn, river
}

func (d *Deck) Deal() []Card {
	hand := d.Cards[:2]
	d.Cards = d.Cards[2:]
	return hand
}

func playHand() {

}
