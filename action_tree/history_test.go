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
				ActionType: Raise50,
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
				h.FlopActions = []PlayerAction{{ActionType: Raise50, Amount: 100}}
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
				h.FlopActions = []PlayerAction{{ActionType: Raise50, Amount: 100}}
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
				h.RiverActions = []PlayerAction{{ActionType: Raise50, Amount: 100}}
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
	// Use reasonable stack and pot sizes for testing
	stackSize := 1000.0
	potSize := 100.0

	// Test the complete 3-bet cap scenario with actual game flow
	t.Run("Complete 3-bet cap sequence", func(t *testing.T) {
		// Start with flop dealt
		h := createFlopHistory()

		// Player 1 raises (1-bet)
		h.ActivePlayer = Player1
		options := GetActionOptionsFromHistory(h, stackSize, potSize)
		if options[0] != Check {
			t.Errorf("Player1 initial: expected first option to be Check, got %v", options[0])
		}
		h = AddToHistory(h, PlayerAction{ActionType: Raise50, Amount: 100})

		// Player 2 can re-raise (2-bet)
		options = GetActionOptionsFromHistory(h, stackSize-100, potSize+100)
		if options[0] != Call {
			t.Errorf("Player2 facing 1-bet: expected first option to be Call, got %v", options[0])
		}
		if options[len(options)-1] != Fold {
			t.Errorf("Player2 facing 1-bet: expected last option to be Fold, got %v", options[len(options)-1])
		}
		h = AddToHistory(h, PlayerAction{ActionType: Raise50, Amount: 300})

		// Player 1 can re-raise again (3-bet)
		options = GetActionOptionsFromHistory(h, stackSize-100, potSize+400)
		if options[0] != Call {
			t.Errorf("Player1 facing 2-bet: expected first option to be Call, got %v", options[0])
		}
		h = AddToHistory(h, PlayerAction{ActionType: Raise50, Amount: 900})

		// Player 2 CANNOT re-raise (facing 3-bet cap)
		options = GetActionOptionsFromHistory(h, stackSize-300, potSize+1300)
		if len(options) != 2 || options[0] != Call || options[1] != Fold {
			t.Errorf("Player2 facing 3-bet: expected [Call, Fold] only, got %v", options)
		}

		// Verify no Raise option is available
		for _, opt := range options {
			if isRaiseAction(opt) {
				t.Errorf("Raise options should not be available when facing 3-bet cap, found %v", opt)
			}
		}
	})

	// Test that the cap applies on each street independently
	t.Run("3-bet cap resets on new street", func(t *testing.T) {
		// Setup: 3-bet on flop, then call to go to turn
		h := createFlopHistory()
		h.FlopActions = []PlayerAction{
			{ActionType: Raise50, Amount: 100},
			{ActionType: Raise50, Amount: 300},
			{ActionType: Raise50, Amount: 900},
			{ActionType: Call, Amount: 900}, // Call to close betting
		}

		// Deal turn
		h = AddToHistory(h, ChanceAction{
			RevealedCards: []game.Card{{Rank: 11, Suit: "Clubs"}},
		})

		// On turn, Player1 should be able to check or raise again (cap resets)
		options := GetActionOptionsFromHistory(h, stackSize, potSize)
		if options[0] != Check {
			t.Errorf("New street should reset betting cap: expected first option to be Check, got %v", options[0])
		}
		// Should have raise options available
		if len(options) < 2 {
			t.Errorf("New street should have raise options available, got %v", options)
		}
	})
}

