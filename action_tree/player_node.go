package action_tree

type PlayerNode struct {
	GameState
	ActionOptions []EnumActionType
}

func (p *PlayerNode) GetGameState() GameState {
	return p.GameState
}

/*
NewPlayerNode creates a new PlayerNode from a parent node, XXX
*/
func NewPlayerNode(parentGameStateNode *GameStateNode, action Action, actionProbability float64, history *History) *GameStateNode {
	newNode := &GameStateNode{
		History:                 *newHistory,
		Player1Cards:            parentGameStateNode.Player1Cards,
		Player2Cards:            parentGameStateNode.Player2Cards,
		Player1StackSize:        parentGameStateNode.Player1StackSize,
		Player2StackSize:        parentGameStateNode.Player2StackSize,
		Player1ReachProbability: parentGameStateNode.Player1ReachProbability,
		Player2ReachProbability: parentGameStateNode.Player2ReachProbability,
		ActivePlayer:            newHistory.ActivePlayer,
		ActionOptions:           GetActionOptionsFromHistory(newHistory),
	}

	switch action := action.(type) {
	case PlayerAction:
		// Calculate Reach Probabilities, Stack Sizes, Pot
		// If the previous active player was Player 1, update Player 1's reach probability
		switch parentGameStateNode.ActivePlayer {
		case Player1:
			newNode.Player1ReachProbability *= actionProbability
			newNode.Player1StackSize -= action.Amount
			newNode.Pot += action.Amount
		case Player2: // If the previous active player was Player 2, update Player 2's reach probability
			newNode.Player2ReachProbability *= actionProbability
			newNode.Player2StackSize -= action.Amount
			newNode.Pot += action.Amount
		}
	case ChanceAction:
		//TODO
	}

	//TODO: define chance player and chance nodes subsequently

	return newNode
}
