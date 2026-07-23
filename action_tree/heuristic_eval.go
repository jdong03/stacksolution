package action_tree

import (
	"fmt"
	"strings"

	"github.com/jdong03/stacksolution/game"
)

// DisplayFirstActionStrategies shows all first action strategies for P1 grouped by hand
func DisplayFirstActionStrategies(trainer *VanillaCFRTrainer) {
	fmt.Println("\n========== P1 FIRST ACTION STRATEGIES ==========")

	// Group info sets by hand rank
	handGroups := map[string][]struct {
		key      string
		strategy []float64
		actions  []EnumActionType
	}{
		"AA": {},
		"KK": {},
		"QQ": {},
	}

	// Find all first-action info sets (no actions after board cards)
	for key, infoSet := range trainer.InformationSetMap {
		// Check if this is a first action info set
		if !isFirstActionInfoSet(key) {
			continue
		}

		strategy := GetFinalStrategy(infoSet)
		actions := trainer.InfoSetActionOptions[key]

		// Determine hand type from the key (first two parts are hole cards)
		parts := strings.Split(key, "_")
		if len(parts) < 2 {
			continue
		}

		// Extract rank from card strings like "As", "Kh", "Qd"
		card1Rank := getCardRankFromString(parts[0])
		card2Rank := getCardRankFromString(parts[1])

		var handType string
		if card1Rank == 14 && card2Rank == 14 {
			handType = "AA"
		} else if card1Rank == 13 && card2Rank == 13 {
			handType = "KK"
		} else if card1Rank == 12 && card2Rank == 12 {
			handType = "QQ"
		} else {
			continue
		}

		handGroups[handType] = append(handGroups[handType], struct {
			key      string
			strategy []float64
			actions  []EnumActionType
		}{key, strategy, actions})
	}

	actionNames := map[EnumActionType]string{
		Check:      "Check",
		Call:       "Call",
		Raise33:    "Bet 33%",
		Raise50:    "Bet 50%",
		Raise75:    "Bet 75%",
		Raise100:   "Bet 100%",
		RaiseAllIn: "All-In",
		Fold:       "Fold",
	}

	// Display each hand group
	for _, handType := range []string{"AA", "KK", "QQ"} {
		hands := handGroups[handType]
		if len(hands) == 0 {
			continue
		}

		fmt.Printf("\n--- %s (%d combos) ---\n", handType, len(hands))
		fmt.Printf("%-12s", "Hand")

		// Get action names from first hand's options
		if len(hands) > 0 {
			for _, action := range hands[0].actions {
				fmt.Printf("%12s", actionNames[action])
			}
		}
		fmt.Println()
		fmt.Println(strings.Repeat("-", 12+12*len(hands[0].actions)))

		// Calculate totals for averaging
		var totalStrategy []float64
		if len(hands) > 0 && len(hands[0].strategy) > 0 {
			totalStrategy = make([]float64, len(hands[0].strategy))
		}

		for _, h := range hands {
			// Format hand name from key (e.g., "As_Ah_..." -> "AsAh")
			parts := strings.Split(h.key, "_")
			handName := parts[0] + parts[1]
			fmt.Printf("%-12s", handName)

			for i, prob := range h.strategy {
				fmt.Printf("%11.1f%%", prob*100)
				if i < len(totalStrategy) {
					totalStrategy[i] += prob
				}
			}
			fmt.Println()
		}

		// Print average
		if len(hands) > 0 {
			fmt.Println(strings.Repeat("-", 12+12*len(hands[0].actions)))
			fmt.Printf("%-12s", "Average")
			for _, total := range totalStrategy {
				fmt.Printf("%11.1f%%", (total/float64(len(hands)))*100)
			}
			fmt.Println()
		}
	}

	fmt.Println("\n================================================")
}

