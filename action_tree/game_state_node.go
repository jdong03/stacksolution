package action_tree

import "github.com/jdong03/stacksolution/game"

// TODO: Move to game package when possible

type EnumActionType int

const (
	Check EnumActionType = iota
	Call
	Raise
	Fold
)

/*
*
Action Amount isn't a fraction ranging from [0, 1], instead it's the bet amount in BB up to two decimal points.
Will have to check that any amount is not more than player stack size (maybe not here but the client of this)
*/
type Action interface{}

type PlayerAction struct {
	ActionType EnumActionType
	Amount     float64
	Player     Player
}

type ChanceAction struct {
	RevealedCards []game.Card
}

// END TODO: Move to game package when possible

/* GameStateNode represents a single concrete game state in the game tree. */
type GameStateNode struct {
	History                 History
	Player1Cards            []game.Card
	Player2Cards            []game.Card
	Player1StackSize        float64
	Player2StackSize        float64
	Pot                     float64
	Player1ReachProbability float64
	Player2ReachProbability float64
	ActivePlayer            Player
	ActionOptions           []EnumActionType
}

/* GetStartingNode creates the starting game state node given the players' hole cards. */
func GetStartingNode(player1Cards []game.Card, player2Cards []game.Card) *GameStateNode {
	node := &GameStateNode{
		History:                 *NewHistory(),
		Player1Cards:            player1Cards,
		Player2Cards:            player2Cards,
		Player1StackSize:        100.0,
		Player2StackSize:        100.0,
		Pot:                     0.0,
		Player1ReachProbability: 1.0,
		Player2ReachProbability: 1.0,
		ActivePlayer:            Player1, // Player 1 always acts first
		ActionOptions:           []EnumActionType{Check, Raise, Fold},
	}
	return node
}

/* NewGameStateNode creates and returns a new GameStateNode with the specified parameters. */
func NewGameStateNode(parentGameStateNode *GameStateNode, action Action, actionProbability float64) *GameStateNode {
	newHistory := AddToHistory(&parentGameStateNode.History, action)
	newNode := &GameStateNode{
		History:                 *newHistory,
		Player1Cards:            parentGameStateNode.Player1Cards,
		Player2Cards:            parentGameStateNode.Player2Cards,
		Player1StackSize:        parentGameStateNode.Player1StackSize,
		Player2StackSize:        parentGameStateNode.Player2StackSize,
		Pot:                     parentGameStateNode.Pot,
		Player1ReachProbability: parentGameStateNode.Player1ReachProbability,
		Player2ReachProbability: parentGameStateNode.Player2ReachProbability,
		ActivePlayer:            newHistory.ActivePlayer,
		ActionOptions:           GetActionOptionsFromHistory(newHistory),
	}

	switch action := action.(type) {
	case PlayerAction:
		// Calculate Reach Probabilities, Stack Sizes, Pot
		// If the previous active player was Player 1, update Player 1's reach probability
		if parentGameStateNode.ActivePlayer == Player1 {
			newNode.Player1ReachProbability *= actionProbability
			newNode.Player1StackSize -= action.Amount
			newNode.Pot += action.Amount
		} else if parentGameStateNode.ActivePlayer == Player2 { // If the previous active player was Player 2, update Player 2's reach probability
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

/*
if player 1 closing action -> chance
if player 1 non closing -> player 2

clsoing action (call or fold)
*/
