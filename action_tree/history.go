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
GetActionOptionsFromHistory returns legal action types based on history
So far, 3-bet will be the biggest bet allowed on any street. Thus, facing a 3-bet, only options are Call or Fold.
*/
func GetActionOptionsFromHistory(history *History) []EnumActionType {
	currentStreetActions := getCurrentStreetActions(history)

	if len(currentStreetActions) == 0 {
		// No actions yet on this street, so options are Check or Raise
		return []EnumActionType{Check, Raise}
	} else {
		// There are actions on this street, check the last action
		lastAction := currentStreetActions[len(currentStreetActions)-1]
		if lastAction.ActionType == Raise {
			if len(currentStreetActions) >= 3 &&
				currentStreetActions[len(currentStreetActions)-2].ActionType == Raise &&
				currentStreetActions[len(currentStreetActions)-3].ActionType == Raise {
				// Facing 3-bet, so options are Call or Fold
				return []EnumActionType{Call, Fold}
			} else {
				// Facing 1-bet or 2-bet, so options are Call, Raise, or Fold
				return []EnumActionType{Call, Raise, Fold}
			}
		} else {
			// Last action was Check, so options are Check or Raise
			return []EnumActionType{Check, Raise}
		}
	}
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
