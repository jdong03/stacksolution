package action_tree

import "github.com/jdong03/stacksolution/game"

type LeafNode struct {
	GameState
	Winner Player
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
	}

	winner := determineWinner(gameState, parentGameStateNode.History.ActivePlayer, action)

	return &LeafNode{
		GameState: gameState,
		Winner:    winner,
	}
}

// TODO: How does it work if all in on flop? Worry about later
// determineWinner determines the winner at a leaf node.
// If someone folded, the other player wins.
// If it's a showdown (river ends with call or check-check), compare hands.
func determineWinner(gameState GameState, lastActivePlayer Player, lastAction PlayerAction) Player {
	// Check if someone folded - the player who folded loses
	if lastAction.ActionType == Fold {
		if lastActivePlayer == Player1 {
			return Player2
		}
		return Player1
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
	if result >= 0 {
		return Player1 // P1 wins or tie (split pot)
	}
	return Player2
}
