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

// String returns a string representation of the PlayerAction (e.g., "x", "c", "r10.00|33", "f")
func (a PlayerAction) String() string {
	actionNames := map[EnumActionType]string{
		Check: "x",
		Call:  "c",
		Fold:  "f",
	}

	// For raises, include amount and percentage type separated by pipe
	if a.ActionType == Raise33 {
		return fmt.Sprintf("r%.2f|33", a.Amount)
	}
	if a.ActionType == Raise50 {
		return fmt.Sprintf("r%.2f|50", a.Amount)
	}
	if a.ActionType == Raise75 {
		return fmt.Sprintf("r%.2f|75", a.Amount)
	}
	if a.ActionType == Raise100 {
		return fmt.Sprintf("r%.2f|100", a.Amount)
	}

	// For all-in, use special formatting
	if a.ActionType == RaiseAllIn {
		return fmt.Sprintf("rAllIn%.2f", a.Amount)
	}

	return actionNames[a.ActionType]
}
