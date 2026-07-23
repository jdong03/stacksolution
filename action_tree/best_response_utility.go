package action_tree

import (
	"math"

	"github.com/jdong03/stacksolution/game"
)

// BestResponseUtility encapsulates logic for evaluating exploitability of the
// current strategy profile stored in a VanillaCFRTrainer.
//
// Conceptually:
//   - It reads the current strategies (σ₁, σ₂) from trainer.InformationSetMap.
//   - For each player i, it computes the value of a best response BRᵢ(σ_{-i})
//     against the opponent's current strategy.
//   - It compares those best-response values to the current strategy value
//     to get an "exploitability" / deviation from Nash.
//
// A perfect Nash equilibrium strategy profile will have zero deviation
// (no player can improve by deviating).
type BestResponseUtility struct {
	trainer *VanillaCFRTrainer
	// You can add config here later, e.g. pot size normalization, flags for
	// using average strategy vs current strategy, etc.
}

// NewBestResponseUtility constructs a BestResponseUtility instance bound
// to a specific trainer. It will use the trainer's InformationSetMap and
// GetInformationSet function to reconstruct infosets and read strategies.
func NewBestResponseUtility(trainer *VanillaCFRTrainer) *BestResponseUtility {
	return &BestResponseUtility{
		trainer: trainer,
	}
}

// TotalDeviation computes a scalar measure of how exploitable the current
// strategy profile (σ₁, σ₂) is, given a fixed public board and sets of
// private hand combinations for both players.
//
// Parameters:
//   - board:       public board cards (flop/turn/river), in whatever
//     representation you use downstream to construct start nodes.
//   - handCombosP1: all possible hole-card combinations for Player 1.
//   - handCombosP2: all possible hole-card combinations for Player 2.
//
// Interpretation:
//   - It should:
//     1. Compute Vσ = EV for Player 1 when both players play (σ₁, σ₂).
//     2. Compute V_BR1 = EV for Player 1 when P1 plays best response BR₁(σ₂).
//     3. Compute V_BR2 = EV for Player 1 when P2 plays best response BR₂(σ₁).
//     4. Turn those into exploitabilities / a single deviation metric
//     (e.g. as a percentage of pot size).
//
// Return value:
//   - A scalar exploitability / deviation metric. At Nash, this should
//     converge toward 0.
func (bru *BestResponseUtility) TotalDeviation(
	board []game.Card,
	handCombosP1 [][]game.Card,
	handCombosP2 [][]game.Card,
) float64 {
	// Use the trainer's initial pot size (or a default)
	initialPotSize := 50.0 // Default pot size, could be made configurable

	// 1) Value under current strategies (sigma1, sigma2): a fixed strategy
	// pair's EV is a plain average over the deal distribution, so the
	// (p1Hand, p2Hand) cross product is the correct thing to average over
	// here - no best-responder is involved, so there's nothing for it to
	// see or not see.
	var valueSigma float64
	var count int
	for _, p1Hand := range handCombosP1 {
		if cardsConflict(p1Hand, board) {
			continue
		}
		for _, p2Hand := range handCombosP2 {
			if cardsConflict(p2Hand, p1Hand) || cardsConflict(p2Hand, board) {
				continue
			}
			startNode := bru.trainer.createStartingNodeWithBoard(board, p1Hand, p2Hand, initialPotSize)
			valueSigma += bru.valueUnderCurrentStrategy(startNode)
			count++
		}
	}
	if count == 0 {
		return 0
	}
	avgVSigma := valueSigma / float64(count)

	// 2) V_BR1: P1 best-responds to P2's average strategy. This must be
	// computed per P1 hand (P1's own info sets can only depend on P1's own
	// card), with a single belief-weighted walk over P2's whole range - see
	// bestResponseValueForPlayer1's doc comment for why the naive per-hand-pair
	// version was wrong.
	var valueBR1 float64
	var countP1 int
	for _, p1Hand := range handCombosP1 {
		if cardsConflict(p1Hand, board) {
			continue
		}
		valueBR1 += bru.bestResponseValueForPlayer1(board, p1Hand, handCombosP2, initialPotSize)
		countP1++
	}
	if countP1 == 0 {
		return 0
	}
	avgVBR1 := valueBR1 / float64(countP1)

	// 3) V_BR2: symmetric, P2 best-responds to P1's average strategy.
	var valueBR2 float64
	var countP2 int
	for _, p2Hand := range handCombosP2 {
		if cardsConflict(p2Hand, board) {
			continue
		}
		valueBR2 += bru.bestResponseValueForPlayer2(board, handCombosP1, p2Hand, initialPotSize)
		countP2++
	}
	if countP2 == 0 {
		return 0
	}
	avgVBR2 := valueBR2 / float64(countP2)

	// Exploitabilities (from P1 perspective, assuming zero-sum):
	//   P1's gain from best response: explP1 = VBR1 - Vσ (how much P1 gains by deviating)
	//   P2's gain from best response: explP2 = Vσ - VBR2 (how much P2 gains, which reduces P1's value)
	explP1 := avgVBR1 - avgVSigma
	explP2 := avgVSigma - avgVBR2

	// Total exploitability is the sum of how much each player can gain by deviating
	// This equals zero at Nash equilibrium
	totalExploitability := explP1 + explP2

	return totalExploitability
}

