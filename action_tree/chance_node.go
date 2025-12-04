package action_tree

import "github.com/jdong03/stacksolution/game"

type ChanceNode struct {
	GameState
	AvailableCards []game.Card
}

func (p *ChanceNode) GetGameState() GameState {
	return p.GameState
}

/*
NewChanceNode creates a new ChanceNode from a parent node and new game state
*/
func NewChanceNode(parentGameStateNode PlayerNode, action PlayerAction, actionProbability float64, newHistory *History) *ChanceNode {
	player1StackSize := parentGameStateNode.Player1StackSize
	player2StackSize := parentGameStateNode.Player2StackSize
	player1ReachProbability := parentGameStateNode.Player1ReachProbability
	player2ReachProbability := parentGameStateNode.Player2ReachProbability

	// Update stack sizes and reach probabilities based on the action taken
	if parentGameStateNode.History.ActivePlayer == Player1 {
		player1StackSize -= action.Amount
		player1ReachProbability *= actionProbability
	} else if parentGameStateNode.History.ActivePlayer == Player2 {
		player2StackSize -= action.Amount
		player2ReachProbability *= actionProbability
	}

	gameState := GameState{
		History:                 *newHistory,
		Player1Cards:            parentGameStateNode.Player1Cards,
		Player2Cards:            parentGameStateNode.Player2Cards,
		Player1StackSize:        player1StackSize,
		Player2StackSize:        player2StackSize,
		Player1ReachProbability: player1ReachProbability,
		Player2ReachProbability: player2ReachProbability,
	}

	availableCards := determineAvailableCards(gameState)

	return &ChanceNode{
		GameState:      gameState,
		AvailableCards: availableCards,
	}
}

func determineAvailableCards(gameState GameState) []game.Card {
	flopCards := gameState.History.FlopCards
	turnCard := gameState.History.TurnCard
	riverCard := gameState.History.RiverCard

	player1Cards := gameState.Player1Cards
	player2Cards := gameState.Player2Cards

	deck := game.NewDeck()

	available := game.CardDifference(
		deck.Cards,
		flopCards,
		turnCard,
		riverCard,
		player1Cards,
		player2Cards,
	)
	return available
}
