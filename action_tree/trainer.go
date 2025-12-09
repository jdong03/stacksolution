package action_tree

import (
	"sort"
	"strings"

	"github.com/jdong03/stacksolution/game"
)

const (
	Player1InitialStackSize = 100.0
	Player2InitialStackSize = 100.0
)

type VanillaCFRTrainer struct {
	Player1InitialStackSize float64
	Player2InitialStackSize float64
	InformationSetMap       map[string]*InformationSet
	Iteration               int
	UpdatingPlayer          Player
}

func NewVanillaCFRTrainer() *VanillaCFRTrainer {
	return &VanillaCFRTrainer{
		Player1InitialStackSize: Player1InitialStackSize,
		Player2InitialStackSize: Player2InitialStackSize,
		InformationSetMap:       make(map[string]*InformationSet),
		Iteration:               0,
		UpdatingPlayer:          Player1,
	}
}

func (trainer *VanillaCFRTrainer) GetInformationSet(playerGameStateNode *PlayerNode) *InformationSet {
	gameState := playerGameStateNode.GetGameState()
	activePlayer := gameState.History.ActivePlayer

	// Get active player's hole cards
	var holeCards []game.Card
	switch activePlayer {
	case Player1:
		holeCards = gameState.Player1Cards
	case Player2:
		holeCards = gameState.Player2Cards
	}

	// Build the info set key string
	key := buildInfoSetKey(holeCards, &gameState.History)

	infoSet, exists := trainer.InformationSetMap[key]
	if !exists {
		infoSet = NewInformationSet(len(playerGameStateNode.ActionOptions))
		trainer.InformationSetMap[key] = infoSet
	}
	return infoSet
}

// buildInfoSetKey creates a canonical string key for an information set
// Format: [hole cards sorted]_[flop cards sorted]_[flop actions]_[turn card]_[turn actions]_[river card]_[river actions]
func buildInfoSetKey(holeCards []game.Card, history *History) string {
	var parts []string

	// Add sorted hole cards
	sortedHoleCards := sortCards(holeCards)
	parts = append(parts, sortedHoleCards[0].String(), sortedHoleCards[1].String())

	// Add sorted flop cards
	if len(history.FlopCards) > 0 {
		sortedFlop := sortCards(history.FlopCards)
		for _, card := range sortedFlop {
			parts = append(parts, card.String())
		}
	}

	// Add flop actions
	for _, action := range history.FlopActions {
		parts = append(parts, action.String())
	}

	// Add turn card (only one card, no sorting needed)
	for _, card := range history.TurnCard {
		parts = append(parts, card.String())
	}

	// Add turn actions
	for _, action := range history.TurnActions {
		parts = append(parts, action.String())
	}

	// Add river card (only one card, no sorting needed)
	for _, card := range history.RiverCard {
		parts = append(parts, card.String())
	}

	// Add river actions
	for _, action := range history.RiverActions {
		parts = append(parts, action.String())
	}

	return strings.Join(parts, "_")
}

// sortCards returns a copy of cards sorted in canonical order (high to low by rank, then by suit)
func sortCards(cards []game.Card) []game.Card {
	sorted := make([]game.Card, len(cards))
	copy(sorted, cards)

	suitOrder := map[string]int{
		"Spades":   4,
		"Hearts":   3,
		"Diamonds": 2,
		"Clubs":    1,
	}

	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Rank != sorted[j].Rank {
			return sorted[i].Rank > sorted[j].Rank // Higher rank first
		}
		return suitOrder[sorted[i].Suit] > suitOrder[sorted[j].Suit] // Higher suit first
	})

	return sorted
}
