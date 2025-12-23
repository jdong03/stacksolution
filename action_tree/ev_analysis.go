package action_tree

import (
	"fmt"
)

// P1 Equilibrium Frequencies from solver
type P1Frequencies struct {
	// First action (check vs bet)
	AACheckFreq float64
	AABetFreq   float64
	KKCheckFreq float64
	KKBetFreq   float64
	QQCheckFreq float64
	QQBetFreq   float64

	// Response to P2 bet after P1 checked (call vs fold)
	AACallFreq float64
	AAFoldFreq float64
	KKCallFreq float64
	KKFoldFreq float64
	QQCallFreq float64
	QQFoldFreq float64
}

// GetDefaultP1Frequencies returns the solver's equilibrium frequencies
func GetDefaultP1Frequencies() P1Frequencies {
	return P1Frequencies{
		// First action
		AACheckFreq: 0.049,
		AABetFreq:   0.951,
		KKCheckFreq: 0.995,
		KKBetFreq:   0.005,
		QQCheckFreq: 0.720,
		QQBetFreq:   0.280,

		// Response to P2 bet after checking
		AACallFreq: 0.999,
		AAFoldFreq: 0.001,
		KKCallFreq: 0.656,
		KKFoldFreq: 0.344,
		QQCallFreq: 0.000,
		QQFoldFreq: 1.000,
	}
}

// P2HeuristicType defines the type of P2 heuristic
type P2HeuristicType int

const (
	P2Aggressive P2HeuristicType = iota // Always bet/call
	P2Passive                           // Always check/fold
)

// Game parameters
const (
	InitialPot = 50.0 // Starting pot
	BetSize    = 25.0 // 50% pot bet
)

// Hand matchup result at showdown
// Returns 1 if P1 wins, 0.5 if tie, 0 if P2 wins
func showdownResult(p1Hand, p2Hand string) float64 {
	// On board 2d2s2c2h3d, all hands have quad 2s
	// Winner determined by kicker: AA > KK > QQ
	rankOrder := map[string]int{"AA": 3, "KK": 2, "QQ": 1}

	p1Rank := rankOrder[p1Hand]
	p2Rank := rankOrder[p2Hand]

	if p1Rank > p2Rank {
		return 1.0 // P1 wins
	} else if p1Rank < p2Rank {
		return 0.0 // P2 wins
	}
	return 0.5 // Tie
}

// CalculateEVvsHeuristics computes P1's EV against aggressive and passive P2
func CalculateEVvsHeuristics() {
	fmt.Println("\n========== P1 EV vs P2 HEURISTICS ==========")

	p1Freq := GetDefaultP1Frequencies()

	// Hand combinations (accounting for card removal)
	// With board 2d2s2c2h3d, no 2s or 3d are available
	// AA: 6 combos, KK: 6 combos, QQ: 6 combos
	hands := []string{"AA", "KK", "QQ"}
	combos := map[string]int{"AA": 6, "KK": 6, "QQ": 6}

	// Calculate EV vs Aggressive P2 (always bets/calls)
	evVsAggressive := calculateTotalEV(p1Freq, P2Aggressive, hands, combos)

	// Calculate EV vs Passive P2 (always checks/folds)
	evVsPassive := calculateTotalEV(p1Freq, P2Passive, hands, combos)

	// Calculate baseline EV (solver vs solver) - approximate as 0 since it's symmetric
	// In the non-raked game, EV should be close to 0 for symmetric ranges
	baselineEV := 0.0 // Solver vs Solver baseline

	fmt.Println("\n--- RESULTS ---")
	fmt.Printf("\nP1 (Solver) vs P2 (Aggressive - always bet/call):\n")
	fmt.Printf("  P1 EV = %.4f BB\n", evVsAggressive)
	fmt.Printf("  EV Gain vs Baseline = %.4f BB\n", evVsAggressive-baselineEV)

	fmt.Printf("\nP1 (Solver) vs P2 (Passive - always check/fold):\n")
	fmt.Printf("  P1 EV = %.4f BB\n", evVsPassive)
	fmt.Printf("  EV Gain vs Baseline = %.4f BB\n", evVsPassive-baselineEV)

	fmt.Println("\n--- DETAILED BREAKDOWN ---")
	printDetailedBreakdown(p1Freq, hands, combos)

	fmt.Println("\n================================================")
}

func calculateTotalEV(p1Freq P1Frequencies, p2Type P2HeuristicType, hands []string, combos map[string]int) float64 {
	var totalEV float64
	var totalWeight float64

	for _, p1Hand := range hands {
		for _, p2Hand := range hands {
			// Calculate weight (number of valid matchup combinations)
			weight := float64(combos[p1Hand] * combos[p2Hand])

			// Account for card removal when same hand type
			if p1Hand == p2Hand {
				// C(4,2) * C(2,2) = 6 * 1 = 6 ways for first player
				// Then only C(2,2) = 1 way for second player
				// So 6 matchups total, not 36
				weight = 6.0
			}

			ev := calculateMatchupEV(p1Hand, p2Hand, p1Freq, p2Type)
			totalEV += ev * weight
			totalWeight += weight
		}
	}

	return totalEV / totalWeight
}

