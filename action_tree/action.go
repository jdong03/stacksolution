package action_tree

import "github.com/jdong03/stacksolution/game"

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