// DisplayP2ResponseStrategies shows P2's response strategies facing P1's check or bet
func DisplayP2ResponseStrategies(trainer *VanillaCFRTrainer) {
	fmt.Println("\n========== P2 RESPONSE STRATEGIES ==========")

	// Group info sets by: facing check vs facing bet, and by hand type
	type infoSetData struct {
		key      string
		strategy []float64
		actions  []EnumActionType
	}

	facingCheck := map[string][]infoSetData{
		"AA": {},
		"KK": {},
		"QQ": {},
	}
	facingBet := map[string][]infoSetData{
		"AA": {},
		"KK": {},
		"QQ": {},
	}

	// Find P2's response info sets
	for key, infoSet := range trainer.InformationSetMap {
		parts := strings.Split(key, "_")
		if len(parts) < 3 {
			continue
		}

		// Get the last part to determine what P1 did
		lastPart := parts[len(parts)-1]

		// Check if this is P2's response to P1's first action
		// P2 responds after exactly one action from P1
		isP2Response := false
		var facingAction string

		if lastPart == "x" {
			// P2 facing a check
			isP2Response = true
			facingAction = "check"
		} else if strings.HasPrefix(lastPart, "r") {
			// P2 facing a bet/raise
			isP2Response = true
			facingAction = "bet"
		}

		if !isP2Response {
			continue
		}

		// Verify this is the first response (only one action in history)
		actionCount := 0
		for _, part := range parts[2:] { // Skip hole cards
			if part == "x" || part == "c" || part == "f" || strings.HasPrefix(part, "r") {
				actionCount++
			}
		}
		if actionCount != 1 {
			continue // Not P2's first response
		}

		strategy := GetFinalStrategy(infoSet)
		actions := trainer.InfoSetActionOptions[key]

		// Determine hand type from the key (first two parts are hole cards)
		card1Rank := getCardRankFromString(parts[0])
		card2Rank := getCardRankFromString(parts[1])

		var handType string
		if card1Rank == 14 && card2Rank == 14 {
			handType = "AA"
		} else if card1Rank == 13 && card2Rank == 13 {
			handType = "KK"
		} else if card1Rank == 12 && card2Rank == 12 {
			handType = "QQ"
		} else {
			continue
		}

		data := infoSetData{key, strategy, actions}
		if facingAction == "check" {
			facingCheck[handType] = append(facingCheck[handType], data)
		} else {
			facingBet[handType] = append(facingBet[handType], data)
		}
	}

	actionNames := map[EnumActionType]string{
		Check:      "Check",
		Call:       "Call",
		Raise33:    "Bet 33%",
		Raise50:    "Bet 50%",
		Raise75:    "Bet 75%",
		Raise100:   "Bet 100%",
		RaiseAllIn: "All-In",
		Fold:       "Fold",
	}

	// Helper to display a group of strategies
	displayGroup := func(handGroups map[string][]infoSetData) {
		for _, handType := range []string{"AA", "KK", "QQ"} {
			hands := handGroups[handType]
			if len(hands) == 0 {
				continue
			}

			fmt.Printf("\n  %s (%d combos)\n", handType, len(hands))
			fmt.Printf("  %-12s", "Hand")

			// Get action names from first hand's options
			if len(hands) > 0 {
				for _, action := range hands[0].actions {
					fmt.Printf("%12s", actionNames[action])
				}
			}
			fmt.Println()
			fmt.Println("  " + strings.Repeat("-", 10+12*len(hands[0].actions)))

			// Calculate totals for averaging
			var totalStrategy []float64
			if len(hands) > 0 && len(hands[0].strategy) > 0 {
				totalStrategy = make([]float64, len(hands[0].strategy))
			}

			for _, h := range hands {
				// Format hand name from key
				parts := strings.Split(h.key, "_")
				handName := parts[0] + parts[1]
				fmt.Printf("  %-12s", handName)

				for i, prob := range h.strategy {
					fmt.Printf("%11.1f%%", prob*100)
					if i < len(totalStrategy) {
						totalStrategy[i] += prob
					}
				}
				fmt.Println()
			}

			// Print average
			if len(hands) > 0 {
				fmt.Println("  " + strings.Repeat("-", 10+12*len(hands[0].actions)))
				fmt.Printf("  %-12s", "Average")
				for _, total := range totalStrategy {
					fmt.Printf("%11.1f%%", (total/float64(len(hands)))*100)
				}
				fmt.Println()
			}
		}
	}

	// Display P2 facing check
	fmt.Println("\n--- P2 FACING CHECK FROM P1 ---")
	displayGroup(facingCheck)

	// Display P2 facing bet
	fmt.Println("\n--- P2 FACING BET FROM P1 ---")
	displayGroup(facingBet)

	fmt.Println("\n================================================")
}

