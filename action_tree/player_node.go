package action_tree

type PlayerNode struct {
	GameState
	ActionOptions []EnumActionType
}

func (p *PlayerNode) GetGameState() GameState {
	return p.GameState
}

/*
NewPlayerNode creates a new PlayerNode from a parent node (either PlayerNode or ChanceNode)
*/
func NewPlayerNode(parentGameStateNode GameStateNode, action Action, actionProbability float64, newHistory *History) *PlayerNode {
	// Extract parent's game state
	parentGameState := parentGameStateNode.GetGameState()

	// Initialize with parent's values
	player1StackSize := parentGameState.Player1StackSize
	player2StackSize := parentGameState.Player2StackSize
	player1ReachProbability := parentGameState.Player1ReachProbability
	player2ReachProbability := parentGameState.Player2ReachProbability

	// Handle different parent types and action types
	switch parentGameStateNode.(type) {
	case *PlayerNode:
		// Parent is PlayerNode, action should be PlayerAction
		playerAction, ok := action.(PlayerAction)
		if !ok {
			panic("Action from PlayerNode parent must be PlayerAction")
		}

		// Update based on which player took the action
		switch parentGameState.History.ActivePlayer {
		case Player1:
			player1StackSize -= playerAction.Amount
			player1ReachProbability *= actionProbability
		case Player2:
			player2StackSize -= playerAction.Amount
			player2ReachProbability *= actionProbability
		}

	case *ChanceNode:
		// Parent is ChanceNode, action should be ChanceAction
		_, ok := action.(ChanceAction)
		if !ok {
			panic("Action from ChanceNode parent must be ChanceAction")
		}

	// Question: Do we need to handle anything if parent is ChanceNode? Idts as newHistory contains the action already.

	default:
		panic("Parent of PlayerNode must be either PlayerNode or ChanceNode")
	}

	// Create the new GameState
	gameState := GameState{
		History:                 *newHistory,
		Player1Cards:            parentGameState.Player1Cards,
		Player2Cards:            parentGameState.Player2Cards,
		Player1StackSize:        player1StackSize,
		Player2StackSize:        player2StackSize,
		Player1ReachProbability: player1ReachProbability,
		Player2ReachProbability: player2ReachProbability,
	}

	actionOptions := GetActionOptionsFromHistory(newHistory)

	return &PlayerNode{
		GameState:     gameState,
		ActionOptions: actionOptions,
	}
}
