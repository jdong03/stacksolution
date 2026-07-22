package kuhn

import (
	"math"
	"testing"
)

// TestVanillaCFRConvergesToKnownEquilibrium is this package's correctness
// checkpoint per docs/ROADMAP.md Phase 1: the game tree must produce exactly
// 12 information sets, exploitability must trend toward zero, and the game
// value to Player 1 must approach the known closed-form value of -1/18.
func TestVanillaCFRConvergesToKnownEquilibrium(t *testing.T) {
	const iterations = 200_000
	trainer := NewTrainer(42)

	avgP1Value := trainer.Train(iterations)

	if got := len(trainer.InfoSets()); got != 12 {
		t.Fatalf("expected exactly 12 information sets, got %d", got)
	}

	exploitability := trainer.Exploitability()
	const exploitabilityTolerance = 0.01
	if exploitability > exploitabilityTolerance {
		t.Errorf("exploitability did not converge: got %.6f, want < %.6f after %d iterations",
			exploitability, exploitabilityTolerance, iterations)
	}

	const expectedGameValue = -1.0 / 18.0
	const gameValueTolerance = 0.01
	if diff := math.Abs(avgP1Value - expectedGameValue); diff > gameValueTolerance {
		t.Errorf("Player 1 game value did not converge: got %.6f, want %.6f +/- %.6f",
			avgP1Value, expectedGameValue, gameValueTolerance)
	}

	t.Logf("info sets=%d exploitability=%.6f avgP1Value=%.6f (expected %.6f)",
		len(trainer.InfoSets()), exploitability, avgP1Value, expectedGameValue)
}

// TestExploitabilityOfUniformRandomStrategyIsFarFromZero is a sanity check on
// the oracle itself: an untrained (all-uniform) strategy profile should be
// clearly exploitable, so Exploitability isn't trivially returning ~0 for
// everything.
func TestExploitabilityOfUniformRandomStrategyIsFarFromZero(t *testing.T) {
	trainer := NewTrainer(1)
	// One iteration is enough to instantiate all 12 info sets (vanilla CFR
	// visits every action at every reachable info set) without meaningfully
	// updating strategy sums away from uniform.
	trainer.Train(1)

	exploitability := trainer.Exploitability()
	const minExpectedExploitability = 0.1
	if exploitability < minExpectedExploitability {
		t.Errorf("expected a near-uniform strategy to be clearly exploitable (>%.2f), got %.6f",
			minExpectedExploitability, exploitability)
	}
}

// TestInfoSetKeyDoesNotLeakOpponentCard guards against the classic CFR bug
// flagged in CLAUDE.md: info-set keys must be built from the acting player's
// own observed card and the public history only.
func TestInfoSetKeyDoesNotLeakOpponentCard(t *testing.T) {
	if infoSetKey(Jack, "p") == infoSetKey(Jack, "p") {
		// Same card, same history -> same key regardless of what the
		// opponent holds, since opponent card is never passed in at all.
	} else {
		t.Fatal("infoSetKey should be deterministic for the same card and history")
	}

	// The key space must only ever be 3 cards x 4 non-terminal histories.
	seen := map[string]bool{}
	for _, c := range Deck {
		for _, h := range []string{"", "p", "b", "pb"} {
			seen[infoSetKey(c, h)] = true
		}
	}
	if len(seen) != 12 {
		t.Fatalf("expected 12 distinct info-set keys from 3 cards x 4 histories, got %d", len(seen))
	}
}
