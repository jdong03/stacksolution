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
	var stackSize float64

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
			stackSize = player1StackSize
			parentGameState.PotSize += playerAction.Amount
		case Player2:
			player2StackSize -= playerAction.Amount
			player2ReachProbability *= actionProbability
			stackSize = player2StackSize
			parentGameState.PotSize += playerAction.Amount
		}

	case *ChanceNode:
		// Parent is ChanceNode, action should be ChanceAction
		_, ok := action.(ChanceAction)
		if !ok {
			panic("Action from ChanceNode parent must be ChanceAction")
		}
		// After chance node, Player1 acts first, so use Player1's stack size
		stackSize = player1StackSize

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
		PotSize:                 parentGameState.PotSize,
	}

	actionOptions := GetActionOptionsFromHistory(newHistory, stackSize, gameState.PotSize)

	return &PlayerNode{
		GameState:     gameState,
		ActionOptions: actionOptions,
	}
}
