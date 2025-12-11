package game

import (
	"fmt"
	"math/rand"
)

// Card represents a playing card with rank and suit.
type Card struct {
	Rank int
	Suit string
}

// String returns a canonical string representation of the card (e.g., "14s" for Ace of Spades)
func (c Card) String() string {
	suitChar := map[string]string{
		"Spades":   "s",
		"Hearts":   "h",
		"Diamonds": "d",
		"Clubs":    "c",
	}
	return fmt.Sprintf("%d%s", c.Rank, suitChar[c.Suit])
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

// GetHandCombinations returns all possible 2-card combinations for a given hand notation.
// Examples:
//   - "AA" -> all pairs of Aces (6 combinations)
//   - "KK" -> all pairs of Kings (6 combinations)
//   - "AKs" -> all Ace-King suited combinations (4 combinations)
//   - "AKo" -> all Ace-King offsuit combinations (12 combinations)
//   - "AK" -> all Ace-King combinations, suited and offsuit (16 combinations)
func GetHandCombinations(handNotation string) [][]Card {
	var combos [][]Card

	// Parse the hand notation
	if len(handNotation) < 2 {
		return combos
	}

	// Parse first rank
	rank1 := parseRank(handNotation[0:1])
	// Parse second rank
	rank2 := parseRank(handNotation[1:2])

	// Check if suited/offsuit is specified
	var suited, offsuit bool
	if len(handNotation) >= 3 {
		if handNotation[2] == 's' {
			suited = true
		} else if handNotation[2] == 'o' {
			offsuit = true
		}
	} else {
		// No specification means both suited and offsuit
		suited = true
		offsuit = true
	}

	suits := []string{"Spades", "Hearts", "Diamonds", "Clubs"}

	// Generate all combinations
	for i, suit1 := range suits {
		for j, suit2 := range suits {
			// Skip invalid combinations
			if rank1 == rank2 && i >= j {
				// For pairs, only generate each unique combination once
				continue
			}
			if rank1 != rank2 && i == j && !suited {
				// Skip suited combinations if only offsuit requested
				continue
			}
			if rank1 != rank2 && i != j && !offsuit {
				// Skip offsuit combinations if only suited requested
				continue
			}

			card1 := Card{Rank: rank1, Suit: suit1}
			card2 := Card{Rank: rank2, Suit: suit2}

			// For pairs, make sure we don't duplicate (e.g., AsAh and AhAs are the same)
			if rank1 == rank2 {
				combos = append(combos, []Card{card1, card2})
			} else {
				combos = append(combos, []Card{card1, card2})
			}
		}
	}

	return combos
}

// parseRank converts a single character rank notation to an integer
// T=10, J=11, Q=12, K=13, A=14
func parseRank(rankStr string) int {
	switch rankStr {
	case "A", "a":
		return 14
	case "K", "k":
		return 13
	case "Q", "q":
		return 12
	case "J", "j":
		return 11
	case "T", "t":
		return 10
	default:
		// Try to parse as number (2-9)
		if len(rankStr) == 1 && rankStr[0] >= '2' && rankStr[0] <= '9' {
			return int(rankStr[0] - '0')
		}
		return 0
	}
}

// ParseBoard converts a board string like "2d, 2s, 2c, 2h, 3d" or "2d 2s 2c 2h 3d"
// into a slice of Cards
func ParseBoard(boardStr string) []Card {
	var cards []Card

	// Split by comma or space
	var cardStrs []string
	currentCard := ""

	for i := 0; i < len(boardStr); i++ {
		char := boardStr[i]
		if char == ',' || char == ' ' {
			if len(currentCard) > 0 {
				cardStrs = append(cardStrs, currentCard)
				currentCard = ""
			}
		} else {
			currentCard += string(char)
		}
	}
	if len(currentCard) > 0 {
		cardStrs = append(cardStrs, currentCard)
	}

	// Parse each card string (e.g., "2d" -> rank=2, suit=Diamonds)
	for _, cardStr := range cardStrs {
		if len(cardStr) < 2 {
			continue
		}

		rankStr := cardStr[0:1]
		suitStr := cardStr[1:2]

		rank := parseRank(rankStr)
		suit := parseSuit(suitStr)

		if rank > 0 && suit != "" {
			cards = append(cards, Card{Rank: rank, Suit: suit})
		}
	}

	return cards
}

// parseSuit converts a single character suit notation to full suit name
func parseSuit(suitStr string) string {
	switch suitStr {
	case "s", "♠":
		return "Spades"
	case "h", "♥":
		return "Hearts"
	case "d", "♦":
		return "Diamonds"
	case "c", "♣":
		return "Clubs"
	default:
		return ""
	}
}
