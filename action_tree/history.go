package action_tree

import (
	"slices"

	"github.com/jdong03/stacksolution/game"
)

/* History is a history of actions and cards revealed in the game so far */
type History struct {
	FlopCards    []game.Card
	FlopActions  []Action
	TurnCard     game.Card
	TurnActions  []Action
	RiverCard    game.Card
	RiverActions []Action
	ActivePlayer Player
}

/* NewHistory creates a new empty history */
func NewHistory() *History {
	return &History{
		FlopCards:    nil,
		FlopActions:  nil,
		TurnCard:     game.Card{},
		TurnActions:  nil,
		RiverCard:    game.Card{},
		RiverActions: nil,
		ActivePlayer: nil,
	}
}

/* Clone creates a deep copy of the history */
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
