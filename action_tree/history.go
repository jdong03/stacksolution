package action_tree

import (
	"slices"

	"github.com/jdong03/stacksolution/game"
)

/*
History is a history of actions and cards revealed in the game so far
*/
type History struct {
	FlopCards    []game.Card
	FlopActions  []PlayerAction
	TurnCard     []game.Card
	TurnActions  []PlayerAction
	RiverCard    []game.Card
	RiverActions []PlayerAction
	ActivePlayer Player
}

/*
NewHistory creates a new empty history
*/
func NewHistory() *History {
	return &History{
		FlopCards:    nil,
		FlopActions:  nil,
		TurnCard:     nil,
		TurnActions:  nil,
		RiverCard:    nil,
		RiverActions: nil,
		ActivePlayer: Player1, // Player 1 always acts first
	}
}

/*
Clone creates a deep copy of the history
*/
func (h *History) Clone() *History {
	if h == nil {
		return nil
	}
	c := *h // shallow copy of value fields
	c.FlopActions = slices.Clone(h.FlopActions)
	c.TurnActions = slices.Clone(h.TurnActions)
	c.RiverActions = slices.Clone(h.RiverActions)
	return &c
}

/*
AddToHistory returns new history based off current action, doesn't mutate history.

Assumes Player1 is always first to act, and as this is a post-flop solver,
Player1 will always be first to act on each street.
AddToHistory doesn't check for validity of the action whether that is bet sizing or action type
Responsibility of client to pass in valid actions.
*/
func AddToHistory(history *History, action Action) *History {
	var newHistory *History = history.Clone()
	switch typedAction := action.(type) {
	case PlayerAction:
		// If action is a PlayerAction, check what street we are on and update the respective actions in history
		if len(newHistory.RiverCard) > 0 {
			newHistory.RiverActions = append(newHistory.RiverActions, typedAction)
		} else if len(newHistory.TurnCard) > 0 {
			newHistory.TurnActions = append(newHistory.TurnActions, typedAction)
		} else if len(newHistory.FlopCards) > 0 {
			newHistory.FlopActions = append(newHistory.FlopActions, typedAction)
		}

		// If action is Fold or (Call on River) or (Check by Player 2 on River), set active player to Leaf (no more actions)
		if typedAction.ActionType == Fold ||
			(typedAction.ActionType == Call && len(newHistory.RiverCard) > 0) ||
			(typedAction.ActionType == Check && len(newHistory.RiverCard) > 0 && history.ActivePlayer == Player2) {
			newHistory.ActivePlayer = Leaf
		} else if typedAction.ActionType == Call || (typedAction.ActionType == Check && history.ActivePlayer == Player2) {
			// If action is Call or (Check by Player 2), move to next street
			newHistory.ActivePlayer = Chance
		} else {
			// Otherwise, switch active player
			if history.ActivePlayer == Player1 {
				newHistory.ActivePlayer = Player2
			} else {
				newHistory.ActivePlayer = Player1
			}
		}

	case ChanceAction:
		// If action is a ChanceAction, update the respective cards in history
		revealedCards := typedAction.RevealedCards

		if len(newHistory.FlopCards) == 0 {
			newHistory.FlopCards = revealedCards
		} else if len(newHistory.TurnCard) == 0 {
			newHistory.TurnCard = revealedCards
		} else if len(newHistory.RiverCard) == 0 {
			newHistory.RiverCard = revealedCards
		}

		newHistory.ActivePlayer = Player1 // After chance action, Player 1 always acts first
	}

	return newHistory
}

/*
GetActionOptionsFromHistory returns legal action types based on history, stack size, and pot size.
Raise sizes are filtered based on whether the bet amount would exceed the player's stack.
3-bet cap: facing a 3-bet, only options are Call or Fold.
If facing all-in or call amount >= stack, can only Call or Fold.
*/
func GetActionOptionsFromHistory(history *History, stackSize float64, potSize float64) []EnumActionType {
	currentStreetActions := getCurrentStreetActions(history)

	if len(currentStreetActions) == 0 {
		// No actions yet on this street, so options are Check or valid Raises
		validRaises := getValidRaiseSizes(stackSize, potSize)
		options := []EnumActionType{Check}
		options = append(options, validRaises...)
		return options
	}

	// There are actions on this street, check the last action
	lastAction := currentStreetActions[len(currentStreetActions)-1]

	if isRaiseAction(lastAction.ActionType) {
		// If facing all-in, can only Call or Fold
		if lastAction.ActionType == RaiseAllIn {
			return []EnumActionType{Call, Fold}
		}

		// If call amount >= our stack, can only Call (all-in) or Fold
		callAmount := lastAction.Amount
		if callAmount >= stackSize {
			return []EnumActionType{Call, Fold}
		}

		// Count consecutive raises to check for 3-bet cap
		raiseCount := countConsecutiveRaises(currentStreetActions)
		if raiseCount >= 3 {
			return []EnumActionType{Call, Fold}
		}

		// Facing 1-bet or 2-bet with chips behind, so options are Call, valid Raises, or Fold
		// Calculate remaining stack after calling and new pot size for raise calculations
		stackAfterCall := stackSize - callAmount
		potAfterCall := potSize + callAmount

		validRaises := getValidRaiseSizes(stackAfterCall, potAfterCall)
		options := []EnumActionType{Call}
		options = append(options, validRaises...)
		options = append(options, Fold)
		return options
	}

	// Last action was Check, so options are Check or valid Raises
	validRaises := getValidRaiseSizes(stackSize, potSize)
	options := []EnumActionType{Check}
	options = append(options, validRaises...)
	return options
}

// getValidRaiseSizes returns raise sizes that don't exceed the player's stack
func getValidRaiseSizes(stackSize float64, potSize float64) []EnumActionType {
	var validRaises []EnumActionType

	// SIMPLIFIED: Only 50% pot raise for faster solving
	betAmount := potSize * 0.50
	if betAmount <= stackSize {
		validRaises = append(validRaises, Raise50)
	}

	// ORIGINAL CODE - multiple raise sizes:
	// raiseSizes := []struct {
	// 	actionType EnumActionType
	// 	percentage float64
	// }{
	// 	{Raise33, 0.33},
	// 	{Raise50, 0.50},
	// 	{Raise75, 0.75},
	// 	{Raise100, 1.00},
	// }
	//
	// for _, rs := range raiseSizes {
	// 	betAmount := potSize * rs.percentage
	// 	if betAmount <= stackSize {
	// 		validRaises = append(validRaises, rs.actionType)
	// 	}
	// }

	// All-in is always available if player has any chips
	// if stackSize > 0 {
	// 	validRaises = append(validRaises, RaiseAllIn)
	// }

	return validRaises
}

// isRaiseAction returns true if the action type is any raise variant
func isRaiseAction(actionType EnumActionType) bool {
	return actionType == Raise33 || actionType == Raise50 ||
		actionType == Raise75 || actionType == Raise100 ||
		actionType == RaiseAllIn
}

// countConsecutiveRaises counts how many consecutive raises occurred at the end of actions
func countConsecutiveRaises(actions []PlayerAction) int {
	count := 0
	for i := len(actions) - 1; i >= 0; i-- {
		if isRaiseAction(actions[i].ActionType) {
			count++
		} else {
			break
		}
	}
	return count
}

func getCurrentStreetActions(history *History) []PlayerAction {
	if len(history.RiverCard) > 0 {
		return history.RiverActions
	} else if len(history.TurnCard) > 0 {
		return history.TurnActions
	} else if len(history.FlopCards) > 0 {
		return history.FlopActions
	}
	return nil
}
