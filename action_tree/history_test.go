package action_tree

import (
	"testing"

	"github.com/jdong03/stacksolution/game"
)

// Helper function to create a sample history with flop cards
func createFlopHistory() *History {
	h := NewHistory()
	h.FlopCards = []game.Card{
		{Rank: 14, Suit: "Hearts"},   // Ace of Hearts
		{Rank: 13, Suit: "Spades"},   // King of Spades
		{Rank: 12, Suit: "Diamonds"}, // Queen of Diamonds
	}
	return h
}

// Helper function to create a sample history with turn card
func createTurnHistory() *History {
	h := createFlopHistory()
	h.TurnCard = []game.Card{{Rank: 11, Suit: "Clubs"}} // Jack of Clubs
	return h
}

// Helper function to create a sample history with river card
func createRiverHistory() *History {
	h := createTurnHistory()
	h.RiverCard = []game.Card{{Rank: 10, Suit: "Hearts"}} // 10 of Hearts
	return h
}

func TestAddToHistory_PlayerActions_Flop(t *testing.T) {
	tests := []struct {
		name           string
		setupHistory   func() *History
		action         Action
		expectedPlayer Player
		description    string
	}{
		{
			name: "Check by Player1 switches to Player2",
			setupHistory: func() *History {
				h := createFlopHistory()
				h.ActivePlayer = Player1
				return h
			},
			action: PlayerAction{
				ActionType: Check,
				Amount:     0,
			},
			expectedPlayer: Player2,
			description:    "Player1 checks on flop, should switch to Player2",
		},
		{
			name: "Check-Check on flop goes to Chance (turn)",
			setupHistory: func() *History {
				h := createFlopHistory()
				h.ActivePlayer = Player2
				h.FlopActions = []PlayerAction{{ActionType: Check, Amount: 0}}
				return h
			},
			action: PlayerAction{
				ActionType: Check,
				Amount:     0,
			},
			expectedPlayer: Chance,
			description:    "Both players check on flop, should go to next street (Chance)",
		},
		{
			name: "Raise by Player1 switches to Player2",
			setupHistory: func() *History {
				h := createFlopHistory()
				h.ActivePlayer = Player1
				return h
			},
			action: PlayerAction{
				ActionType: Raise,
				Amount:     100,
			},
			expectedPlayer: Player2,
			description:    "Player1 raises on flop, should switch to Player2",
		},
		{
			name: "Call closes betting round, goes to Chance",
			setupHistory: func() *History {
				h := createFlopHistory()
				h.ActivePlayer = Player2
				h.FlopActions = []PlayerAction{{ActionType: Raise, Amount: 100}}
				return h
			},
			action: PlayerAction{
				ActionType: Call,
				Amount:     100,
			},
			expectedPlayer: Chance,
			description:    "Player2 calls Player1's raise, should go to next street",
		},
		{
			name: "Fold ends the hand",
			setupHistory: func() *History {
				h := createFlopHistory()
				h.ActivePlayer = Player2
				h.FlopActions = []PlayerAction{{ActionType: Raise, Amount: 100}}
				return h
			},
			action: PlayerAction{
				ActionType: Fold,
				Amount:     0,
			},
			expectedPlayer: Leaf,
			description:    "Player2 folds to raise, should end the hand",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			history := tt.setupHistory()
			newHistory := AddToHistory(history, tt.action)

			if newHistory.ActivePlayer != tt.expectedPlayer {
				t.Errorf("%s: expected ActivePlayer %v, got %v",
					tt.description, tt.expectedPlayer, newHistory.ActivePlayer)
			}

			// Verify action was added to correct street
			if len(history.FlopCards) > 0 && len(history.TurnCard) == 0 {
				playerAction := tt.action.(PlayerAction)
				lastAction := newHistory.FlopActions[len(newHistory.FlopActions)-1]
				if lastAction.ActionType != playerAction.ActionType {
					t.Errorf("Action not properly added to FlopActions")
				}
			}
		})
	}
}

