package action_tree

import (
	"testing"

	"github.com/jdong03/stacksolution/game"
)

/*
Tests for game_state_node.go

This test suite covers:
1. GetStartingNode - creates initial PlayerNode with correct stack sizes and reach probabilities
2. NewGameStateNode transitions:
   - PlayerNode -> PlayerNode (when player raises or checks)
   - PlayerNode -> ChanceNode (when street ends via call or check-check)
   - PlayerNode -> LeafNode (on fold, river call, or river check by Player2)
   - ChanceNode -> PlayerNode (after dealing cards)
3. Stack size updates - decremented by bet amounts
4. Reach probability updates - multiplied by action probabilities
5. Action options - determined by GetActionOptionsFromHistory
6. 3-bet cap enforcement - only Call/Fold after 3 consecutive raises

Clarification answers based on implementation:
1. 3-bet cap is tested in TestNewGameStateNode_ThreeBetCap
2. Pot size is not currently tracked in GameState
3. ActionOptions are set using GetActionOptionsFromHistory (see player_node.go line 68)
4. ChanceAction probabilities would be 1/(remaining cards in deck)
5. River check by Player2 ending the game is tested in TestNewGameStateNode_RiverCheckByPlayer2
*/

func TestGetStartingNode(t *testing.T) {
	// Create sample hole cards
	player1Cards := []game.Card{
		{Rank: 14, Suit: "Hearts"}, // Ace of Hearts
		{Rank: 14, Suit: "Spades"}, // Ace of Spades
	}
	player2Cards := []game.Card{
		{Rank: 13, Suit: "Hearts"}, // King of Hearts
		{Rank: 13, Suit: "Clubs"},  // King of Clubs
	}

	node := GetStartingNode(player1Cards, player2Cards)

	// Verify it returns a PlayerNode (starts with Player1 action on flop)
	playerNode, ok := node.(*PlayerNode)
	if !ok {
		t.Fatalf("Expected PlayerNode, got %T", node)
	}

	// Verify game state
	gameState := playerNode.GetGameState()

	// Check initial stack sizes (should be default)
	if gameState.Player1StackSize != Player1InitialStackSize {
		t.Errorf("Expected Player1 stack %f, got %f", Player1InitialStackSize, gameState.Player1StackSize)
	}
	if gameState.Player2StackSize != Player2InitialStackSize {
		t.Errorf("Expected Player2 stack %f, got %f", Player2InitialStackSize, gameState.Player2StackSize)
	}

	// Check initial reach probabilities
	if gameState.Player1ReachProbability != 1.0 {
		t.Errorf("Expected Player1 reach probability 1.0, got %f", gameState.Player1ReachProbability)
	}
	if gameState.Player2ReachProbability != 1.0 {
		t.Errorf("Expected Player2 reach probability 1.0, got %f", gameState.Player2ReachProbability)
	}

	// Check cards are properly assigned
	if len(gameState.Player1Cards) != 2 {
		t.Errorf("Expected 2 cards for Player1, got %d", len(gameState.Player1Cards))
	}
	if len(gameState.Player2Cards) != 2 {
		t.Errorf("Expected 2 cards for Player2, got %d", len(gameState.Player2Cards))
	}

	// Check history - Player1 should act first
	if gameState.History.ActivePlayer != Player1 {
		t.Errorf("Expected Player1 to act first, got %v", gameState.History.ActivePlayer)
	}

	// Check action options - should be Check, Raise at the start
	expectedOptions := []EnumActionType{Check, Raise}
	if len(playerNode.ActionOptions) != len(expectedOptions) {
		t.Errorf("Expected %d action options, got %d", len(expectedOptions), len(playerNode.ActionOptions))
	}
	for i, opt := range expectedOptions {
		if i < len(playerNode.ActionOptions) && playerNode.ActionOptions[i] != opt {
			t.Errorf("Expected action option %v at index %d, got %v", opt, i, playerNode.ActionOptions[i])
		}
	}
}