// DisplayP1ResponseToP2Bet shows P1's response after P1 check -> P2 bet
func DisplayP1ResponseToP2Bet(trainer *VanillaCFRTrainer) {
	fmt.Println("\n========== P1 RESPONSE TO P2 BET (after P1 check) ==========")

	type infoSetData struct {
		key      string
		strategy []float64
		actions  []EnumActionType
	}

	handGroups := map[string][]infoSetData{
		"AA": {},
		"KK": {},
		"QQ": {},
	}

	// Find P1's response info sets: P1 check -> P2 bet -> P1 responds
	for key, infoSet := range trainer.InformationSetMap {
		parts := strings.Split(key, "_")
		if len(parts) < 3 {
			continue
		}

		// Count actions and identify the pattern: x (check) followed by r (bet)
		actionCount := 0
		actions := []string{}
		for _, part := range parts[2:] { // Skip hole cards
			if part == "x" || part == "c" || part == "f" || strings.HasPrefix(part, "r") {
				actionCount++
				actions = append(actions, part)
			}
		}

		// We want exactly 2 actions: P1 check (x), then P2 bet (r...)
		if actionCount != 2 {
			continue
		}
		if len(actions) < 2 {
			continue
		}
		if actions[0] != "x" {
			continue // First action must be P1 check
		}
		if !strings.HasPrefix(actions[1], "r") {
			continue // Second action must be P2 bet
		}

		strategy := GetFinalStrategy(infoSet)
		actionOptions := trainer.InfoSetActionOptions[key]

		// Determine hand type
		card1Rank := getCardRankFromString(parts[0])
		card2Rank := getCardRankFromString(parts[1])

		var handType string
		if card1Rank == 14 && card2Rank == 14 {
			handType = "AA"
		} else if card1Rank == 13 && card2Rank == 13 {
			handType = "KK"
		} else if card1Rank == 12 && card2Rank == 12 {
			handType = "QQ"
		} else {
			continue
		}

		handGroups[handType] = append(handGroups[handType], infoSetData{key, strategy, actionOptions})
	}

	actionNames := map[EnumActionType]string{
		Check:      "Check",
		Call:       "Call",
		Raise33:    "Raise 33%",
		Raise50:    "Raise 50%",
		Raise75:    "Raise 75%",
		Raise100:   "Raise 100%",
		RaiseAllIn: "All-In",
		Fold:       "Fold",
	}

	fmt.Println("\n--- P1 FACING BET (after P1 checked) ---")

	for _, handType := range []string{"AA", "KK", "QQ"} {
		hands := handGroups[handType]
		if len(hands) == 0 {
			continue
		}

		fmt.Printf("\n  %s (%d combos)\n", handType, len(hands))
		fmt.Printf("  %-12s", "Hand")

		// Get action names from first hand's options
		if len(hands) > 0 {
			for _, action := range hands[0].actions {
				fmt.Printf("%12s", actionNames[action])
			}
		}
		fmt.Println()
		fmt.Println("  " + strings.Repeat("-", 10+12*len(hands[0].actions)))

		// Calculate totals for averaging
		var totalStrategy []float64
		if len(hands) > 0 && len(hands[0].strategy) > 0 {
			totalStrategy = make([]float64, len(hands[0].strategy))
		}

		for _, h := range hands {
			// Format hand name from key
			parts := strings.Split(h.key, "_")
			handName := parts[0] + parts[1]
			fmt.Printf("  %-12s", handName)

			for i, prob := range h.strategy {
				fmt.Printf("%11.1f%%", prob*100)
				if i < len(totalStrategy) {
					totalStrategy[i] += prob
				}
			}
			fmt.Println()
		}

		// Print average
		if len(hands) > 0 {
			fmt.Println("  " + strings.Repeat("-", 10+12*len(hands[0].actions)))
			fmt.Printf("  %-12s", "Average")
			for _, total := range totalStrategy {
				fmt.Printf("%11.1f%%", (total/float64(len(hands)))*100)
			}
			fmt.Println()
		}
	}

	fmt.Println("\n================================================")
}

