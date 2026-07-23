package action_tree

import "github.com/jdong03/stacksolution/game"

type LeafNode struct {
	GameState
}

func (p *LeafNode) GetGameState() GameState {
	return p.GameState
}

/*
NewLeafNode creates a new LeafNode from a parent node, action, and action probability
*/
func NewLeafNode(parentGameStateNode PlayerNode, action PlayerAction, actionProbability float64, newHistory *History) *LeafNode {
	player1StackSize := parentGameStateNode.Player1StackSize
	player2StackSize := parentGameStateNode.Player2StackSize
	player1ReachProbability := parentGameStateNode.Player1ReachProbability
	player2ReachProbability := parentGameStateNode.Player2ReachProbability
	potSize := parentGameStateNode.PotSize

	// Update stack sizes and reach probabilities based on the action taken
	switch parentGameStateNode.History.ActivePlayer {
	case Player1:
		player1StackSize -= action.Amount
		player1ReachProbability *= actionProbability
		potSize += action.Amount

	case Player2:
		player2StackSize -= action.Amount
		player2ReachProbability *= actionProbability
		potSize += action.Amount
	}

	gameState := GameState{
		History:                 *newHistory,
		Player1Cards:            parentGameStateNode.Player1Cards,
		Player2Cards:            parentGameStateNode.Player2Cards,
		Player1StackSize:        player1StackSize,
		Player2StackSize:        player2StackSize,
		Player1ReachProbability: player1ReachProbability,
		Player2ReachProbability: player2ReachProbability,
		PotSize:                 potSize,
		InitialPotSize:          parentGameStateNode.InitialPotSize,
	}

	hand_winner := determineWinner(gameState, parentGameStateNode.History.ActivePlayer, action)
	switch hand_winner {
	case 1:
		gameState.Player1StackSize += gameState.PotSize
	case -1:
		gameState.Player2StackSize += gameState.PotSize
	case 0:
		// If it's a tie, both players keep their stacks as is
		gameState.Player1StackSize += gameState.PotSize / 2
		gameState.Player2StackSize += gameState.PotSize / 2
	}

	return &LeafNode{
		GameState: gameState,
	}
}

// leafUtility returns Player 1's net profit/loss for the hand at a terminal
// node's game state: the change in P1's stack size, netted against P1's own
// share of the pot that already existed before the tree started (both
// players are assumed to have funded InitialPotSize equally). This is the
// single source of truth for terminal-value calculations - previously this
// formula was independently duplicated (and had already drifted) across
// trainer.go's calculateLeafUtility, best_response_utility.go's
// calculateLeafValue, and heuristic_eval.go's evalNode leaf case.
func leafUtility(gameState GameState, player1InitialStackSize float64) float64 {
	return gameState.Player1StackSize - (player1InitialStackSize + gameState.InitialPotSize/2)
}

// TODO: How does it work if all in on flop? Worry about later
// determineWinner determines the winner at a leaf node.
// If someone folded, the other player wins.
// If it's a showdown (river ends with call or check-check), compare hands.
func determineWinner(gameState GameState, lastActivePlayer Player, lastAction PlayerAction) int {
	// Check if someone folded - the player who folded loses
	if lastAction.ActionType == Fold {
		if lastActivePlayer == Player1 {
			return -1
		}
		return 1
	}

	// It's a showdown - compare hands using game.CompareHands
	history := gameState.History

	// Get turn and river cards (they're slices, need single Card)
	var turnCard, riverCard game.Card
	if len(history.TurnCard) > 0 {
		turnCard = history.TurnCard[0]
	}
	if len(history.RiverCard) > 0 {
		riverCard = history.RiverCard[0]
	}

	result := game.CompareHands(
		gameState.Player1Cards,
		gameState.Player2Cards,
		history.FlopCards,
		turnCard,
		riverCard,
	)

	// result: 1 = P1 wins, -1 = P2 wins, 0 = tie
	return result
}
