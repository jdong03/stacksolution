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

type Action struct {
	ActionType EnumActionType
	Amount     int
	Player     Player
}

// END TODO: Move to game package when possible

/* GameStateNode represents a single concrete game state in the game tree. */
type GameStateNode struct {
	History                 History
	Player1Cards            []game.Card
	Player2Cards            []game.Card
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
		Player1ReachProbability: 1.0,
		Player2ReachProbability: 1.0,
		ActivePlayer:            Player1, // Player 1 always acts first
		ActionOptions:           []EnumActionType{Check, Raise, Fold},
	}
	return node
}
