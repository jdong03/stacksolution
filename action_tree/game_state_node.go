package action_tree

import "github.com/jdong03/stacksolution/game"

/*
GameStateNode represents a single concrete game state in the game tree.
GameStateNode can be a PlayerNode, ChanceNode, or LeafNode.
*/
type GameStateNode interface {
	GetGameState() GameState
}

/*
GameState is a struct containg shared attributes for all instances of a GameStateNode
*/
type GameState struct {
	History                 History
	Player1Cards            []game.Card
	Player2Cards            []game.Card
	Player1StackSize        float64
	Player2StackSize        float64
	Player1ReachProbability float64
	Player2ReachProbability float64
	PotSize                 float64
}

/*
GetStartingNode creates the starting game state node (PlayerNode)
given the players' hole cards.
*/
func GetStartingNode(player1Cards []game.Card, player2Cards []game.Card, initialPotSize float64) GameStateNode {
	history := NewHistory()
	// Player1 acts first, so use Player1's stack size
	actionOptions := GetActionOptionsFromHistory(history, Player1InitialStackSize, initialPotSize)

	node := &PlayerNode{
		GameState: GameState{
			History:                 *history,
			Player1Cards:            player1Cards,
			Player2Cards:            player2Cards,
			Player1StackSize:        Player1InitialStackSize,
			Player2StackSize:        Player2InitialStackSize,
			Player1ReachProbability: 1.0,
			Player2ReachProbability: 1.0,
			PotSize:                 initialPotSize,
		},
		ActionOptions: actionOptions,
	}
	return node
}

/* NewGameStateNode creates and returns a new GameStateNode with the specified parameters. */
func NewGameStateNode(parentGameStateNode GameStateNode, action Action, actionProbability float64) GameStateNode {
	// Extract the GameState from the parent node
	parentGameState := parentGameStateNode.GetGameState()

	newHistory := AddToHistory(&parentGameState.History, action)
	if newHistory.ActivePlayer == Player1 || newHistory.ActivePlayer == Player2 {
		// Return Player Node
		return NewPlayerNode(parentGameStateNode, action, actionProbability, newHistory)
	} else if newHistory.ActivePlayer == Chance {
		// Return Chance Node
		// ChanceNode expects a PlayerNode as parent and PlayerAction
		parentPlayerNode, ok := parentGameStateNode.(*PlayerNode)
		if !ok {
			panic("Parent of ChanceNode must be a PlayerNode")
		}
		playerAction, ok := action.(PlayerAction)
		if !ok {
			panic("Action for ChanceNode must be a PlayerAction")
		}
		return NewChanceNode(*parentPlayerNode, playerAction, actionProbability, newHistory)
	} else {
		// Return Leaf Node
		// LeafNode expects a PlayerNode as parent and PlayerAction
		parentPlayerNode, ok := parentGameStateNode.(*PlayerNode)
		if !ok {
			panic("Parent of LeafNode must be a PlayerNode")
		}
		playerAction, ok := action.(PlayerAction)
		if !ok {
			panic("Action for LeafNode must be a PlayerAction")
		}
		return NewLeafNode(*parentPlayerNode, playerAction, actionProbability, newHistory)
	}
} /*
if player 1 closing action -> chance
if player 1 non closing -> player 2

clsoing action (call or fold)
*/
