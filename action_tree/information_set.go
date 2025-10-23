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
    return Normalize(posRegrets)
}

func AddToStrategySum(infoSet *InformationSet, strategy []float64, activePlayerReachProbability float64) {
	for i, prob := range strategy {
		infoSet.StrategySum[i] += prob * activePlayerReachProbability
	}
}

func GetFinalStrategy()