func TestGetActionOptionsFromHistory(t *testing.T) {
	// Use reasonable stack and pot sizes
	stackSize := 1000.0
	potSize := 100.0

	t.Run("No actions on street - can Check plus raises", func(t *testing.T) {
		h := createFlopHistory()
		h.ActivePlayer = Player1
		options := GetActionOptionsFromHistory(h, stackSize, potSize)
		if options[0] != Check {
			t.Errorf("First option should be Check, got %v", options[0])
		}
		if len(options) < 2 {
			t.Errorf("Should have raise options available, got %v", options)
		}
	})

	t.Run("After Check - can Check plus raises", func(t *testing.T) {
		h := createFlopHistory()
		h.FlopActions = []PlayerAction{{ActionType: Check, Amount: 0}}
		h.ActivePlayer = Player2
		options := GetActionOptionsFromHistory(h, stackSize, potSize)
		if options[0] != Check {
			t.Errorf("First option should be Check, got %v", options[0])
		}
	})

	t.Run("Facing 1-bet - can Call, raises, or Fold", func(t *testing.T) {
		h := createFlopHistory()
		h.FlopActions = []PlayerAction{{ActionType: Raise50, Amount: 100}}
		h.ActivePlayer = Player2
		options := GetActionOptionsFromHistory(h, stackSize, potSize+100)
		if options[0] != Call {
			t.Errorf("First option should be Call, got %v", options[0])
		}
		if options[len(options)-1] != Fold {
			t.Errorf("Last option should be Fold, got %v", options[len(options)-1])
		}
	})

	t.Run("Facing 2-bet - can Call, raises, or Fold", func(t *testing.T) {
		h := createFlopHistory()
		h.FlopActions = []PlayerAction{
			{ActionType: Raise50, Amount: 100},
			{ActionType: Raise50, Amount: 300},
		}
		h.ActivePlayer = Player1
		options := GetActionOptionsFromHistory(h, stackSize-100, potSize+400)
		if options[0] != Call {
			t.Errorf("First option should be Call, got %v", options[0])
		}
		if options[len(options)-1] != Fold {
			t.Errorf("Last option should be Fold, got %v", options[len(options)-1])
		}
	})

	t.Run("Facing 3-bet (cap) - can only Call or Fold", func(t *testing.T) {
		h := createFlopHistory()
		h.FlopActions = []PlayerAction{
			{ActionType: Raise50, Amount: 100},
			{ActionType: Raise50, Amount: 300},
			{ActionType: Raise50, Amount: 900},
		}
		h.ActivePlayer = Player2
		options := GetActionOptionsFromHistory(h, stackSize-300, potSize+1300)
		if len(options) != 2 || options[0] != Call || options[1] != Fold {
			t.Errorf("Facing 3-bet cap should only have [Call, Fold], got %v", options)
		}
	})

	t.Run("Turn - No actions yet", func(t *testing.T) {
		h := createTurnHistory()
		h.ActivePlayer = Player1
		options := GetActionOptionsFromHistory(h, stackSize, potSize)
		if options[0] != Check {
			t.Errorf("First option on turn should be Check, got %v", options[0])
		}
	})

	t.Run("River - Facing raise", func(t *testing.T) {
		h := createRiverHistory()
		h.RiverActions = []PlayerAction{{ActionType: Raise50, Amount: 100}}
		h.ActivePlayer = Player2
		options := GetActionOptionsFromHistory(h, stackSize, potSize+100)
		if options[0] != Call {
			t.Errorf("First option should be Call, got %v", options[0])
		}
		if options[len(options)-1] != Fold {
			t.Errorf("Last option should be Fold, got %v", options[len(options)-1])
		}
	})
}

