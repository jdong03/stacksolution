package action_tree

import "github.com/jdong03/stacksolution/game"

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
	// TODO: Implement best-response evaluation.
	//
	// High-level sketch:
	//   var (
	//       valueSigma float64 // Vσ
	//       valueBR1   float64 // V_BR1
	//       valueBR2   float64 // V_BR2
	//       count      int
	//   )
	//
	//   // Iterate over all valid (handP1, handP2) pairs (no card conflicts with board or each other).
	//   for each valid (p1Hand, p2Hand) {
	//       startNode := New starting PlayerNode / GameState for this hand matchup
	//
	//       // 1) Evaluate value under current strategies (σ₁, σ₂)
	//       valueSigma += bru.valueUnderCurrentStrategy(startNode)
	//
	//       // 2) Evaluate P1 best response vs σ₂
	//       valueBR1 += bru.bestResponseValueForPlayer1(startNode)
	//
	//       // 3) Evaluate P2 best response vs σ₁
	//       valueBR2 += bru.bestResponseValueForPlayer2(startNode)
	//
	//       count++
	//   }
	//
	//   if count == 0 {
	//       return 0
	//   }
	//
	//   Vσ   := valueSigma / float64(count)
	//   VBR1 := valueBR1 / float64(count)
	//   VBR2 := valueBR2 / float64(count)
	//
	//   // Exploitabilities (from P1 perspective, assuming zero-sum):
	//   //   P1 exploited by P2:   explP1 = VBR2 - Vσ
	//   //   P2 exploited by P1:   explP2 = Vσ - VBR1
	//   //
	//   // Then combine / normalize, e.g.:
	//   //   avgExploit := 0.5 * (math.Abs(explP1) + math.Abs(explP2))
	//   //   deviation  := avgExploit / potSize  // percentage of pot
	//
	//   return deviation

	return 0.0
}

// valueUnderCurrentStrategy returns the expected value for Player 1 at the given
// node, assuming both players follow the current strategy profile stored in
// trainer.InformationSetMap.
//
// This is analogous to your CFR recursion (CalculateNodeUtility), but without
// updating regrets. You can either:
//   - reuse CalculateNodeUtility if it already behaves like pure evaluation
//     when you're not updating regrets, or
//   - implement a dedicated DFS here.
func (bru *BestResponseUtility) valueUnderCurrentStrategy(node *PlayerNode) float64 {
	// TODO: Implement:
	//   - if node is terminal: return terminal payoff from P1 perspective
	//   - if chance node: average over outcomes
	//   - if decision node:
	//       * look up infoset via bru.trainer.GetInformationSet(node)
	//       * get strategy σ(I) from infoset
	//       * return sum_a σ(a|I) * childValue
	return 0.0
}

// bestResponseValueForPlayer1 returns the expected value for Player 1 at the
// given node if Player 1 plays a best response to Player 2's current strategy σ₂,
// while Player 2 continues to play σ₂.
//
// Behaviour:
//   - At P1 decision nodes: choose the action with maximum child EV (greedy).
//   - At P2 decision nodes: average over actions using σ₂ from infosets.
//   - At chance nodes: average over chance outcomes.
//   - At terminal nodes: return terminal payoff from P1 perspective.
func (bru *BestResponseUtility) bestResponseValueForPlayer1(node *PlayerNode) float64 {
	// TODO: Implement best-response DFS for P1.
	//   - Use bru.trainer.GetInformationSet(node) to fetch infosets
	//   - Use regrets / strategy sums to derive σ₂ at P2 nodes
	return 0.0
}

// bestResponseValueForPlayer2 returns the expected value for Player 1 at the
// given node if Player 2 plays a best response to Player 1's current strategy σ₁,
// while Player 1 continues to play σ₁.
//
// Behaviour:
//   - At P2 decision nodes: choose the action with maximum EV *for Player 2*,
//     which corresponds to minimum EV for Player 1 in a zero-sum game.
//   - At P1 decision nodes: average over actions using σ₁ from infosets.
//   - At chance nodes: average over chance outcomes.
//   - At terminal nodes: return terminal payoff from P1 perspective.
func (bru *BestResponseUtility) bestResponseValueForPlayer2(node *PlayerNode) float64 {
	// TODO: Implement best-response DFS for P2 (minimizing P1's EV).
	return 0.0
}
