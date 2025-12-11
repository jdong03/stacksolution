package action_tree

import (
	"math"

	"github.com/jdong03/stacksolution/game"
)

// BestResponseUtility encapsulates logic for evaluating exploitability of the
// current strategy profile stored in a VanillaCFRTrainer.
//
// Conceptually:
//   - It reads the current strategies (σ₁, σ₂) from trainer.InformationSetMap.
//   - For each player i, it computes the value of a best response BRᵢ(σ_{-i})
//     against the opponent's current strategy.
//   - It compares those best-response values to the current strategy value
//     to get an "exploitability" / deviation from Nash.
//
// A perfect Nash equilibrium strategy profile will have zero deviation
// (no player can improve by deviating).
type BestResponseUtility struct {
	trainer *VanillaCFRTrainer
	// You can add config here later, e.g. pot size normalization, flags for
	// using average strategy vs current strategy, etc.
}

// NewBestResponseUtility constructs a BestResponseUtility instance bound
// to a specific trainer. It will use the trainer's InformationSetMap and
// GetInformationSet function to reconstruct infosets and read strategies.
func NewBestResponseUtility(trainer *VanillaCFRTrainer) *BestResponseUtility {
	return &BestResponseUtility{
		trainer: trainer,
	}
}

// TotalDeviation computes a scalar measure of how exploitable the current
// strategy profile (σ₁, σ₂) is, given a fixed public board and sets of
// private hand combinations for both players.
//
// Parameters:
//   - board:       public board cards (flop/turn/river), in whatever
//     representation you use downstream to construct start nodes.
//   - handCombosP1: all possible hole-card combinations for Player 1.
//   - handCombosP2: all possible hole-card combinations for Player 2.
//
// Interpretation:
//   - It should:
//     1. Compute Vσ = EV for Player 1 when both players play (σ₁, σ₂).
//     2. Compute V_BR1 = EV for Player 1 when P1 plays best response BR₁(σ₂).
//     3. Compute V_BR2 = EV for Player 1 when P2 plays best response BR₂(σ₁).
//     4. Turn those into exploitabilities / a single deviation metric
//     (e.g. as a percentage of pot size).
//
// Return value:
//   - A scalar exploitability / deviation metric. At Nash, this should
//     converge toward 0.
func (bru *BestResponseUtility) TotalDeviation(
	board []game.Card,
	handCombosP1 [][]game.Card,
	handCombosP2 [][]game.Card,
) float64 {
	var (
		valueSigma float64 // Vσ - value under current strategy
		valueBR1   float64 // V_BR1 - value when P1 plays best response
		valueBR2   float64 // V_BR2 - value when P2 plays best response
		count      int
	)

	// Use the trainer's initial pot size (or a default)
	initialPotSize := 50.0 // Default pot size, could be made configurable

	// Iterate over all valid (handP1, handP2) pairs (no card conflicts with board or each other)
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

			// Create starting node for this hand matchup
			startNode := bru.trainer.createStartingNodeWithBoard(board, p1Hand, p2Hand, initialPotSize)

			// 1) Evaluate value under current strategies (σ₁, σ₂)
			valueSigma += bru.valueUnderCurrentStrategy(startNode)

			// 2) Evaluate P1 best response vs σ₂
			valueBR1 += bru.bestResponseValueForPlayer1(startNode)

			// 3) Evaluate P2 best response vs σ₁
			valueBR2 += bru.bestResponseValueForPlayer2(startNode)

			count++
		}
	}

	if count == 0 {
		return 0
	}

	// Average values
	avgVSigma := valueSigma / float64(count)
	avgVBR1 := valueBR1 / float64(count)
	avgVBR2 := valueBR2 / float64(count)

	// Exploitabilities (from P1 perspective, assuming zero-sum):
	//   P1's gain from best response: explP1 = VBR1 - Vσ (how much P1 gains by deviating)
	//   P2's gain from best response: explP2 = Vσ - VBR2 (how much P2 gains, which reduces P1's value)
	explP1 := avgVBR1 - avgVSigma
	explP2 := avgVSigma - avgVBR2

	// Total exploitability is the sum of how much each player can gain by deviating
	// This equals zero at Nash equilibrium
	totalExploitability := explP1 + explP2

	return totalExploitability
}

// valueUnderCurrentStrategy returns the expected value for Player 1 at the given
// node, assuming both players follow the current strategy profile stored in
// trainer.InformationSetMap.
//
// This is analogous to your CFR recursion (CalculateNodeUtility), but without
// updating regrets. Uses the average strategy (GetFinalStrategy) for evaluation.
func (bru *BestResponseUtility) valueUnderCurrentStrategy(node GameStateNode) float64 {
	switch n := node.(type) {
	case *LeafNode:
		return bru.calculateLeafValue(n)

	case *ChanceNode:
		return bru.calculateChanceValue(n, bru.valueUnderCurrentStrategy)

	case *PlayerNode:
		// Get information set and average strategy
		infoSet := bru.trainer.GetInformationSet(n)
		strategy := GetFinalStrategy(infoSet)

		// Calculate expected value as weighted sum over actions
		var nodeValue float64
		for i, actionType := range n.ActionOptions {
			actionProb := strategy[i]
			if actionProb == 0 {
				continue
			}

			action := bru.trainer.createActionForType(actionType, n)
			childNode := NewGameStateNode(n, action, actionProb)
			childValue := bru.valueUnderCurrentStrategy(childNode)
			nodeValue += actionProb * childValue
		}
		return nodeValue

	default:
		panic("Unknown node type in valueUnderCurrentStrategy")
	}
}

