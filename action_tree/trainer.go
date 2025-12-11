package action_tree

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jdong03/stacksolution/game"
)

const (
	Player1InitialStackSize = 50.0
	Player2InitialStackSize = 50.0
)

type VanillaCFRTrainer struct {
	Player1InitialStackSize float64
	Player2InitialStackSize float64
	InformationSetMap       map[string]*InformationSet
	InfoSetActionOptions    map[string][]EnumActionType // Maps info set key -> action options
	Iteration               int
	UpdatingPlayer          Player
	BestResponseUtility     *BestResponseUtility
	Metrics                 *TrainingMetrics // Optional metrics tracking for graphing
}

func NewVanillaCFRTrainer() *VanillaCFRTrainer {
	trainer := &VanillaCFRTrainer{
		Player1InitialStackSize: Player1InitialStackSize,
		Player2InitialStackSize: Player2InitialStackSize,
		InformationSetMap:       make(map[string]*InformationSet),
		InfoSetActionOptions:    make(map[string][]EnumActionType),
		Iteration:               0,
		UpdatingPlayer:          Player1,
	}
	// BestResponseUtility needs access to the trainer's infosets / logic to read strategies.
	trainer.BestResponseUtility = NewBestResponseUtility(trainer)
	return trainer
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
		// TODO: Handle different raise sizes appropriately.
		// ActionOptions may have to include the variable raise sizes so as to not to treat different raise sizes as same action.
		// This raises a problem of how we implemented ActionOptions as EnumActionType which doesn't include different raise sizes.
		// I think fixed this one, but read below for another problem
		// TODO: Problem with Chance action options, I'm wondering if we need to extend this method to handle chance nodes as well.
		infoSet = NewInformationSet(len(playerGameStateNode.ActionOptions))
		trainer.InformationSetMap[key] = infoSet
		// Store action options for display purposes
		trainer.InfoSetActionOptions[key] = playerGameStateNode.ActionOptions
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

// Train runs the CFR algorithm for the specified number of iterations.
// It iterates through all possible hand combinations for both players,
// skipping hands that conflict with each other or the board.
//
// Parameters:
//   - numberIterations: number of CFR iterations to run
//   - board: community cards (flop, optionally turn and river)
//   - handCombosP1: all possible hole card combinations for Player 1
//   - handCombosP2: all possible hole card combinations for Player 2
//   - initialPotSize: the starting pot size
//
// Returns the average Player 1 utility from the last iteration.
func (trainer *VanillaCFRTrainer) Train(
	numberIterations int,
	board []game.Card,
	handCombosP1 [][]game.Card,
	handCombosP2 [][]game.Card,
	initialPotSize float64,
) float64 {
	trainer.Iteration = 0
	var p1Util float64
	var utilP1Count int

	for i := 0; i < numberIterations; i++ {
		trainer.Iteration = i
		// Alternate updating player each iteration
		if i%2 == 0 {
			trainer.UpdatingPlayer = Player1
		} else {
			trainer.UpdatingPlayer = Player2
		}

		p1Util = 0
		utilP1Count = 0

		// Iterate through all possible hand combinations
		for _, p1Hand := range handCombosP1 {
			// Skip P1 hands that conflict with the board
			if cardsConflict(p1Hand, board) {
				continue
			}

			for _, p2Hand := range handCombosP2 {
				// Skip P2 hands that conflict with P1 hands
				if cardsConflict(p2Hand, p1Hand) {
					continue
				}

				// Skip P2 hands that conflict with the board
				if cardsConflict(p2Hand, board) {
					continue
				}

				// Create starting node with the board and hands
				startNode := trainer.createStartingNodeWithBoard(board, p1Hand, p2Hand, initialPotSize)

				// Begin the CFR recursion
				p1Util += trainer.CalculateNodeUtility(startNode)
				utilP1Count++
			}
		}

		// Log iteration progress
		fmt.Printf("Iteration %d complete.\n", i)
		exploitability := trainer.BestResponseUtility.TotalDeviation(board, handCombosP1, handCombosP2)
		fmt.Printf("Strategy Exploitability: %.6f\n\n", exploitability)

		// Record metrics if tracking is enabled
		if trainer.Metrics != nil {
			trainer.Metrics.RecordIteration(trainer, exploitability)
		}
	}

	// Return player 1 utility of last iteration
	if utilP1Count == 0 {
		return 0
	}
	return p1Util / float64(utilP1Count)
}

// createStartingNodeWithBoard creates a starting PlayerNode with the board cards already dealt.
func (trainer *VanillaCFRTrainer) createStartingNodeWithBoard(
	board []game.Card,
	p1Hand []game.Card,
	p2Hand []game.Card,
	initialPotSize float64,
) *PlayerNode {
	history := NewHistory()

	// Set up board cards based on how many are provided
	if len(board) >= 3 {
		history.FlopCards = board[0:3]
	}
	if len(board) >= 4 {
		history.TurnCard = board[3:4]
	}
	if len(board) >= 5 {
		history.RiverCard = board[4:5]
	}

	// Player1 acts first on the flop
	history.ActivePlayer = Player1

	actionOptions := GetActionOptionsFromHistory(history, trainer.Player1InitialStackSize, initialPotSize)

	return &PlayerNode{
		GameState: GameState{
			History:                 *history,
			Player1Cards:            p1Hand,
			Player2Cards:            p2Hand,
			Player1StackSize:        trainer.Player1InitialStackSize,
			Player2StackSize:        trainer.Player2InitialStackSize,
			Player1ReachProbability: 1.0,
			Player2ReachProbability: 1.0,
			PotSize:                 initialPotSize,
		},
		ActionOptions: actionOptions,
	}
}

// cardsConflict returns true if any card in set1 appears in set2.
func cardsConflict(set1, set2 []game.Card) bool {
	for _, c1 := range set1 {
		for _, c2 := range set2 {
			if c1.Rank == c2.Rank && c1.Suit == c2.Suit {
				return true
			}
		}
	}
	return false
}

// CalculateNodeUtility recursively calculates the expected utility for Player 1
// at the given game state node using the CFR algorithm.
//
// This is the core CFR recursion that:
//   - At terminal nodes: returns the payoff from P1's perspective
//   - At chance nodes: averages over all possible outcomes
//   - At player nodes: computes strategy, recurses on children, and updates regrets
func (trainer *VanillaCFRTrainer) CalculateNodeUtility(node GameStateNode) float64 {
	switch n := node.(type) {
	case *LeafNode:
		return trainer.calculateLeafUtility(n)

	case *ChanceNode:
		return trainer.calculateChanceUtility(n)

	case *PlayerNode:
		return trainer.calculatePlayerUtility(n)

	default:
		panic("Unknown node type in CalculateNodeUtility")
	}
}

// calculateLeafUtility returns the utility for Player 1 at a terminal node.
// Utility is measured as P1's profit/loss from this hand.
func (trainer *VanillaCFRTrainer) calculateLeafUtility(node *LeafNode) float64 {
	gameState := node.GetGameState()

	// Player 1's utility is their stack size change
	return gameState.Player1StackSize - trainer.Player1InitialStackSize

}

// calculateChanceUtility averages utility over all possible chance outcomes.
func (trainer *VanillaCFRTrainer) calculateChanceUtility(node *ChanceNode) float64 {
	availableCards := node.AvailableCards

	// Each card has equal probability
	actionProbability := 1.0 / float64(len(availableCards))

	var nodeUtility float64

	// Iterate over each possible card that could be dealt
	for _, card := range availableCards {
		// Create a ChanceAction for dealing this card
		chanceAction := ChanceAction{
			RevealedCards: []game.Card{card},
		}

		// Create child node
		childNode := NewGameStateNode(node, chanceAction, actionProbability)

		// Recursively calculate utility
		childUtility := trainer.CalculateNodeUtility(childNode)

		// Average the utilities (weight by probability)
		nodeUtility += actionProbability * childUtility
	}

	return nodeUtility
}

// calculatePlayerUtility computes utility at a player decision node using CFR.
func (trainer *VanillaCFRTrainer) calculatePlayerUtility(node *PlayerNode) float64 {
	gameState := node.GetGameState()
	activePlayer := gameState.History.ActivePlayer

	// Get active player's reach probability
	var activePlayerReachProbability float64
	if activePlayer == Player1 {
		activePlayerReachProbability = gameState.Player1ReachProbability
	} else {
		activePlayerReachProbability = gameState.Player2ReachProbability
	}

	// Get information set and current strategy
	infoSet := trainer.GetInformationSet(node)
	strategy := GetStrategy(infoSet)

	// Add to strategy sum only for the updating player
	if activePlayer == trainer.UpdatingPlayer {
		AddToStrategySum(infoSet, strategy, activePlayerReachProbability)
	}

	// Calculate utility for each action
	actionUtilities := make([]float64, len(node.ActionOptions))

	for i, actionType := range node.ActionOptions {
		actionProbability := strategy[i]

		// Create the PlayerAction for this action type
		action := trainer.createActionForType(actionType, node)

		// Create child node
		childNode := NewGameStateNode(node, action, actionProbability)

		// Recursively calculate utility
		actionUtilities[i] = trainer.CalculateNodeUtility(childNode)
	}

	// Calculate node utility as dot product of action utilities and strategy
	var nodeUtility float64
	for i, utility := range actionUtilities {
		nodeUtility += strategy[i] * utility
	}

	// Add to cumulative regrets only for the updating player
	if activePlayer == trainer.UpdatingPlayer {
		var gsn GameStateNode = node
		AddToCumulativeRegrets(infoSet, &gsn, actionUtilities, nodeUtility)
	}

	return nodeUtility
}

// createActionForType creates a PlayerAction for the given action type.
// The Amount is calculated based on the pot size and action type.
func (trainer *VanillaCFRTrainer) createActionForType(actionType EnumActionType, node *PlayerNode) PlayerAction {
	gameState := node.GetGameState()
	potSize := gameState.PotSize
	currentStreetActions := getCurrentStreetActions(&gameState.History)

	var activeStackSize float64
	if gameState.History.ActivePlayer == Player1 {
		activeStackSize = gameState.Player1StackSize
	} else {
		activeStackSize = gameState.Player2StackSize
	}

	var amount float64
	switch actionType {
	case Check:
		amount = 0
	case Call:
		// Call amount is the last raise amount
		if len(currentStreetActions) > 0 {
			lastAction := currentStreetActions[len(currentStreetActions)-1]
			if isRaiseAction(lastAction.ActionType) {
				amount = lastAction.Amount
			}
		}
	case Raise33:
		amount = potSize * 0.33
	case Raise50:
		amount = potSize * 0.50
	case Raise75:
		amount = potSize * 0.75
	case Raise100:
		amount = potSize * 1.00
	case RaiseAllIn:
		amount = activeStackSize
	case Fold:
		amount = 0
	}

	return PlayerAction{
		ActionType: actionType,
		Amount:     amount,
		Player:     gameState.History.ActivePlayer,
	}
}
