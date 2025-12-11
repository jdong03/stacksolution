package main

import (
	"fmt"

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
	numIterations := 50
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

	// Create trainer
	trainer := action_tree.NewVanillaCFRTrainer()

	// Train
	fmt.Println("Training...")
	avgUtil := trainer.Train(numIterations, board, p1Combos, p2Combos, startPotSize)

	// Display results
	action_tree.PrintSummary(trainer, avgUtil, numIterations)
	action_tree.DisplayStrategies(trainer, len(board))
}