func TestNewGameStateNode_PlayerToPlayer_Raise(t *testing.T) {
	// Test Player1 raises -> Player2's turn
	startNode := createTestStartingNode()

	raiseAction := PlayerAction{
		ActionType: Raise,
		Amount:     10,
	}

	newNode := NewGameStateNode(startNode, raiseAction, 0.5)

	// Should return a PlayerNode
	playerNode, ok := newNode.(*PlayerNode)
	if !ok {
		t.Fatalf("Expected PlayerNode after Player1 raise, got %T", newNode)
	}

	gameState := playerNode.GetGameState()
	originalGameState := startNode.GetGameState()

	// Active player should be Player2
	if gameState.History.ActivePlayer != Player2 {
		t.Errorf("Expected Player2 to act after Player1 raise, got %v", gameState.History.ActivePlayer)
	}

	// Player1's stack should be reduced by raise amount
	expectedStack := originalGameState.Player1StackSize - 10
	if gameState.Player1StackSize != expectedStack {
		t.Errorf("Expected Player1 stack %f after raise, got %f", expectedStack, gameState.Player1StackSize)
	}

	// Player2's stack should be unchanged
	if gameState.Player2StackSize != originalGameState.Player2StackSize {
		t.Errorf("Player2 stack should be unchanged, expected %f, got %f",
			originalGameState.Player2StackSize, gameState.Player2StackSize)
	}

	// Player1's reach probability should be updated
	expectedReach := originalGameState.Player1ReachProbability * 0.5
	if gameState.Player1ReachProbability != expectedReach {
		t.Errorf("Expected Player1 reach probability %f, got %f", expectedReach, gameState.Player1ReachProbability)
	}

	// Player2's reach probability should be unchanged
	if gameState.Player2ReachProbability != originalGameState.Player2ReachProbability {
		t.Errorf("Player2 reach probability should be unchanged, expected %f, got %f",
			originalGameState.Player2ReachProbability, gameState.Player2ReachProbability)
	}

	// Check that raise was added to flop actions
	if len(gameState.History.FlopActions) != 1 {
		t.Errorf("Expected 1 flop action, got %d", len(gameState.History.FlopActions))
	}
	if gameState.History.FlopActions[0].ActionType != Raise {
		t.Errorf("Expected Raise action in history, got %v", gameState.History.FlopActions[0].ActionType)
	}

	// Player2 should have Call, Raise, Fold options
	expectedOptions := []EnumActionType{Call, Raise, Fold}
	if len(playerNode.ActionOptions) != len(expectedOptions) {
		t.Errorf("Expected %d action options for Player2, got %d", len(expectedOptions), len(playerNode.ActionOptions))
	}
	for i, opt := range expectedOptions {
		if i < len(playerNode.ActionOptions) && playerNode.ActionOptions[i] != opt {
			t.Errorf("Expected action option %v at index %d, got %v", opt, i, playerNode.ActionOptions[i])
		}
	}
}

func TestNewGameStateNode_PlayerToPlayer_Check(t *testing.T) {
	// Test Player1 checks -> Player2's turn
	startNode := createTestStartingNode()

	checkAction := PlayerAction{
		ActionType: Check,
		Amount:     0,
	}

	newNode := NewGameStateNode(startNode, checkAction, 1.0)

	// Should return a PlayerNode
	playerNode, ok := newNode.(*PlayerNode)
	if !ok {
		t.Fatalf("Expected PlayerNode after Player1 check, got %T", newNode)
	}

	gameState := playerNode.GetGameState()

	// Active player should be Player2
	if gameState.History.ActivePlayer != Player2 {
		t.Errorf("Expected Player2 to act after Player1 check, got %v", gameState.History.ActivePlayer)
	}

	// Both stacks should be unchanged
	originalGameState := startNode.GetGameState()
	if gameState.Player1StackSize != originalGameState.Player1StackSize {
		t.Errorf("Player1 stack should be unchanged after check")
	}
	if gameState.Player2StackSize != originalGameState.Player2StackSize {
		t.Errorf("Player2 stack should be unchanged after check")
	}

	// Player2 should have Check, Raise options
	expectedOptions := []EnumActionType{Check, Raise}
	if len(playerNode.ActionOptions) != len(expectedOptions) {
		t.Errorf("Expected %d action options for Player2 after check, got %d", len(expectedOptions), len(playerNode.ActionOptions))
	}
	for i, opt := range expectedOptions {
		if i < len(playerNode.ActionOptions) && playerNode.ActionOptions[i] != opt {
			t.Errorf("Expected action option %v at index %d, got %v", opt, i, playerNode.ActionOptions[i])
		}
	}
}