// valueUnderCurrentStrategy returns the expected value for Player 1 at the given
// node, assuming both players follow the current strategy profile stored in
// trainer.InformationSetMap.
//
// This is analogous to your CFR recursion (CalculateNodeUtility), but without
// updating regrets. Uses the average strategy (GetFinalStrategy) for evaluation.
func (bru *BestResponseUtility) valueUnderCurrentStrategy(node GameStateNode) float64 {
	switch n := node.(type) {
	case *LeafNode:
		return bru.calculateLeafValue(n)

	case *ChanceNode:
		return bru.calculateChanceValue(n, bru.valueUnderCurrentStrategy)

	case *PlayerNode:
		// Get information set and average strategy
		infoSet := bru.trainer.GetInformationSet(n)
		strategy := GetFinalStrategy(infoSet)

		// Calculate expected value as weighted sum over actions
		var nodeValue float64
		for i, actionType := range n.ActionOptions {
			actionProb := strategy[i]
			if actionProb == 0 {
				continue
			}

			action := bru.trainer.createActionForType(actionType, n)
			childNode := NewGameStateNode(n, action, actionProb)
			childValue := bru.valueUnderCurrentStrategy(childNode)
			nodeValue += actionProb * childValue
		}
		return nodeValue

	default:
		panic("Unknown node type in valueUnderCurrentStrategy")
	}
}

// bestResponseValueForPlayer1 returns the expected value for Player 1, over
// Player 1's own hand p1Hand, when Player 1 plays a true information-set-level
// best response to Player 2's average strategy sigma2 (Player 2 continues to
// play sigma2).
//
// This replaces a version that built one concrete (p1Hand, p2Hand) node per
// hand-combo pair and maximized independently within each - which let the
// "best responder" implicitly see Player 2's exact hand, since the action
// chosen at a given info set could silently vary depending on which p2Hand
// happened to be in that outer-loop iteration. That is exactly the
// omniscient-best-responder bug already diagnosed and fixed for Kuhn poker
// (see kuhn/exploitability.go and docs/ROADMAP.md's 2026-07-22 entry): it
// produces a strictly-too-high exploitability number that looks like a
// solver bug even when the trainer is correct.
//
// The fix: hold p1Hand fixed and carry a belief distribution over which of
// Player 2's hands (from handCombosP2) is real, split by Player 2's actual
// strategy at Player 2's decision nodes and by each candidate's own available
// cards at chance nodes, and commit to exactly one action per Player 1 info
// set - maximizing the belief-weighted aggregate value across all still-live
// candidates, never a value computed for one specific opponent hand.
func (bru *BestResponseUtility) bestResponseValueForPlayer1(
	board []game.Card,
	p1Hand []game.Card,
	handCombosP2 [][]game.Card,
	initialPotSize float64,
) float64 {
	var candidates [][]game.Card
	for _, p2Hand := range handCombosP2 {
		if cardsConflict(p2Hand, p1Hand) || cardsConflict(p2Hand, board) {
			continue
		}
		candidates = append(candidates, p2Hand)
	}
	if len(candidates) == 0 {
		return 0
	}

	weights := make([]float64, len(candidates))
	for i := range weights {
		weights[i] = 1.0 / float64(len(candidates)) // uniform prior over P2's range
	}

	history := NewHistory()
	if len(board) >= 3 {
		history.FlopCards = board[0:3]
	}
	if len(board) >= 4 {
		history.TurnCard = board[3:4]
	}
	if len(board) >= 5 {
		history.RiverCard = board[4:5]
	}
	history.ActivePlayer = Player1

	return bru.brNode1(history, bru.trainer.Player1InitialStackSize, bru.trainer.Player2InitialStackSize,
		initialPotSize, initialPotSize, p1Hand, candidates, weights)
}

