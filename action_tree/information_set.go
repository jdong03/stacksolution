package action_tree

import (
	"math"

	"github.com/jdong03/stacksolution/utils"
)

/*
InfoSet represents a single information set in the game tree.

An information set groups together all concrete game states that look identical
to the acting player - for example, the same public board, same action history,
and same private cards (or card bucket). From this player's viewpoint, all of
these game states are indistinguishable, so they share one InfoSet object.

Each InfoSet stores two arrays of cumulative values across CFR iterations:
- RegretSum[a]: total regret for taking action "a"
- StrategySum[a]: total weighted frequency of playing action "a"

The algorithm updates these over time to derive both the current and the average
strategy that converge toward equilibrium.
*/

type InformationSet struct {
	/* RegretSum holds the cumulative regret for each legal action.
	Positive regret means the action has historically performed better
	than the current mixed strategy; negative regret means worse. */
	RegretSum []float64

	/* StrategySum accumulates how often each action has been chosen,
	   weighted by the reach probability of this player.  It is later
	   normalized to compute the final average strategy. */
	StrategySum []float64
}

/*
NewInformationSet constructs and returns a new NewInformationSet for a situation that has
'numActions' legal actions available. Both arrays values are initialized to zero.
*/
func NewInformationSet(numActions int) *InformationSet {
	return &InformationSet{
		RegretSum:   make([]float64, numActions),
		StrategySum: make([]float64, numActions),
	}
}

/*
GetStrategy computes the current mixed strategy (probability distribution)
for this information set using the regret-matching rule.

Actions with higher positive cumulative regrets are given proportionally
higher probability weights:

	σ(a) = max(RegretSum[a], 0) / Σ_b max(RegretSum[b], 0)

If all regrets are non-positive, a uniform strategy is returned.
This method is typically called once per iteration to determine
the strategy used at this node during the CFR traversal.
*/
func GetStrategy(infoSet *InformationSet) []float64 {
	posRegrets := make([]float64, len(infoSet.RegretSum))
	for i, regret := range infoSet.RegretSum {
		posRegrets[i] = math.Max(0, regret)
	}
	return utils.Normalize(posRegrets)
}

/*
AddToStrategySum updates the cumulative StrategySum using the current
strategy probabilities, weighted by the active player's reach probability (πᵢ).

This keeps track of how frequently each action has been taken in expectation
over all previous iterations.
*/
func AddToStrategySum(infoSet *InformationSet, strategy []float64, activePlayerReachProbability float64) {
	for i, prob := range strategy {
		infoSet.StrategySum[i] += prob * activePlayerReachProbability
	}
}

/*
GetFinalStrategy returns the average strategy derived from the cumulative
StrategySum values. This represents the long-run equilibrium policy
approximated by the CFR algorithm:

	σ̄(a) = StrategySum[a] / Σ_b StrategySum[b]

If the sums are all zero, a uniform distribution is returned.
*/
func GetFinalStrategy(infoSet *InformationSet) []float64 {
	return utils.Normalize(infoSet.StrategySum)
}

/*
AddToCumulativeRegrets updates the cumulative counterfactual regrets at an
information set for the *acting* player, as in Counterfactual Regret
Minimization (CFR).
CFR update (per action a at infoset I, acting player i):
  r_t(I, a) += π^{-i}(I) * ( u_i(σ_{I→a}) - u_i(σ) )
where
  • π^{-i}(I) is the counterfactual reach of the *other side* (opponent × chance),
    i.e., how often the opponent (and chance) deliver us to I independent of i’s choices.
    In practice we drop chance because it is constant across actions at I and cancels
    in normalization; so we use only the opponent’s reach.
  • u_i(σ_{I→a}) is the utility for the acting player i if we FORCE action a at I,
    then follow the current strategy σ thereafter.
  • u_i(σ) is the utility at I under the current mixed strategy.
This implementation assumes a *zero-sum* game and that `actionUtilities` and
`nodeUtility` are measured from **Player1’s** perspective (u₁). Therefore at Player2
nodes we must flip the sign (since u₂ = −u₁):
  u₂(σ_{I→a}) - u₂(σ) = u₁(σ) - u₁(σ_{I→a})
*/

func AddToCumulativeRegrets(infoSet *InformationSet,
	gameStateNode *GameStateNode,
	actionUtilities []float64, // u₁(σ_{I→a}) for each legal action a at I
	nodeUtility float64, // u₁(σ) at I
) {
	if gameStateNode.ActivePlayer == Player1 {
		counterFactualReachProbability := gameStateNode.Player2ReachProbability
		for i := range actionUtilities {
			// r[a] += π^{-1}(I) * (u₁(σ_{I→a}) - u₁(σ))
			infoSet.RegretSum[i] += counterFactualReachProbability * (actionUtilities[i] - nodeUtility)
		}
	} else if gameStateNode.ActivePlayer == Player2 {
		counterFactualReachProbability := gameStateNode.Player1ReachProbability
		for i := range actionUtilities {
			// Utilities are stored as u₁(·). For zero-sum, u₂ = −u₁, hence sign flip:
			// r[a] += π^{-2}(I) * (u₂(σ_{I→a}) - u₂(σ)) = π^{-2}(I) * (u₁(σ) - u₁(σ_{I→a}))
			infoSet.RegretSum[i] += counterFactualReachProbability * (nodeUtility - actionUtilities[i])
		}
	}
}