func TestNewGameStateNode_PlayerToChance_Call(t *testing.T) {
	// Test Player2 calls Player1's raise -> ChanceNode
	startNode := createTestStartingNode()

	// Player1 raises
	raiseAction := PlayerAction{ActionType: Raise, Amount: 10}
	node1 := NewGameStateNode(startNode, raiseAction, 1.0)

	// Player2 calls -> should create ChanceNode
	callAction := PlayerAction{ActionType: Call, Amount: 10}
	node2 := NewGameStateNode(node1, callAction, 0.8)

	// Should return a ChanceNode
	chanceNode, ok := node2.(*ChanceNode)
	if !ok {
		t.Fatalf("Expected ChanceNode after call, got %T", node2)
	}

	gameState := chanceNode.GetGameState()
	node1GameState := node1.GetGameState()

	// Active player should be Chance (waiting to deal turn card)
	if gameState.History.ActivePlayer != Chance {
		t.Errorf("Expected Chance to be active, got %v", gameState.History.ActivePlayer)
	}

	// Player2's stack should be reduced by call amount
	expectedStack := node1GameState.Player2StackSize - 10
	if gameState.Player2StackSize != expectedStack {
		t.Errorf("Expected Player2 stack %f after call, got %f", expectedStack, gameState.Player2StackSize)
	}

	// Player2's reach probability should be updated
	expectedReach := node1GameState.Player2ReachProbability * 0.8
	if gameState.Player2ReachProbability != expectedReach {
		t.Errorf("Expected Player2 reach probability %f, got %f", expectedReach, gameState.Player2ReachProbability)
	}
}

func TestNewGameStateNode_PlayerToChance_CheckCheck(t *testing.T) {
	// Test check-check -> ChanceNode
	startNode := createTestStartingNode()

	// Player1 checks
	checkAction := PlayerAction{ActionType: Check, Amount: 0}
	node1 := NewGameStateNode(startNode, checkAction, 1.0)

	// Player2 checks -> should create ChanceNode
	node2 := NewGameStateNode(node1, checkAction, 1.0)

	// Should return a ChanceNode
	chanceNode, ok := node2.(*ChanceNode)
	if !ok {
		t.Fatalf("Expected ChanceNode after check-check, got %T", node2)
	}

	gameState := chanceNode.GetGameState()

	// Active player should be Chance (waiting for turn card)
	if gameState.History.ActivePlayer != Chance {
		t.Errorf("Expected Chance to be active, got %v", gameState.History.ActivePlayer)
	}

	// Both stacks should be unchanged
	originalGameState := startNode.GetGameState()
	if gameState.Player1StackSize != originalGameState.Player1StackSize {
		t.Errorf("Stacks should be unchanged after check-check")
	}
	if gameState.Player2StackSize != originalGameState.Player2StackSize {
		t.Errorf("Stacks should be unchanged after check-check")
	}
}

func TestNewGameStateNode_PlayerToLeaf_Fold(t *testing.T) {
	// Test Player2 folds to raise -> LeafNode
	startNode := createTestStartingNode()

	// Player1 raises
	raiseAction := PlayerAction{ActionType: Raise, Amount: 10}
	node1 := NewGameStateNode(startNode, raiseAction, 1.0)

	// Player2 folds -> should create LeafNode
	foldAction := PlayerAction{ActionType: Fold, Amount: 0}
	node2 := NewGameStateNode(node1, foldAction, 0.3)

	// Should return a LeafNode
	leafNode, ok := node2.(*LeafNode)
	if !ok {
		t.Fatalf("Expected LeafNode after fold, got %T", node2)
	}

	gameState := leafNode.GetGameState()

	// Active player should be Leaf (game over)
	if gameState.History.ActivePlayer != Leaf {
		t.Errorf("Expected Leaf (game over) after fold, got %v", gameState.History.ActivePlayer)
	}

	// Player2's stack should be unchanged (fold costs nothing)
	node1GameState := node1.GetGameState()
	if gameState.Player2StackSize != node1GameState.Player2StackSize {
		t.Errorf("Player2 stack should be unchanged after fold")
	}

	// Player2's reach probability should be updated
	expectedReach := node1GameState.Player2ReachProbability * 0.3
	if gameState.Player2ReachProbability != expectedReach {
		t.Errorf("Expected Player2 reach probability %f, got %f", expectedReach, gameState.Player2ReachProbability)
	}
}