// brNode1 is the belief-propagation walk for bestResponseValueForPlayer1: p1Hand
// is fixed for the whole walk (Player 1 always knows their own card), and
// candidates/weights track the current belief over Player 2's possible hands
// at this exact public history. p1Stack/p2Stack/potSize are shared across all
// candidates - they're determined entirely by the public history and actions
// taken, never by which hand is real.
func (bru *BestResponseUtility) brNode1(
	history *History,
	p1Stack, p2Stack, potSize, initialPotSize float64,
	p1Hand []game.Card,
	candidates [][]game.Card,
	weights []float64,
) float64 {
	switch history.ActivePlayer {
	case Player1:
		options := GetActionOptionsFromHistory(history, p1Stack, potSize)
		best := math.Inf(-1)
		for _, actionType := range options {
			action := bru.trainer.createActionForType(actionType, actionNodeStub(history, p1Stack, p2Stack, potSize))
			v := bru.brApply1(history, p1Stack, p2Stack, potSize, initialPotSize, action, p1Hand, candidates, weights)
			if v > best {
				best = v
			}
		}
		return best

	case Player2:
		options := GetActionOptionsFromHistory(history, p2Stack, potSize)
		// Each candidate has its own card, hence its own info set, hence its
		// own average strategy - fetch each once per action loop below.
		candidateStrategies := make([][]float64, len(candidates))
		for ci, cand := range candidates {
			if weights[ci] == 0 {
				continue
			}
			key := buildInfoSetKey(cand, history)
			if infoSet, ok := bru.trainer.InformationSetMap[key]; ok {
				candidateStrategies[ci] = GetFinalStrategy(infoSet)
			} else {
				candidateStrategies[ci] = uniformStrategy(len(options))
			}
		}
		var total float64
		for ai, actionType := range options {
			action := bru.trainer.createActionForType(actionType, actionNodeStub(history, p1Stack, p2Stack, potSize))
			childWeights := make([]float64, len(candidates))
			for ci := range candidates {
				if weights[ci] == 0 {
					continue
				}
				childWeights[ci] = weights[ci] * candidateStrategies[ci][ai]
			}
			total += bru.brApply1(history, p1Stack, p2Stack, potSize, initialPotSize, action, p1Hand, candidates, childWeights)
		}
		return total

	default:
		panic("brNode1 called on a non-decision history")
	}
}

// brApply1 applies one action (shared across all candidates - it's a public
// action, not dependent on which hand is real) and dispatches on the
// resulting history: to another decision node, a chance node (with
// per-candidate available-card filtering), or a leaf (with per-candidate
// showdown evaluation).
func (bru *BestResponseUtility) brApply1(
	history *History,
	p1Stack, p2Stack, potSize, initialPotSize float64,
	action PlayerAction,
	p1Hand []game.Card,
	candidates [][]game.Card,
	weights []float64,
) float64 {
	actingPlayer := history.ActivePlayer
	newHistory := AddToHistory(history, action)

	newP1Stack, newP2Stack, newPot := p1Stack, p2Stack, potSize
	if actingPlayer == Player1 {
		newP1Stack -= action.Amount
	} else {
		newP2Stack -= action.Amount
	}
	newPot += action.Amount

	switch newHistory.ActivePlayer {
	case Player1, Player2:
		return bru.brNode1(newHistory, newP1Stack, newP2Stack, newPot, initialPotSize, p1Hand, candidates, weights)

	case Chance:
		return bru.brChance1(newHistory, newP1Stack, newP2Stack, newPot, initialPotSize, p1Hand, candidates, weights)

	case Leaf:
		var total float64
		for ci, cand := range candidates {
			if weights[ci] == 0 {
				continue
			}
			total += weights[ci] * brLeafValue(bru.trainer.Player1InitialStackSize, newHistory,
				newP1Stack, newP2Stack, newPot, initialPotSize, p1Hand, cand, actingPlayer, action)
		}
		return total

	default:
		panic("brApply1: unexpected active player after transition")
	}
}