func TestAddToHistory_PlayerActions_River(t *testing.T) {
	tests := []struct {
		name           string
		setupHistory   func() *History
		action         Action
		expectedPlayer Player
		description    string
	}{
		{
			name: "Check by Player1 on river switches to Player2",
			setupHistory: func() *History {
				h := createRiverHistory()
				h.ActivePlayer = Player1
				return h
			},
			action: PlayerAction{
				ActionType: Check,
				Amount:     0,
			},
			expectedPlayer: Player2,
			description:    "Player1 checks on river, should switch to Player2",
		},
		{
			name: "Check-Check on river goes to showdown",
			setupHistory: func() *History {
				h := createRiverHistory()
				h.ActivePlayer = Player2
				h.RiverActions = []PlayerAction{{ActionType: Check, Amount: 0}}
				return h
			},
			action: PlayerAction{
				ActionType: Check,
				Amount:     0,
			},
			expectedPlayer: Leaf,
			description:    "Both players check on river, should go to showdown (Leaf)",
		},
		{
			name: "Call on river goes to showdown",
			setupHistory: func() *History {
				h := createRiverHistory()
				h.ActivePlayer = Player2
				h.RiverActions = []PlayerAction{{ActionType: Raise, Amount: 100}}
				return h
			},
			action: PlayerAction{
				ActionType: Call,
				Amount:     100,
			},
			expectedPlayer: Leaf,
			description:    "Player2 calls on river, should go to showdown (Leaf)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			history := tt.setupHistory()
			newHistory := AddToHistory(history, tt.action)

			if newHistory.ActivePlayer != tt.expectedPlayer {
				t.Errorf("%s: expected ActivePlayer %v, got %v",
					tt.description, tt.expectedPlayer, newHistory.ActivePlayer)
			}
		})
	}
}

func TestAddToHistory_ChanceActions(t *testing.T) {
	tests := []struct {
		name             string
		setupHistory     func() *History
		action           Action
		expectedPlayer   Player
		expectedCardSlot string
		description      string
	}{
		{
			name: "Deal flop cards",
			setupHistory: func() *History {
				return NewHistory()
			},
			action: ChanceAction{
				RevealedCards: []game.Card{
					{Rank: 14, Suit: "Hearts"},
					{Rank: 13, Suit: "Spades"},
					{Rank: 12, Suit: "Diamonds"},
				},
			},
			expectedPlayer:   Player1,
			expectedCardSlot: "flop",
			description:      "Dealing flop should set Player1 as active",
		},
		{
			name: "Deal turn card",
			setupHistory: func() *History {
				h := createFlopHistory()
				h.ActivePlayer = Chance
				return h
			},
			action: ChanceAction{
				RevealedCards: []game.Card{{Rank: 11, Suit: "Clubs"}},
			},
			expectedPlayer:   Player1,
			expectedCardSlot: "turn",
			description:      "Dealing turn should set Player1 as active",
		},
		{
			name: "Deal river card",
			setupHistory: func() *History {
				h := createTurnHistory()
				h.ActivePlayer = Chance
				return h
			},
			action: ChanceAction{
				RevealedCards: []game.Card{{Rank: 10, Suit: "Hearts"}},
			},
			expectedPlayer:   Player1,
			expectedCardSlot: "river",
			description:      "Dealing river should set Player1 as active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			history := tt.setupHistory()
			newHistory := AddToHistory(history, tt.action)

			if newHistory.ActivePlayer != tt.expectedPlayer {
				t.Errorf("%s: expected ActivePlayer %v, got %v",
					tt.description, tt.expectedPlayer, newHistory.ActivePlayer)
			}

			// Verify cards were dealt to correct slot
			chanceAction := tt.action.(ChanceAction)
			switch tt.expectedCardSlot {
			case "flop":
				if len(newHistory.FlopCards) != len(chanceAction.RevealedCards) {
					t.Errorf("Flop cards not properly dealt")
				}
			case "turn":
				if len(newHistory.TurnCard) != 1 {
					t.Errorf("Turn card not properly dealt")
				}
			case "river":
				if len(newHistory.RiverCard) != 1 {
					t.Errorf("River card not properly dealt")
				}
			}
		})
	}
}