// getCardRankFromString extracts rank from card string like "As", "Kh", "2d"
func getCardRankFromString(cardStr string) int {
	if len(cardStr) < 1 {
		return 0
	}
	switch cardStr[0] {
	case 'A':
		return 14
	case 'K':
		return 13
	case 'Q':
		return 12
	case 'J':
		return 11
	case 'T':
		return 10
	default:
		if cardStr[0] >= '2' && cardStr[0] <= '9' {
			return int(cardStr[0] - '0')
		}
		return 0
	}
}

// HeuristicStrategy defines a fixed strategy based on hand strength
type HeuristicStrategy struct {
	Name string
	// GetStrategy returns action probabilities for a given situation
	// handRank: 14=AA, 13=KK, 12=QQ
	// isFirstAction: true if first to act (no bets yet)
	// actionOptions: available actions at this node
	GetStrategy func(handRank int, isFirstAction bool, actionOptions []EnumActionType) []float64
}

// SimpleHeuristic: AA always bets, KK/QQ check-fold
func SimpleHeuristic() *HeuristicStrategy {
	return &HeuristicStrategy{
		Name: "Simple (AA bets, KK/QQ check-fold)",
		GetStrategy: func(handRank int, isFirstAction bool, actionOptions []EnumActionType) []float64 {
			strategy := make([]float64, len(actionOptions))

			// Find action indices
			checkIdx, betIdx, callIdx, foldIdx := -1, -1, -1, -1
			for i, action := range actionOptions {
				switch action {
				case Check:
					checkIdx = i
				case Raise50, Raise33, Raise75, Raise100:
					if betIdx == -1 {
						betIdx = i // Use first available bet size
					}
				case Call:
					callIdx = i
				case Fold:
					foldIdx = i
				}
			}

			switch handRank {
			case 14: // AA - always bet/call
				if isFirstAction {
					if betIdx >= 0 {
						strategy[betIdx] = 1.0
					} else if checkIdx >= 0 {
						strategy[checkIdx] = 1.0
					}
				} else {
					if callIdx >= 0 {
						strategy[callIdx] = 1.0
					} else if foldIdx >= 0 {
						strategy[foldIdx] = 1.0
					}
				}
			case 13, 12: // KK, QQ - check/fold
				if isFirstAction {
					if checkIdx >= 0 {
						strategy[checkIdx] = 1.0
					} else if foldIdx >= 0 {
						strategy[foldIdx] = 1.0
					}
				} else {
					if foldIdx >= 0 {
						strategy[foldIdx] = 1.0
					} else if callIdx >= 0 {
						strategy[callIdx] = 1.0
					}
				}
			default:
				// Default: first available action
				if len(strategy) > 0 {
					strategy[0] = 1.0
				}
			}

			return strategy
		},
	}
}

// getHandRank extracts the hand rank from hole cards (assumes pocket pair)
func getHandRank(cards []game.Card) int {
	if len(cards) >= 2 && cards[0].Rank == cards[1].Rank {
		return cards[0].Rank
	}
	if len(cards) >= 2 {
		if cards[0].Rank > cards[1].Rank {
			return cards[0].Rank
		}
		return cards[1].Rank
	}
	return 0
}

// isFirstActionOnStreet checks if this is the first action on the street
func isFirstActionOnStreet(node *PlayerNode) bool {
	history := node.GetGameState().History
	return len(history.RiverActions) == 0
}