// brChance1 deals with a chance node under belief propagation: candidates
// whose own hand contains the revealed card are impossible for that branch
// (filtered out), and each surviving candidate's weight is split by *its own*
// probability of that card being the one dealt (1 / its own available-card
// count), since different candidate hands remove different cards from the
// deck. Candidates that would still be live for a given card are kept
// together for that branch, not recursed independently - that's what lets
// brNode1's Player 1 maximization aggregate over the whole surviving belief
// set instead of one candidate at a time.
func (bru *BestResponseUtility) brChance1(
	history *History,
	p1Stack, p2Stack, potSize, initialPotSize float64,
	p1Hand []game.Card,
	candidates [][]game.Card,
	weights []float64,
) float64 {
	availablePerCandidate := make([]map[game.Card]bool, len(candidates))
	cardUnion := map[game.Card]bool{}
	for ci, cand := range candidates {
		if weights[ci] == 0 {
			continue
		}
		gs := GameState{History: *history, Player1Cards: p1Hand, Player2Cards: cand}
		avail := determineAvailableCards(gs)
		m := make(map[game.Card]bool, len(avail))
		for _, c := range avail {
			m[c] = true
			cardUnion[c] = true
		}
		availablePerCandidate[ci] = m
	}

	var total float64
	for card := range cardUnion {
		childWeights := make([]float64, len(candidates))
		var anyLive bool
		for ci := range candidates {
			m := availablePerCandidate[ci]
			if weights[ci] == 0 || m == nil || !m[card] {
				continue
			}
			childWeights[ci] = weights[ci] / float64(len(m))
			anyLive = true
		}
		if !anyLive {
			continue
		}
		newHistory := AddToHistory(history, ChanceAction{RevealedCards: []game.Card{card}})
		total += bru.brNode1(newHistory, p1Stack, p2Stack, potSize, initialPotSize, p1Hand, candidates, childWeights)
	}
	return total
}

// bestResponseValueForPlayer2 is bestResponseValueForPlayer1's mirror image:
// Player 2's own hand p2Hand is fixed, belief is carried over Player 1's
// possible hands (handCombosP1), and Player 2 minimizes Player 1's utility
// (equivalently, maximizes its own) since all utilities are stored as u1.
func (bru *BestResponseUtility) bestResponseValueForPlayer2(
	board []game.Card,
	handCombosP1 [][]game.Card,
	p2Hand []game.Card,
	initialPotSize float64,
) float64 {
	var candidates [][]game.Card
	for _, p1Hand := range handCombosP1 {
		if cardsConflict(p1Hand, p2Hand) || cardsConflict(p1Hand, board) {
			continue
		}
		candidates = append(candidates, p1Hand)
	}
	if len(candidates) == 0 {
		return 0
	}

	weights := make([]float64, len(candidates))
	for i := range weights {
		weights[i] = 1.0 / float64(len(candidates))
	}

	history := NewHistory()
	if len(board) >= 3 {
		history.FlopCards = board[0:3]
	}
	if len(board) >= 4 {
		history.TurnCard = board[3:4]
	}
	if len(board) >= 5 {
		history.RiverCard = board[4:5]
	}
	history.ActivePlayer = Player1

	return bru.brNode2(history, bru.trainer.Player1InitialStackSize, bru.trainer.Player2InitialStackSize,
		initialPotSize, initialPotSize, p2Hand, candidates, weights)
}