func TestHistoryClone(t *testing.T) {
	// Test that Clone creates a deep copy
	original := createFlopHistory()
	original.FlopActions = []PlayerAction{
		{ActionType: Check, Amount: 0},
		{ActionType: Raise50, Amount: 100},
	}

	cloned := original.Clone()

	// Modify the cloned history
	cloned.FlopActions[0] = PlayerAction{ActionType: Raise50, Amount: 200}

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
					{ActionType: Raise50, Amount: 100},
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
				h.TurnActions = []PlayerAction{{ActionType: Raise50, Amount: 100}}
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

func TestValidRaiseSizes(t *testing.T) {
	t.Run("All raises available with large stack", func(t *testing.T) {
		// Pot=100, Stack=1000 -> all raises affordable
		// 33% = 33, 50% = 50, 75% = 75, 100% = 100, all < 1000
		h := createFlopHistory()
		h.ActivePlayer = Player1
		options := GetActionOptionsFromHistory(h, 1000, 100)

		// Should have: Check, Raise33, Raise50, Raise75, Raise100, RaiseAllIn
		if options[0] != Check {
			t.Errorf("Expected Check first, got %v", options[0])
		}
		// Check that all raise types are present
		hasRaise33 := false
		hasRaise50 := false
		hasRaise75 := false
		hasRaise100 := false
		hasRaiseAllIn := false
		for _, opt := range options {
			switch opt {
			case Raise33:
				hasRaise33 = true
			case Raise50:
				hasRaise50 = true
			case Raise75:
				hasRaise75 = true
			case Raise100:
				hasRaise100 = true
			case RaiseAllIn:
				hasRaiseAllIn = true
			}
		}
		if !hasRaise33 || !hasRaise50 || !hasRaise75 || !hasRaise100 || !hasRaiseAllIn {
			t.Errorf("Expected all raise sizes available, got %v", options)
		}
	})

	t.Run("Limited raises with small stack", func(t *testing.T) {
		// Pot=100, Stack=40
		// 33% = 33 (affordable), 50% = 50 (not affordable), 75% = 75 (not), 100% = 100 (not)
		h := createFlopHistory()
		h.ActivePlayer = Player1
		options := GetActionOptionsFromHistory(h, 40, 100)

		// Should have: Check, Raise33, RaiseAllIn
		hasRaise33 := false
		hasRaise50 := false
		hasRaiseAllIn := false
		for _, opt := range options {
			switch opt {
			case Raise33:
				hasRaise33 = true
			case Raise50:
				hasRaise50 = true
			case RaiseAllIn:
				hasRaiseAllIn = true
			}
		}
		if !hasRaise33 {
			t.Errorf("Expected Raise33 to be available (33 <= 40)")
		}
		if hasRaise50 {
			t.Errorf("Raise50 should NOT be available (50 > 40)")
		}
		if !hasRaiseAllIn {
			t.Errorf("RaiseAllIn should always be available")
		}
	})

	t.Run("Only all-in when stack is tiny", func(t *testing.T) {
		// Pot=100, Stack=10
		// 33% = 33 > 10, so only RaiseAllIn available
		h := createFlopHistory()
		h.ActivePlayer = Player1
		options := GetActionOptionsFromHistory(h, 10, 100)

		// Should have: Check, RaiseAllIn
		if len(options) != 2 {
			t.Errorf("Expected 2 options (Check, RaiseAllIn), got %v", options)
		}
		if options[0] != Check {
			t.Errorf("Expected Check, got %v", options[0])
		}
		if options[1] != RaiseAllIn {
			t.Errorf("Expected RaiseAllIn, got %v", options[1])
		}
	})

	t.Run("Facing all-in can only call or fold", func(t *testing.T) {
		h := createFlopHistory()
		h.FlopActions = []PlayerAction{{ActionType: RaiseAllIn, Amount: 500}}
		h.ActivePlayer = Player2
		options := GetActionOptionsFromHistory(h, 1000, 600)

		if len(options) != 2 || options[0] != Call || options[1] != Fold {
			t.Errorf("Facing all-in should only have [Call, Fold], got %v", options)
		}
	})

	t.Run("Call amount >= stack can only call or fold", func(t *testing.T) {
		// Opponent raised 100, we only have 80 left
		h := createFlopHistory()
		h.FlopActions = []PlayerAction{{ActionType: Raise100, Amount: 100}}
		h.ActivePlayer = Player2
		options := GetActionOptionsFromHistory(h, 80, 200)

		if len(options) != 2 || options[0] != Call || options[1] != Fold {
			t.Errorf("When call amount >= stack, should only have [Call, Fold], got %v", options)
		}
	})

	t.Run("Raise options after call based on remaining stack", func(t *testing.T) {
		// Pot=100, facing raise of 50, stack=200
		// After calling 50: stackAfterCall=150, potAfterCall=150
		// 33% of 150 = 49.5 (affordable), 50% = 75, 75% = 112.5, 100% = 150 (exactly affordable)
		h := createFlopHistory()
		h.FlopActions = []PlayerAction{{ActionType: Raise50, Amount: 50}}
		h.ActivePlayer = Player2
		options := GetActionOptionsFromHistory(h, 200, 100)

		// Should have: Call, Raise33, Raise50, Raise75, Raise100, RaiseAllIn, Fold
		if options[0] != Call {
			t.Errorf("Expected Call first, got %v", options[0])
		}
		if options[len(options)-1] != Fold {
			t.Errorf("Expected Fold last, got %v", options[len(options)-1])
		}
		// All raises should be available since 150 stack covers all pot% bets
		hasRaise100 := false
		for _, opt := range options {
			if opt == Raise100 {
				hasRaise100 = true
			}
		}
		if !hasRaise100 {
			t.Errorf("Expected Raise100 to be available, got %v", options)
		}
	})
}