func TestNewGameStateNode_PlayerToLeaf_RiverCall(t *testing.T) {
	// Test river call -> LeafNode
	// Create a node on the river where Player2 faces a raise
	h := NewHistory()
	h.FlopCards = []game.Card{
		{Rank: 14, Suit: "Hearts"},
		{Rank: 13, Suit: "Hearts"},
		{Rank: 12, Suit: "Hearts"},
	}
	h.TurnCard = []game.Card{{Rank: 11, Suit: "Hearts"}}
	h.RiverCard = []game.Card{{Rank: 10, Suit: "Hearts"}}
	h.RiverActions = []PlayerAction{{ActionType: Raise, Amount: 20}}
	h.ActivePlayer = Player2

	parentNode := &PlayerNode{
		GameState: GameState{
			History:                 *h,
			Player1Cards:            []game.Card{{Rank: 14, Suit: "Spades"}, {Rank: 14, Suit: "Clubs"}},
			Player2Cards:            []game.Card{{Rank: 13, Suit: "Spades"}, {Rank: 13, Suit: "Clubs"}},
			Player1StackSize:        80, // Already bet 20
			Player2StackSize:        100,
			Player1ReachProbability: 1.0,
			Player2ReachProbability: 1.0,
		},
		ActionOptions: []EnumActionType{Call, Raise, Fold},
	}

	// Player2 calls on river
	callAction := PlayerAction{ActionType: Call, Amount: 20}
	newNode := NewGameStateNode(parentNode, callAction, 0.6)

	// Should return a LeafNode
	leafNode, ok := newNode.(*LeafNode)
	if !ok {
		t.Fatalf("Expected LeafNode after river call, got %T", newNode)
	}

	gameState := leafNode.GetGameState()

	// Active player should be Leaf (game over)
	if gameState.History.ActivePlayer != Leaf {
		t.Errorf("Expected Leaf after river call, got %v", gameState.History.ActivePlayer)
	}

	// Player2's stack should be reduced by call amount
	expectedStack := parentNode.Player2StackSize - 20
	if gameState.Player2StackSize != expectedStack {
		t.Errorf("Expected Player2 stack %f after river call, got %f", expectedStack, gameState.Player2StackSize)
	}

	// Player2's reach probability should be updated
	expectedReach := parentNode.Player2ReachProbability * 0.6
	if gameState.Player2ReachProbability != expectedReach {
		t.Errorf("Expected Player2 reach probability %f, got %f", expectedReach, gameState.Player2ReachProbability)
	}
}

func TestNewGameStateNode_ChanceToPlayer(t *testing.T) {
	// Test ChanceNode dealing turn -> PlayerNode
	startNode := createTestStartingNode()

	// Player1 checks
	checkAction := PlayerAction{ActionType: Check, Amount: 0}
	node1 := NewGameStateNode(startNode, checkAction, 1.0)

	// Player2 checks -> ChanceNode
	node2 := NewGameStateNode(node1, checkAction, 1.0)

	chanceNode, ok := node2.(*ChanceNode)
	if !ok {
		t.Fatalf("Expected ChanceNode after check-check, got %T", node2)
	}

	// ChanceNode deals turn card
	turnAction := ChanceAction{
		RevealedCards: []game.Card{
			{Rank: 9, Suit: "Clubs"},
		},
	}

	node3 := NewGameStateNode(chanceNode, turnAction, 1.0/45.0)

	// Should return a PlayerNode
	playerNode, ok := node3.(*PlayerNode)
	if !ok {
		t.Fatalf("Expected PlayerNode after dealing turn, got %T", node3)
	}

	gameState := playerNode.GetGameState()

	// Player1 should act first on new street
	if gameState.History.ActivePlayer != Player1 {
		t.Errorf("Expected Player1 to act first on turn, got %v", gameState.History.ActivePlayer)
	}

	// Turn card should be dealt
	if len(gameState.History.TurnCard) != 1 {
		t.Errorf("Expected 1 turn card, got %d", len(gameState.History.TurnCard))
	}

	// Player1 should have Check, Raise options on new street
	expectedOptions := []EnumActionType{Check, Raise}
	if len(playerNode.ActionOptions) != len(expectedOptions) {
		t.Errorf("Expected %d action options on new street, got %d", len(expectedOptions), len(playerNode.ActionOptions))
	}
	for i, opt := range expectedOptions {
		if i < len(playerNode.ActionOptions) && playerNode.ActionOptions[i] != opt {
			t.Errorf("Expected action option %v at index %d, got %v", opt, i, playerNode.ActionOptions[i])
		}
	}

	// Stack sizes and reach probabilities should be unchanged by chance action
	originalGameState := chanceNode.GetGameState()
	if gameState.Player1StackSize != originalGameState.Player1StackSize {
		t.Errorf("Player1 stack should be unchanged after chance action")
	}
	if gameState.Player2StackSize != originalGameState.Player2StackSize {
		t.Errorf("Player2 stack should be unchanged after chance action")
	}
	if gameState.Player1ReachProbability != originalGameState.Player1ReachProbability {
		t.Errorf("Player1 reach probability should be unchanged after chance action")
	}
	if gameState.Player2ReachProbability != originalGameState.Player2ReachProbability {
		t.Errorf("Player2 reach probability should be unchanged after chance action")
	}
}