func (bru *BestResponseUtility) brNode2(
	history *History,
	p1Stack, p2Stack, potSize, initialPotSize float64,
	p2Hand []game.Card,
	candidates [][]game.Card,
	weights []float64,
) float64 {
	switch history.ActivePlayer {
	case Player2:
		options := GetActionOptionsFromHistory(history, p2Stack, potSize)
		// P2 minimizes P1's utility (= maximizes its own, u2 = -u1).
		best := math.Inf(1)
		for _, actionType := range options {
			action := bru.trainer.createActionForType(actionType, actionNodeStub(history, p1Stack, p2Stack, potSize))
			v := bru.brApply2(history, p1Stack, p2Stack, potSize, initialPotSize, action, p2Hand, candidates, weights)
			if v < best {
				best = v
			}
		}
		return best

	case Player1:
		options := GetActionOptionsFromHistory(history, p1Stack, potSize)
		candidateStrategies := make([][]float64, len(candidates))
		for ci, cand := range candidates {
			if weights[ci] == 0 {
				continue
			}
			key := buildInfoSetKey(cand, history)
			if infoSet, ok := bru.trainer.InformationSetMap[key]; ok {
				candidateStrategies[ci] = GetFinalStrategy(infoSet)
			} else {
				candidateStrategies[ci] = uniformStrategy(len(options))
			}
		}
		var total float64
		for ai, actionType := range options {
			action := bru.trainer.createActionForType(actionType, actionNodeStub(history, p1Stack, p2Stack, potSize))
			childWeights := make([]float64, len(candidates))
			for ci := range candidates {
				if weights[ci] == 0 {
					continue
				}
				childWeights[ci] = weights[ci] * candidateStrategies[ci][ai]
			}
			total += bru.brApply2(history, p1Stack, p2Stack, potSize, initialPotSize, action, p2Hand, candidates, childWeights)
		}
		return total

	default:
		panic("brNode2 called on a non-decision history")
	}
}

func (bru *BestResponseUtility) brApply2(
	history *History,
	p1Stack, p2Stack, potSize, initialPotSize float64,
	action PlayerAction,
	p2Hand []game.Card,
	candidates [][]game.Card,
	weights []float64,
) float64 {
	actingPlayer := history.ActivePlayer
	newHistory := AddToHistory(history, action)

	newP1Stack, newP2Stack, newPot := p1Stack, p2Stack, potSize
	if actingPlayer == Player1 {
		newP1Stack -= action.Amount
	} else {
		newP2Stack -= action.Amount
	}
	newPot += action.Amount

	switch newHistory.ActivePlayer {
	case Player1, Player2:
		return bru.brNode2(newHistory, newP1Stack, newP2Stack, newPot, initialPotSize, p2Hand, candidates, weights)

	case Chance:
		return bru.brChance2(newHistory, newP1Stack, newP2Stack, newPot, initialPotSize, p2Hand, candidates, weights)

	case Leaf:
		var total float64
		for ci, cand := range candidates {
			if weights[ci] == 0 {
				continue
			}
			total += weights[ci] * brLeafValue(bru.trainer.Player1InitialStackSize, newHistory,
				newP1Stack, newP2Stack, newPot, initialPotSize, cand, p2Hand, actingPlayer, action)
		}
		return total

	default:
		panic("brApply2: unexpected active player after transition")
	}
}

func (bru *BestResponseUtility) brChance2(
	history *History,
	p1Stack, p2Stack, potSize, initialPotSize float64,
	p2Hand []game.Card,
	candidates [][]game.Card,
	weights []float64,
) float64 {
	availablePerCandidate := make([]map[game.Card]bool, len(candidates))
	cardUnion := map[game.Card]bool{}
	for ci, cand := range candidates {
		if weights[ci] == 0 {
			continue
		}
		gs := GameState{History: *history, Player1Cards: cand, Player2Cards: p2Hand}
		avail := determineAvailableCards(gs)
		m := make(map[game.Card]bool, len(avail))
		for _, c := range avail {
			m[c] = true
			cardUnion[c] = true
		}
		availablePerCandidate[ci] = m
	}

	var total float64
	for card := range cardUnion {
		childWeights := make([]float64, len(candidates))
		var anyLive bool
		for ci := range candidates {
			m := availablePerCandidate[ci]
			if weights[ci] == 0 || m == nil || !m[card] {
				continue
			}
			childWeights[ci] = weights[ci] / float64(len(m))
			anyLive = true
		}
		if !anyLive {
			continue
		}
		newHistory := AddToHistory(history, ChanceAction{RevealedCards: []game.Card{card}})
		total += bru.brNode2(newHistory, p1Stack, p2Stack, potSize, initialPotSize, p2Hand, candidates, childWeights)
	}
	return total
}