// bestResponseValueForPlayer1 returns the expected value for Player 1 at the
// given node if Player 1 plays a best response to Player 2's current strategy σ₂,
// while Player 2 continues to play σ₂.
//
// Behaviour:
//   - At P1 decision nodes: choose the action with maximum child EV (greedy).
//   - At P2 decision nodes: average over actions using σ₂ from infosets.
//   - At chance nodes: average over chance outcomes.
//   - At terminal nodes: return terminal payoff from P1 perspective.
func (bru *BestResponseUtility) bestResponseValueForPlayer1(node GameStateNode) float64 {
	switch n := node.(type) {
	case *LeafNode:
		return bru.calculateLeafValue(n)

	case *ChanceNode:
		return bru.calculateChanceValue(n, bru.bestResponseValueForPlayer1)

	case *PlayerNode:
		gameState := n.GetGameState()
		activePlayer := gameState.History.ActivePlayer

		if activePlayer == Player1 {
			// P1 plays best response: choose action with maximum EV
			maxValue := math.Inf(-1)
			for _, actionType := range n.ActionOptions {
				action := bru.trainer.createActionForType(actionType, n)
				childNode := NewGameStateNode(n, action, 1.0) // probability doesn't matter for BR
				childValue := bru.bestResponseValueForPlayer1(childNode)
				maxValue = math.Max(maxValue, childValue)
			}
			return maxValue
		} else {
			// P2 plays current strategy: weighted average
			infoSet := bru.trainer.GetInformationSet(n)
			strategy := GetFinalStrategy(infoSet)

			var nodeValue float64
			for i, actionType := range n.ActionOptions {
				actionProb := strategy[i]
				if actionProb == 0 {
					continue
				}

				action := bru.trainer.createActionForType(actionType, n)
				childNode := NewGameStateNode(n, action, actionProb)
				childValue := bru.bestResponseValueForPlayer1(childNode)
				nodeValue += actionProb * childValue
			}
			return nodeValue
		}

	default:
		panic("Unknown node type in bestResponseValueForPlayer1")
	}
}

// bestResponseValueForPlayer2 returns the expected value for Player 1 at the
// given node if Player 2 plays a best response to Player 1's current strategy σ₁,
// while Player 1 continues to play σ₁.
//
// Behaviour:
//   - At P2 decision nodes: choose the action with maximum EV *for Player 2*,
//     which corresponds to minimum EV for Player 1 in a zero-sum game.
//   - At P1 decision nodes: average over actions using σ₁ from infosets.
//   - At chance nodes: average over chance outcomes.
//   - At terminal nodes: return terminal payoff from P1 perspective.
func (bru *BestResponseUtility) bestResponseValueForPlayer2(node GameStateNode) float64 {
	switch n := node.(type) {
	case *LeafNode:
		return bru.calculateLeafValue(n)

	case *ChanceNode:
		return bru.calculateChanceValue(n, bru.bestResponseValueForPlayer2)

	case *PlayerNode:
		gameState := n.GetGameState()
		activePlayer := gameState.History.ActivePlayer

		if activePlayer == Player2 {
			// P2 plays best response: choose action that minimizes P1's EV (max for P2)
			minValue := math.Inf(1)
			for _, actionType := range n.ActionOptions {
				action := bru.trainer.createActionForType(actionType, n)
				childNode := NewGameStateNode(n, action, 1.0) // probability doesn't matter for BR
				childValue := bru.bestResponseValueForPlayer2(childNode)
				minValue = math.Min(minValue, childValue)
			}
			return minValue
		} else {
			// P1 plays current strategy: weighted average
			infoSet := bru.trainer.GetInformationSet(n)
			strategy := GetFinalStrategy(infoSet)

			var nodeValue float64
			for i, actionType := range n.ActionOptions {
				actionProb := strategy[i]
				if actionProb == 0 {
					continue
				}

				action := bru.trainer.createActionForType(actionType, n)
				childNode := NewGameStateNode(n, action, actionProb)
				childValue := bru.bestResponseValueForPlayer2(childNode)
				nodeValue += actionProb * childValue
			}
			return nodeValue
		}

	default:
		panic("Unknown node type in bestResponseValueForPlayer2")
	}
}

// calculateLeafValue returns P1's utility at a terminal node.
func (bru *BestResponseUtility) calculateLeafValue(node *LeafNode) float64 {
	gameState := node.GetGameState()

	// Player 1's utility is their stack size change
	p1Utility := gameState.Player1StackSize - bru.trainer.Player1InitialStackSize

	if node.Winner == Player1 {
		return p1Utility
	}
	return -p1Utility
}

// calculateChanceValue averages utility over all possible chance outcomes.
// Takes a recursive evaluation function to allow reuse across different evaluation modes.
func (bru *BestResponseUtility) calculateChanceValue(node *ChanceNode, evalFunc func(GameStateNode) float64) float64 {
	availableCards := node.AvailableCards
	if len(availableCards) == 0 {
		return 0.0
	}

	// Each card has equal probability
	actionProbability := 1.0 / float64(len(availableCards))

	var nodeUtility float64

	// Iterate over each possible card that could be dealt
	for _, card := range availableCards {
		chanceAction := ChanceAction{
			RevealedCards: []game.Card{card},
		}

		childNode := NewGameStateNode(node, chanceAction, actionProbability)
		childUtility := evalFunc(childNode)
		nodeUtility += actionProbability * childUtility
	}

	return nodeUtility
}
