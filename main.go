package main

import (
	"fmt"
	"os"

	"github.com/jdong03/stacksolution/action_tree"
	"github.com/jdong03/stacksolution/game"
)

func main() {
	Run()
}

func Run() {
	fmt.Println("Starting Poker Solver - Vanilla CFR")
	fmt.Println("\n===================================")

	// Game configuration
	numIterations := 1000
	boardStr := "2d, 2s, 2c, 2h, 3d"
	player1Range := []string{"AA", "KK", "QQ"}
	player2Range := []string{"AA", "KK", "QQ"}
	startPotSize := 50.0

	// Parse board
	board := game.ParseBoard(boardStr)
	fmt.Printf("Board: %v\n", board)

	// Generate hand combinations
	var p1Combos [][]game.Card
	for _, handRange := range player1Range {
		combos := game.GetHandCombinations(handRange)
		p1Combos = append(p1Combos, combos...)
	}
	fmt.Printf("Player 1 hand combinations: %d\n", len(p1Combos))

	var p2Combos [][]game.Card
	for _, handRange := range player2Range {
		combos := game.GetHandCombinations(handRange)
		p2Combos = append(p2Combos, combos...)
	}
	fmt.Printf("Player 2 hand combinations: %d\n", len(p2Combos))

	fmt.Printf("Starting pot: %.0f BB\n", startPotSize)
	fmt.Printf("Stack sizes: %.0f BB\n\n", action_tree.Player1InitialStackSize)

	// Create trainer with metrics tracking
	trainer := action_tree.NewVanillaCFRTrainer()

	// Set up metrics to track QQ hands - FIRST ACTION ONLY (opening action on river)
	// Q = rank 12, shown as "Qs", "Qh", "Qd", "Qc"
	metrics := action_tree.NewTrainingMetrics()
	metrics.FirstActionOnly = true // Only track opening action, not responses
	metrics.TrackInfoSet("Qs_Q")   // Track QQ hands (QsQh, QsQd, QsQc)
	metrics.TrackInfoSet("Qh_Q")   // QhQd, QhQc
	metrics.TrackInfoSet("Qd_Qc")  // QdQc
	trainer.Metrics = metrics

	// Train
	fmt.Println("Training...")
	avgUtil := trainer.Train(numIterations, board, p1Combos, p2Combos, startPotSize)

	// Display results
	action_tree.PrintSummary(trainer, avgUtil, numIterations)
	action_tree.DisplayStrategies(trainer, len(board))

	// Display P1 first action strategies for all hands
	action_tree.DisplayFirstActionStrategies(trainer)

	// Compare solver strategy against simple heuristic
	action_tree.EvaluateVsHeuristic(trainer, board, p1Combos, p2Combos, startPotSize, action_tree.SimpleHeuristic())

	// Write strategy file (remove old one first)
	os.Remove("strategies.txt")
	action_tree.WriteStrategiesToFile(trainer, len(board), numIterations, avgUtil, "strategies.txt")

	// Clean up and recreate data directory for CSV files
	os.RemoveAll("data")
	os.MkdirAll("data", 0755)

	// Write CSV files for graphing
	metrics.WriteExploitabilityCSV("data/exploitability.csv")
	metrics.WriteBetFrequencyCSV(trainer, "data/bet_frequency.csv")
	metrics.WriteAllTrackedStrategiesCSV(trainer, "data/strategy")

	fmt.Println("\nCSV files written to data/ directory:")
	fmt.Println("  - data/exploitability.csv")
	fmt.Println("  - data/bet_frequency.csv")
	fmt.Println("  - data/strategy_*.csv (for each tracked info set)")
	fmt.Println("\nRun 'python3 plot_metrics.py' to generate graphs")
}