func TestNewGameStateNode_ThreeBetCap(t *testing.T) {
	// Test 3-bet cap scenario
	startNode := createTestStartingNode()

	// Player1 raises (1-bet)
	raise1 := PlayerAction{ActionType: Raise, Amount: 10}
	node1 := NewGameStateNode(startNode, raise1, 1.0)

	// Player2 raises (2-bet)
	raise2 := PlayerAction{ActionType: Raise, Amount: 30}
	node2 := NewGameStateNode(node1, raise2, 1.0)

	// Player1 raises (3-bet)
	raise3 := PlayerAction{ActionType: Raise, Amount: 50}
	node3 := NewGameStateNode(node2, raise3, 1.0)

	playerNode, ok := node3.(*PlayerNode)
	if !ok {
		t.Fatalf("Expected PlayerNode after 3-bet, got %T", node3)
	}

	// Player2 should only have Call, Fold options (no Raise due to cap)
	expectedOptions := []EnumActionType{Call, Fold}
	if len(playerNode.ActionOptions) != len(expectedOptions) {
		t.Errorf("Expected %d action options facing 3-bet, got %d", len(expectedOptions), len(playerNode.ActionOptions))
	}
	for i, opt := range expectedOptions {
		if i < len(playerNode.ActionOptions) && playerNode.ActionOptions[i] != opt {
			t.Errorf("Expected action option %v at index %d, got %v", opt, i, playerNode.ActionOptions[i])
		}
	}

	// Verify no Raise option
	for _, opt := range playerNode.ActionOptions {
		if opt == Raise {
			t.Error("Raise should not be available when facing 3-bet cap")
		}
	}
}

func TestNewGameStateNode_CompleteGameFlow(t *testing.T) {
	// Test a complete game flow: flop -> turn -> river -> showdown
	startNode := createTestStartingNode()

	// Flop: check-check
	check := PlayerAction{ActionType: Check, Amount: 0}
	node1 := NewGameStateNode(startNode, check, 1.0)
	node2 := NewGameStateNode(node1, check, 1.0) // -> ChanceNode

	// Deal turn
	turnAction := ChanceAction{
		RevealedCards: []game.Card{{Rank: 11, Suit: "Clubs"}},
	}
	node3 := NewGameStateNode(node2, turnAction, 1.0/45.0) // -> PlayerNode

	// Turn: Player1 raises, Player2 calls
	raise := PlayerAction{ActionType: Raise, Amount: 10}
	node4 := NewGameStateNode(node3, raise, 1.0)
	call := PlayerAction{ActionType: Call, Amount: 10}
	node5 := NewGameStateNode(node4, call, 1.0) // -> ChanceNode

	// Deal river
	riverAction := ChanceAction{
		RevealedCards: []game.Card{{Rank: 10, Suit: "Diamonds"}},
	}
	node6 := NewGameStateNode(node5, riverAction, 1.0/44.0) // -> PlayerNode

	// River: Player1 raises, Player2 calls -> showdown
	node7 := NewGameStateNode(node6, raise, 1.0)
	finalNode := NewGameStateNode(node7, call, 1.0) // -> LeafNode

	// Should end in LeafNode
	leafNode, ok := finalNode.(*LeafNode)
	if !ok {
		t.Fatalf("Expected LeafNode at showdown, got %T", finalNode)
	}

	gameState := leafNode.GetGameState()

	// Game should be over
	if gameState.History.ActivePlayer != Leaf {
		t.Errorf("Expected Leaf (game over) at showdown, got %v", gameState.History.ActivePlayer)
	}

	// Verify final stack sizes (both players bet 20 total)
	originalGameState := startNode.GetGameState()
	expectedP1Stack := originalGameState.Player1StackSize - 20
	expectedP2Stack := originalGameState.Player2StackSize - 20

	if gameState.Player1StackSize != expectedP1Stack {
		t.Errorf("Expected Player1 final stack %f, got %f", expectedP1Stack, gameState.Player1StackSize)
	}
	if gameState.Player2StackSize != expectedP2Stack {
		t.Errorf("Expected Player2 final stack %f, got %f", expectedP2Stack, gameState.Player2StackSize)
	}

	// Verify cards were dealt
	if len(gameState.History.FlopCards) != 3 {
		t.Errorf("Expected 3 flop cards, got %d", len(gameState.History.FlopCards))
	}
	if len(gameState.History.TurnCard) != 1 {
		t.Errorf("Expected 1 turn card, got %d", len(gameState.History.TurnCard))
	}
	if len(gameState.History.RiverCard) != 1 {
		t.Errorf("Expected 1 river card, got %d", len(gameState.History.RiverCard))
	}
}

