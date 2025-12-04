package action_tree

type LeafNode struct {
	GameState
	winner Player
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

	// Update stack sizes and reach probabilities based on the action taken
	if parentGameStateNode.History.ActivePlayer == Player1 {
		player1StackSize -= action.Amount
		player1ReachProbability *= actionProbability
	} else if parentGameStateNode.History.ActivePlayer == Player2 {
		player2StackSize -= action.Amount
		player2ReachProbability *= actionProbability
	}

	gameState := GameState{
		History:                 *newHistory,
		Player1Cards:            parentGameStateNode.Player1Cards,
		Player2Cards:            parentGameStateNode.Player2Cards,
		Player1StackSize:        player1StackSize,
		Player2StackSize:        player2StackSize,
		Player1ReachProbability: player1ReachProbability,
		Player2ReachProbability: player2ReachProbability,
	}

	winner := DetermineWinner(gameState) // TODO: implement DetermineWinner somewhere

	return &LeafNode{
		GameState: gameState,
		winner:    winner,
	}
}

// TODO: implement DetermineWinner and move to appropriate location
func DetermineWinner(gameState GameState) Player {
	return Player1 // Placeholder implementation
}