func calculateMatchupEV(p1Hand, p2Hand string, p1Freq P1Frequencies, p2Type P2HeuristicType) float64 {
	// Get P1's frequencies for this hand
	var p1CheckFreq, p1BetFreq, p1CallFreq, p1FoldFreq float64

	switch p1Hand {
	case "AA":
		p1CheckFreq, p1BetFreq = p1Freq.AACheckFreq, p1Freq.AABetFreq
		p1CallFreq, p1FoldFreq = p1Freq.AACallFreq, p1Freq.AAFoldFreq
	case "KK":
		p1CheckFreq, p1BetFreq = p1Freq.KKCheckFreq, p1Freq.KKBetFreq
		p1CallFreq, p1FoldFreq = p1Freq.KKCallFreq, p1Freq.KKFoldFreq
	case "QQ":
		p1CheckFreq, p1BetFreq = p1Freq.QQCheckFreq, p1Freq.QQBetFreq
		p1CallFreq, p1FoldFreq = p1Freq.QQCallFreq, p1Freq.QQFoldFreq
	}

	showdown := showdownResult(p1Hand, p2Hand)

	var ev float64

	// Game tree:
	// P1 Check (freq: p1CheckFreq)
	//   -> P2 Check (passive) or P2 Bet (aggressive)
	//      -> If P2 checks: showdown with pot = 50
	//      -> If P2 bets: P1 calls or folds
	// P1 Bet (freq: p1BetFreq)
	//   -> P2 Fold (passive) or P2 Call (aggressive)
	//      -> If P2 folds: P1 wins pot = 50
	//      -> If P2 calls: showdown with pot = 100

	switch p2Type {
	case P2Aggressive:
		// P2 always bets when facing check, always calls when facing bet

		// Branch 1: P1 checks -> P2 bets -> P1 responds
		// EV = P1CheckFreq * (P1CallFreq * ShowdownEV + P1FoldFreq * FoldEV)
		// ShowdownEV after call: pot = 50 + 25 + 25 = 100, P1 invested extra 25
		showdownEVAfterCall := showdown*100.0 - 25.0 - 25.0 // P1 wins pot or loses, minus P1's bet and initial contribution
		foldEV := -25.0                                     // P1 loses initial pot contribution

		branch1EV := p1CheckFreq * (p1CallFreq*showdownEVAfterCall + p1FoldFreq*foldEV)

		// Branch 2: P1 bets -> P2 calls -> showdown
		// pot = 50 + 25 + 25 = 100
		showdownEVAfterP1Bet := showdown*100.0 - 25.0 - 25.0
		branch2EV := p1BetFreq * showdownEVAfterP1Bet

		ev = branch1EV + branch2EV

	case P2Passive:
		// P2 always checks when facing check, always folds when facing bet

		// Branch 1: P1 checks -> P2 checks -> showdown
		// pot = 50
		showdownEVCheck := showdown*50.0 - 25.0 // P1 wins pot minus initial contribution
		branch1EV := p1CheckFreq * showdownEVCheck

		// Branch 2: P1 bets -> P2 folds -> P1 wins pot
		// P1 wins pot = 50, invested 25 initially
		winPotEV := 50.0 - 25.0 // P1 wins pot minus initial contribution
		branch2EV := p1BetFreq * winPotEV

		ev = branch1EV + branch2EV
	}

	return ev
}

func printDetailedBreakdown(p1Freq P1Frequencies, hands []string, combos map[string]int) {
	fmt.Println("\n[P1 vs Aggressive P2]")
	fmt.Printf("%-10s %-10s %10s %10s %10s\n", "P1 Hand", "P2 Hand", "Matchups", "EV/Matchup", "Total EV")
	fmt.Println(string(make([]byte, 55)))

	for _, p1Hand := range hands {
		for _, p2Hand := range hands {
			weight := float64(combos[p1Hand] * combos[p2Hand])
			if p1Hand == p2Hand {
				weight = 6.0
			}
			ev := calculateMatchupEV(p1Hand, p2Hand, p1Freq, P2Aggressive)
			fmt.Printf("%-10s %-10s %10.0f %10.4f %10.4f\n", p1Hand, p2Hand, weight, ev, ev*weight)
		}
	}

	fmt.Println("\n[P1 vs Passive P2]")
	fmt.Printf("%-10s %-10s %10s %10s %10s\n", "P1 Hand", "P2 Hand", "Matchups", "EV/Matchup", "Total EV")
	fmt.Println(string(make([]byte, 55)))

	for _, p1Hand := range hands {
		for _, p2Hand := range hands {
			weight := float64(combos[p1Hand] * combos[p2Hand])
			if p1Hand == p2Hand {
				weight = 6.0
			}
			ev := calculateMatchupEV(p1Hand, p2Hand, p1Freq, P2Passive)
			fmt.Printf("%-10s %-10s %10.0f %10.4f %10.4f\n", p1Hand, p2Hand, weight, ev, ev*weight)
		}
	}
}