func TestNewGameStateNode_RiverCheckByPlayer2(t *testing.T) {
	// Test transition from PlayerNode to LeafNode on river check by Player2
	h := NewHistory()
	h.FlopCards = []game.Card{
		{Rank: 14, Suit: "Hearts"},
		{Rank: 13, Suit: "Hearts"},
		{Rank: 12, Suit: "Hearts"},
	}
	h.FlopActions = []PlayerAction{{ActionType: Check, Amount: 0}, {ActionType: Check, Amount: 0}}
	h.TurnCard = []game.Card{{Rank: 11, Suit: "Clubs"}}
	h.TurnActions = []PlayerAction{{ActionType: Check, Amount: 0}, {ActionType: Check, Amount: 0}}
	h.RiverCard = []game.Card{{Rank: 10, Suit: "Diamonds"}}
	h.RiverActions = []PlayerAction{{ActionType: Check, Amount: 0}}
	h.ActivePlayer = Player2

	parentNode := &PlayerNode{
		GameState: GameState{
			History:                 *h,
			Player1Cards:            []game.Card{{Rank: 14, Suit: "Spades"}, {Rank: 14, Suit: "Clubs"}},
			Player2Cards:            []game.Card{{Rank: 13, Suit: "Spades"}, {Rank: 13, Suit: "Clubs"}},
			Player1StackSize:        100.0,
			Player2StackSize:        100.0,
			Player1ReachProbability: 0.5,
			Player2ReachProbability: 0.7,
		},
		ActionOptions: []EnumActionType{Check, Raise},
	}

	// Player2 checks on river (ends the game)
	checkAction := PlayerAction{ActionType: Check, Amount: 0}
	newNode := NewGameStateNode(parentNode, checkAction, 0.8)

	// Should return a LeafNode since river check by Player2 ends the game
	leafNode, ok := newNode.(*LeafNode)
	if !ok {
		t.Fatalf("Expected LeafNode after river check by Player2, got %T", newNode)
	}

	gameState := leafNode.GetGameState()

	// Active player should be Leaf (game over)
	if gameState.History.ActivePlayer != Leaf {
		t.Errorf("Expected Leaf after river check by Player2, got %v", gameState.History.ActivePlayer)
	}

	// Stacks should be unchanged
	if gameState.Player1StackSize != 100.0 {
		t.Errorf("Expected Player1 stack 100.0, got %f", gameState.Player1StackSize)
	}
	if gameState.Player2StackSize != 100.0 {
		t.Errorf("Expected Player2 stack 100.0, got %f", gameState.Player2StackSize)
	}

	// Player2's reach probability should be updated
	expectedReach := 0.56 // 0.7 * 0.8
	tolerance := 0.0001
	diff := gameState.Player2ReachProbability - expectedReach
	if diff < -tolerance || diff > tolerance {
		t.Errorf("Expected Player2 reach probability %f, got %f", expectedReach, gameState.Player2ReachProbability)
	}
}

// Helper function to create a test starting node with flop already dealt
func createTestStartingNode() *PlayerNode {
	player1Cards := []game.Card{
		{Rank: 14, Suit: "Hearts"},
		{Rank: 14, Suit: "Spades"},
	}
	player2Cards := []game.Card{
		{Rank: 13, Suit: "Hearts"},
		{Rank: 13, Suit: "Clubs"},
	}

	// Create starting node
	node := GetStartingNode(player1Cards, player2Cards).(*PlayerNode)

	// Add flop cards to history for testing
	node.GameState.History.FlopCards = []game.Card{
		{Rank: 12, Suit: "Hearts"},
		{Rank: 11, Suit: "Hearts"},
		{Rank: 10, Suit: "Hearts"},
	}

	// Update action options since we're on flop
	node.ActionOptions = GetActionOptionsFromHistory(&node.GameState.History)

	return node
}