// actionNodeStub wraps shared (candidate-independent) path state in a
// throwaway *PlayerNode so trainer.createActionForType - which only reads
// History/stack sizes/pot size, never cards - can be reused as-is instead of
// duplicating its bet-sizing logic here.
func actionNodeStub(history *History, p1Stack, p2Stack, potSize float64) *PlayerNode {
	return &PlayerNode{
		GameState: GameState{
			History:          *history,
			Player1StackSize: p1Stack,
			Player2StackSize: p2Stack,
			PotSize:          potSize,
		},
	}
}

// uniformStrategy is the fallback belief for an info set CFR never actually
// visited (matches GetStrategy's own uniform fallback when all regrets are
// non-positive).
func uniformStrategy(numActions int) []float64 {
	s := make([]float64, numActions)
	if numActions == 0 {
		return s
	}
	p := 1.0 / float64(numActions)
	for i := range s {
		s[i] = p
	}
	return s
}

// brLeafValue computes P1's utility at a terminal history for one specific
// (p1Hand, p2Hand) pair, replicating NewLeafNode's pot-crediting exactly
// (win/lose/tie) before handing off to the single shared leafUtility formula.
func brLeafValue(
	player1InitialStackSize float64,
	history *History,
	p1Stack, p2Stack, potSize, initialPotSize float64,
	p1Hand, p2Hand []game.Card,
	lastActivePlayer Player,
	lastAction PlayerAction,
) float64 {
	gs := GameState{
		History:          *history,
		Player1Cards:     p1Hand,
		Player2Cards:     p2Hand,
		Player1StackSize: p1Stack,
		Player2StackSize: p2Stack,
		PotSize:          potSize,
		InitialPotSize:   initialPotSize,
	}
	switch determineWinner(gs, lastActivePlayer, lastAction) {
	case 1:
		gs.Player1StackSize += gs.PotSize
	case -1:
		gs.Player2StackSize += gs.PotSize
	case 0:
		gs.Player1StackSize += gs.PotSize / 2
		gs.Player2StackSize += gs.PotSize / 2
	}
	return leafUtility(gs, player1InitialStackSize)
}

// ComputeAverageP1Utility computes the average P1 utility across all hand matchups
// when both players follow their trained strategies.
func (bru *BestResponseUtility) ComputeAverageP1Utility(
	board []game.Card,
	handCombosP1 [][]game.Card,
	handCombosP2 [][]game.Card,
	initialPotSize float64,
) float64 {
	var totalUtil float64
	var count int

	for _, p1Hand := range handCombosP1 {
		if cardsConflict(p1Hand, board) {
			continue
		}
		for _, p2Hand := range handCombosP2 {
			if cardsConflict(p2Hand, p1Hand) || cardsConflict(p2Hand, board) {
				continue
			}
			startNode := bru.trainer.createStartingNodeWithBoard(board, p1Hand, p2Hand, initialPotSize)
			totalUtil += bru.valueUnderCurrentStrategy(startNode)
			count++
		}
	}

	if count == 0 {
		return 0
	}
	return totalUtil / float64(count)
}

// calculateLeafValue returns P1's utility at a terminal node.
func (bru *BestResponseUtility) calculateLeafValue(node *LeafNode) float64 {
	return leafUtility(node.GetGameState(), bru.trainer.Player1InitialStackSize)
}

// calculateChanceValue averages utility over all possible chance outcomes.
// Takes a recursive evaluation function to allow reuse across different evaluation modes.
func (bru *BestResponseUtility) calculateChanceValue(node *ChanceNode, evalFunc func(GameStateNode) float64) float64 {
	availableCards := node.AvailableCards
	if len(availableCards) == 0 {
		return 0.0
	}

	// Each card has equal probability
	actionProbability := 1.0 / float64(len(availableCards))

	var nodeUtility float64

	// Iterate over each possible card that could be dealt
	for _, card := range availableCards {
		chanceAction := ChanceAction{
			RevealedCards: []game.Card{card},
		}

		childNode := NewGameStateNode(node, chanceAction, actionProbability)
		childUtility := evalFunc(childNode)
		nodeUtility += actionProbability * childUtility
	}

	return nodeUtility
}
