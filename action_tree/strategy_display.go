package action_tree

import (
	"fmt"
	"sort"
	"strings"
)

// DisplayStrategies prints the final average strategies for all information sets
func DisplayStrategies(trainer *VanillaCFRTrainer, numBoardCards int) {
	fmt.Println("\n========== FINAL STRATEGIES ==========")

	// Group info sets by street
	flopInfoSets := make(map[string]*InformationSet)
	turnInfoSets := make(map[string]*InformationSet)
	riverInfoSets := make(map[string]*InformationSet)

	for key, infoSet := range trainer.InformationSetMap {
		// Count underscores to determine street
		// Format: card1_card2_boardcards_actions

		if numBoardCards >= 5 {
			riverInfoSets[key] = infoSet
		} else if numBoardCards >= 4 {
			turnInfoSets[key] = infoSet
		} else {
			flopInfoSets[key] = infoSet
		}
	}

	// Display flop strategies
	if len(flopInfoSets) > 0 {
		fmt.Println("=== FLOP STRATEGIES ===")
		displayInfoSets(flopInfoSets)
	}

	// Display turn strategies
	if len(turnInfoSets) > 0 {
		fmt.Println("\n=== TURN STRATEGIES ===")
		displayInfoSets(turnInfoSets)
	}

	// Display river strategies
	if len(riverInfoSets) > 0 {
		fmt.Println("\n=== RIVER STRATEGIES ===")
		displayInfoSets(riverInfoSets)
	}
}

// displayInfoSets prints strategies for a group of information sets
func displayInfoSets(infoSets map[string]*InformationSet) {
	// Sort keys for consistent output
	keys := make([]string, 0, len(infoSets))
	for key := range infoSets {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		infoSet := infoSets[key]
		strategy := GetFinalStrategy(infoSet)

		fmt.Printf("\nInfo Set: %s\n", formatInfoSetKey(key))
		fmt.Println("Strategy:")

		// Display each action's probability
		actionNames := []string{"Check", "Call", "Raise33", "Raise50", "Raise75", "Raise100", "RaiseAllIn", "Fold"}
		for i, prob := range strategy {
			if prob > 0.001 { // Only show actions with >0.1% probability
				actionName := "Unknown"
				if i < len(actionNames) {
					actionName = actionNames[i]
				}
				fmt.Printf("  %s: %.1f%%\n", actionName, prob*100)
			}
		}
	}
}

// formatInfoSetKey makes the info set key more readable
// Converts "14s_14h_2d_2s_2c_2h_3d_x" to something like "AsAh | Board: 2d2s2c2h3d | History: check"
func formatInfoSetKey(key string) string {
	parts := strings.Split(key, "_")
	if len(parts) < 2 {
		return key
	}

	// First two parts are hole cards
	holeCards := fmt.Sprintf("%s%s", parts[0], parts[1])

	// Remaining parts are board cards and actions
	var boardCards []string
	var actions []string

	for i := 2; i < len(parts); i++ {
		part := parts[i]
		// If it looks like a card (contains rank and suit), it's a board card
		if len(part) >= 2 && (strings.Contains(part, "s") || strings.Contains(part, "h") ||
			strings.Contains(part, "d") || strings.Contains(part, "c")) &&
			(part[0] >= '0' && part[0] <= '9' || part[0] == '1') {
			boardCards = append(boardCards, part)
		} else {
			// Otherwise it's an action
			actions = append(actions, formatAction(part))
		}
	}

	result := fmt.Sprintf("Hand: %s", holeCards)
	if len(boardCards) > 0 {
		result += fmt.Sprintf(" | Board: %s", strings.Join(boardCards, " "))
	}
	if len(actions) > 0 {
		result += fmt.Sprintf(" | History: %s", strings.Join(actions, ", "))
	}

	return result
}

// formatAction converts action codes to readable names
func formatAction(actionCode string) string {
	switch {
	case actionCode == "x":
		return "check"
	case actionCode == "c":
		return "call"
	case actionCode == "f":
		return "fold"
	case strings.HasPrefix(actionCode, "r"):
		return "raise(" + actionCode[1:] + ")"
	default:
		return actionCode
	}
}

// PrintSummary prints a summary of the training results
func PrintSummary(trainer *VanillaCFRTrainer, avgUtility float64, numIterations int) {
	fmt.Println("\n========== TRAINING SUMMARY ==========")
	fmt.Printf("Iterations: %d\n", numIterations)
	fmt.Printf("Player 1 Average Utility (last iteration): %.4f BB\n", avgUtility)
	fmt.Printf("Information Sets Created: %d\n", len(trainer.InformationSetMap))
	fmt.Println("\n======================================")
}
