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
}

/*
GetStartingNode creates the starting game state node (PlayerNode)
given the players' hole cards.
*/
func GetStartingNode(player1Cards []game.Card, player2Cards []game.Card) GameStateNode {
	node := &PlayerNode{
		GameState: GameState{
			History:                 *NewHistory(),
			Player1Cards:            player1Cards,
			Player2Cards:            player2Cards,
			Player1StackSize:        Player1InitialStackSize,
			Player2StackSize:        Player2InitialStackSize,
			Player1ReachProbability: 1.0,
			Player2ReachProbability: 1.0,
		},
		ActionOptions: []EnumActionType{Check, Raise, Fold},
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
		return NewChanceNode(parentGameStateNode, action, actionProbability, newHistory)
	} else {
		// Return Leaf Node
		return NewLeafNode(parentGameStateNode, action, actionProbability, newHistory)
	}
} /*
if player 1 closing action -> chance
if player 1 non closing -> player 2

clsoing action (call or fold)
*/