func TestThreeBetCapScenario(t *testing.T) {
	// Test the complete 3-bet cap scenario with actual game flow
	t.Run("Complete 3-bet cap sequence", func(t *testing.T) {
		// Start with flop dealt
		h := createFlopHistory()

		// Player 1 raises (1-bet)
		h.ActivePlayer = Player1
		options := GetActionOptionsFromHistory(h)
		if len(options) != 2 || options[0] != Check || options[1] != Raise {
			t.Errorf("Player1 initial: expected [Check, Raise], got %v", options)
		}
		h = AddToHistory(h, PlayerAction{ActionType: Raise, Amount: 100})

		// Player 2 can re-raise (2-bet)
		options = GetActionOptionsFromHistory(h)
		if len(options) != 3 || options[0] != Call || options[1] != Raise || options[2] != Fold {
			t.Errorf("Player2 facing 1-bet: expected [Call, Raise, Fold], got %v", options)
		}
		h = AddToHistory(h, PlayerAction{ActionType: Raise, Amount: 300})

		// Player 1 can re-raise again (3-bet)
		options = GetActionOptionsFromHistory(h)
		if len(options) != 3 || options[0] != Call || options[1] != Raise || options[2] != Fold {
			t.Errorf("Player1 facing 2-bet: expected [Call, Raise, Fold], got %v", options)
		}
		h = AddToHistory(h, PlayerAction{ActionType: Raise, Amount: 900})

		// Player 2 CANNOT re-raise (facing 3-bet cap)
		options = GetActionOptionsFromHistory(h)
		if len(options) != 2 || options[0] != Call || options[1] != Fold {
			t.Errorf("Player2 facing 3-bet: expected [Call, Fold] only, got %v", options)
		}

		// Verify no Raise option is available
		for _, opt := range options {
			if opt == Raise {
				t.Error("Raise should not be available when facing 3-bet cap")
			}
		}
	})

	// Test that the cap applies on each street independently
	t.Run("3-bet cap resets on new street", func(t *testing.T) {
		// Setup: 3-bet on flop, then call to go to turn
		h := createFlopHistory()
		h.FlopActions = []PlayerAction{
			{ActionType: Raise, Amount: 100},
			{ActionType: Raise, Amount: 300},
			{ActionType: Raise, Amount: 900},
			{ActionType: Call, Amount: 900}, // Call to close betting
		}

		// Deal turn
		h = AddToHistory(h, ChanceAction{
			RevealedCards: []game.Card{{Rank: 11, Suit: "Clubs"}},
		})

		// On turn, Player1 should be able to check or raise again (cap resets)
		options := GetActionOptionsFromHistory(h)
		if len(options) != 2 || options[0] != Check || options[1] != Raise {
			t.Errorf("New street should reset betting cap: expected [Check, Raise], got %v", options)
		}
	})
}

