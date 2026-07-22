package kuhn

import (
	"math"
	"math/rand"

	"github.com/jdong03/stacksolution/utils"
)

// InfoSet holds cumulative regrets and strategy weights for one (card,
// history) information set, following the same regret-matching convention
// as action_tree/information_set.go: RegretSum drives the current strategy,
// StrategySum accumulates the reach-weighted average strategy that converges
// to equilibrium.
type InfoSet struct {
	RegretSum   []float64
	StrategySum []float64
}

func newInfoSet() *InfoSet {
	return &InfoSet{
		RegretSum:   make([]float64, len(Actions)),
		StrategySum: make([]float64, len(Actions)),
	}
}

// currentStrategy computes sigma(a) = max(RegretSum[a], 0) / sum, falling
// back to uniform when all regrets are non-positive.
func (is *InfoSet) currentStrategy() []float64 {
	posRegrets := make([]float64, len(is.RegretSum))
	for i, r := range is.RegretSum {
		posRegrets[i] = math.Max(0, r)
	}
	return utils.Normalize(posRegrets)
}

// AverageStrategy returns the equilibrium approximation: the StrategySum
// normalized to a probability distribution.
func (is *InfoSet) AverageStrategy() []float64 {
	return utils.Normalize(is.StrategySum)
}

// infoSetKey identifies an information set by the acting player's own card
// plus the public betting history - never the opponent's card.
func infoSetKey(card Card, history string) string {
	return card.String() + history
}

// Trainer runs vanilla CFR over the full Kuhn poker game tree.
type Trainer struct {
	infoSets map[string]*InfoSet
	rng      *rand.Rand
}

// NewTrainer creates a Trainer with its own seeded RNG for deal shuffling,
// so runs are reproducible.
func NewTrainer(seed int64) *Trainer {
	return &Trainer{
		infoSets: make(map[string]*InfoSet),
		rng:      rand.New(rand.NewSource(seed)),
	}
}

// InfoSets exposes the trained information sets (read-only by convention).
// A correctly-built Kuhn tree produces exactly 12 of these: 4 non-terminal
// histories ("", "p", "b", "pb") x 3 possible cards for the player to act.
func (t *Trainer) InfoSets() map[string]*InfoSet {
	return t.infoSets
}

func (t *Trainer) getOrCreateInfoSet(card Card, history string) *InfoSet {
	key := infoSetKey(card, history)
	is, ok := t.infoSets[key]
	if !ok {
		is = newInfoSet()
		t.infoSets[key] = is
	}
	return is
}

// Train runs vanilla CFR for the given number of iterations, dealing a fresh
// random pair of distinct cards each iteration, and returns the average
// per-iteration game value to Player 1. As iterations grow this trends
// toward the known Kuhn poker value of -1/18.
func (t *Trainer) Train(iterations int) float64 {
	cards := append([]Card{}, Deck...)
	var totalUtil float64

	for i := 0; i < iterations; i++ {
		t.rng.Shuffle(len(cards), func(a, b int) { cards[a], cards[b] = cards[b], cards[a] })
		dealt := [2]Card{cards[0], cards[1]}
		totalUtil += t.cfr(dealt, "", 1.0, 1.0)
	}

	return totalUtil / float64(iterations)
}

// cfr recursively computes the counterfactual value to Player 1 of the
// current node, updating regrets and strategy sums for whichever player is
// acting. p1/p2 are Player 1's and Player 2's reach probabilities to this
// node under the current strategy profile.
//
// All returned utilities are from Player 1's perspective throughout (never
// negated on the way back up) - Player 2's regret update flips sign instead,
// matching the zero-sum convention used in action_tree/information_set.go.
func (t *Trainer) cfr(cards [2]Card, history string, p1, p2 float64) float64 {
	if IsTerminal(history) {
		return TerminalUtility(history, cards)
	}

	active := activePlayer(history)
	infoSet := t.getOrCreateInfoSet(cards[active], history)
	strategy := infoSet.currentStrategy()

	activeReach := p1
	if active == 1 {
		activeReach = p2
	}
	for i, prob := range strategy {
		infoSet.StrategySum[i] += activeReach * prob
	}

	actionUtils := make([]float64, len(Actions))
	var nodeUtil float64
	for i, a := range Actions {
		nextHistory := history + string(a)
		var u float64
		if active == 0 {
			u = t.cfr(cards, nextHistory, p1*strategy[i], p2)
		} else {
			u = t.cfr(cards, nextHistory, p1, p2*strategy[i])
		}
		actionUtils[i] = u
		nodeUtil += strategy[i] * u
	}

	// Regret is weighted by the *opponent's* (counterfactual) reach - never
	// the acting player's own reach.
	if active == 0 {
		cfReach := p2
		for i := range actionUtils {
			infoSet.RegretSum[i] += cfReach * (actionUtils[i] - nodeUtil)
		}
	} else {
		cfReach := p1
		// Utilities are stored as u1(*); Player 2's regret is
		// u2(a) - u2(sigma) = u1(sigma) - u1(a), hence the flipped operands.
		for i := range actionUtils {
			infoSet.RegretSum[i] += cfReach * (nodeUtil - actionUtils[i])
		}
	}

	return nodeUtil
}
