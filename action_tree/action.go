package action_tree

import (
	"fmt"

	"github.com/jdong03/stacksolution/game"
)

type EnumActionType int

const (
	Check EnumActionType = iota
	Call
	Raise33
	Raise50
	Raise75
	Raise100
	RaiseAllIn
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

// String returns a string representation of the PlayerAction (e.g., "x", "c", "r10.00", "f")
func (a PlayerAction) String() string {
	actionNames := map[EnumActionType]string{
		Check:      "x",
		Call:       "c",
		Raise33:    "r33",
		Raise50:    "r50",
		Raise75:    "r75",
		Raise100:   "r100",
		RaiseAllIn: "rAllIn",
		Fold:       "f",
	}
	if a.ActionType == Raise33 || a.ActionType == Raise50 || a.ActionType == Raise75 || a.ActionType == Raise100 || a.ActionType == RaiseAllIn {
		return fmt.Sprintf("%s%.2f", actionNames[a.ActionType], a.Amount)
	}
	return actionNames[a.ActionType]
}