func TestGetActionOptionsFromHistory(t *testing.T) {
	tests := []struct {
		name            string
		setupHistory    func() *History
		expectedOptions []EnumActionType
		description     string
	}{
		{
			name: "No actions on street - can Check or Raise",
			setupHistory: func() *History {
				h := createFlopHistory()
				h.ActivePlayer = Player1
				return h
			},
			expectedOptions: []EnumActionType{Check, Raise},
			description:     "First to act on a street can Check or Raise",
		},
		{
			name: "After Check - can Check or Raise",
			setupHistory: func() *History {
				h := createFlopHistory()
				h.FlopActions = []PlayerAction{{ActionType: Check, Amount: 0}}
				h.ActivePlayer = Player2
				return h
			},
			expectedOptions: []EnumActionType{Check, Raise},
			description:     "After opponent checks, can Check or Raise",
		},
		{
			name: "Facing 1-bet - can Call, Raise, or Fold",
			setupHistory: func() *History {
				h := createFlopHistory()
				h.FlopActions = []PlayerAction{{ActionType: Raise, Amount: 100}}
				h.ActivePlayer = Player2
				return h
			},
			expectedOptions: []EnumActionType{Call, Raise, Fold},
			description:     "Facing a single raise, can Call, Raise, or Fold",
		},
		{
			name: "Facing 2-bet - can Call, Raise, or Fold",
			setupHistory: func() *History {
				h := createFlopHistory()
				h.FlopActions = []PlayerAction{
					{ActionType: Raise, Amount: 100},
					{ActionType: Raise, Amount: 300},
				}
				h.ActivePlayer = Player1
				return h
			},
			expectedOptions: []EnumActionType{Call, Raise, Fold},
			description:     "Facing a re-raise (2-bet), can still Call, Raise, or Fold",
		},
		{
			name: "Facing 3-bet (cap) - can only Call or Fold",
			setupHistory: func() *History {
				h := createFlopHistory()
				h.FlopActions = []PlayerAction{
					{ActionType: Raise, Amount: 100},
					{ActionType: Raise, Amount: 300},
					{ActionType: Raise, Amount: 900},
				}
				h.ActivePlayer = Player2
				return h
			},
			expectedOptions: []EnumActionType{Call, Fold},
			description:     "Facing 3-bet (cap), can only Call or Fold",
		},
		{
			name: "Turn - No actions yet",
			setupHistory: func() *History {
				h := createTurnHistory()
				h.ActivePlayer = Player1
				return h
			},
			expectedOptions: []EnumActionType{Check, Raise},
			description:     "First to act on turn can Check or Raise",
		},
		{
			name: "River - Facing raise",
			setupHistory: func() *History {
				h := createRiverHistory()
				h.RiverActions = []PlayerAction{{ActionType: Raise, Amount: 100}}
				h.ActivePlayer = Player2
				return h
			},
			expectedOptions: []EnumActionType{Call, Raise, Fold},
			description:     "Facing raise on river, can Call, Raise, or Fold",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			history := tt.setupHistory()
			options := GetActionOptionsFromHistory(history)

			if len(options) != len(tt.expectedOptions) {
				t.Errorf("%s: expected %d options, got %d",
					tt.description, len(tt.expectedOptions), len(options))
				return
			}

			// Check each option is present
			for i, expectedOption := range tt.expectedOptions {
				if options[i] != expectedOption {
					t.Errorf("%s: expected option %v at position %d, got %v",
						tt.description, expectedOption, i, options[i])
				}
			}
		})
	}
}

func TestHistoryClone(t *testing.T) {
	// Test that Clone creates a deep copy
	original := createFlopHistory()
	original.FlopActions = []PlayerAction{
		{ActionType: Check, Amount: 0},
		{ActionType: Raise, Amount: 100},
	}

	cloned := original.Clone()

	// Modify the cloned history
	cloned.FlopActions[0] = PlayerAction{ActionType: Raise, Amount: 200}

	// Verify original is unchanged
	if original.FlopActions[0].ActionType != Check {
		t.Error("Clone did not create a deep copy - original was modified")
	}

	if original.FlopActions[0].Amount != 0 {
		t.Error("Clone did not create a deep copy - original amount was modified")
	}
}

func TestGetCurrentStreetActions(t *testing.T) {
	tests := []struct {
		name            string
		setupHistory    func() *History
		expectedActions int
		description     string
	}{
		{
			name: "River actions returned when on river",
			setupHistory: func() *History {
				h := createRiverHistory()
				h.RiverActions = []PlayerAction{
					{ActionType: Check, Amount: 0},
					{ActionType: Raise, Amount: 100},
				}
				return h
			},
			expectedActions: 2,
			description:     "Should return river actions when river card exists",
		},
		{
			name: "Turn actions returned when on turn",
			setupHistory: func() *History {
				h := createTurnHistory()
				h.TurnActions = []PlayerAction{{ActionType: Raise, Amount: 100}}
				return h
			},
			expectedActions: 1,
			description:     "Should return turn actions when turn card exists but no river",
		},
		{
			name: "Flop actions returned when on flop",
			setupHistory: func() *History {
				h := createFlopHistory()
				h.FlopActions = []PlayerAction{
					{ActionType: Check, Amount: 0},
					{ActionType: Check, Amount: 0},
				}
				return h
			},
			expectedActions: 2,
			description:     "Should return flop actions when only flop cards exist",
		},
		{
			name: "Empty when no community cards",
			setupHistory: func() *History {
				return NewHistory()
			},
			expectedActions: 0,
			description:     "Should return nil when no community cards dealt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			history := tt.setupHistory()
			actions := getCurrentStreetActions(history)

			actualCount := 0
			if actions != nil {
				actualCount = len(actions)
			}

			if actualCount != tt.expectedActions {
				t.Errorf("%s: expected %d actions, got %d",
					tt.description, tt.expectedActions, actualCount)
			}
		})
	}
}