// EvaluateVsHeuristic computes expected values when solver plays against heuristic
func EvaluateVsHeuristic(
	trainer *VanillaCFRTrainer,
	board []game.Card,
	handCombosP1 [][]game.Card,
	handCombosP2 [][]game.Card,
	initialPotSize float64,
	heuristic *HeuristicStrategy,
) {
	fmt.Printf("\n========== SOLVER vs HEURISTIC: %s ==========\n", heuristic.Name)

	// P1 (solver) vs P2 (heuristic)
	p1SolverEV := computeEV(trainer, board, handCombosP1, handCombosP2, initialPotSize, heuristic, true)
	fmt.Printf("P1 (Solver) vs P2 (Heuristic): P1 EV = %.4f BB\n", p1SolverEV)

	// P1 (heuristic) vs P2 (solver)
	p1HeuristicEV := computeEV(trainer, board, handCombosP1, handCombosP2, initialPotSize, heuristic, false)
	fmt.Printf("P1 (Heuristic) vs P2 (Solver): P1 EV = %.4f BB\n", p1HeuristicEV)

	// Solver vs Solver (baseline)
	p1SolverVsSolverEV := trainer.BestResponseUtility.ComputeAverageP1Utility(board, handCombosP1, handCombosP2, initialPotSize)
	fmt.Printf("P1 (Solver) vs P2 (Solver): P1 EV = %.4f BB\n", p1SolverVsSolverEV)

	fmt.Println("================================================")
}

func computeEV(
	trainer *VanillaCFRTrainer,
	board []game.Card,
	handCombosP1 [][]game.Card,
	handCombosP2 [][]game.Card,
	initialPotSize float64,
	heuristic *HeuristicStrategy,
	p1UsesSolver bool,
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
			startNode := trainer.createStartingNodeWithBoard(board, p1Hand, p2Hand, initialPotSize)
			totalUtil += evalNode(trainer, startNode, heuristic, p1UsesSolver)
			count++
		}
	}

	if count == 0 {
		return 0
	}
	return totalUtil / float64(count)
}

func evalNode(trainer *VanillaCFRTrainer, node GameStateNode, heuristic *HeuristicStrategy, p1UsesSolver bool) float64 {
	switch n := node.(type) {
	case *LeafNode:
		return leafUtility(n.GetGameState(), trainer.Player1InitialStackSize)

	case *ChanceNode:
		availableCards := n.AvailableCards
		if len(availableCards) == 0 {
			return 0.0
		}
		actionProbability := 1.0 / float64(len(availableCards))
		var nodeUtility float64
		for _, card := range availableCards {
			chanceAction := ChanceAction{RevealedCards: []game.Card{card}}
			childNode := NewGameStateNode(n, chanceAction, actionProbability)
			nodeUtility += actionProbability * evalNode(trainer, childNode, heuristic, p1UsesSolver)
		}
		return nodeUtility

	case *PlayerNode:
		gameState := n.GetGameState()
		activePlayer := gameState.History.ActivePlayer

		var strategy []float64

		if (activePlayer == Player1 && p1UsesSolver) || (activePlayer == Player2 && !p1UsesSolver) {
			// This player uses solver strategy
			infoSet := trainer.GetInformationSet(n)
			strategy = GetFinalStrategy(infoSet)
		} else {
			// This player uses heuristic strategy
			var cards []game.Card
			if activePlayer == Player1 {
				cards = gameState.Player1Cards
			} else {
				cards = gameState.Player2Cards
			}
			handRank := getHandRank(cards)
			isFirst := isFirstActionOnStreet(n)
			strategy = heuristic.GetStrategy(handRank, isFirst, n.ActionOptions)
		}

		var nodeValue float64
		for i, actionType := range n.ActionOptions {
			actionProb := strategy[i]
			if actionProb == 0 {
				continue
			}
			action := trainer.createActionForType(actionType, n)
			childNode := NewGameStateNode(n, action, actionProb)
			nodeValue += actionProb * evalNode(trainer, childNode, heuristic, p1UsesSolver)
		}
		return nodeValue

	default:
		panic("Unknown node type")
	}
}
