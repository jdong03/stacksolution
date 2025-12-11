package action_tree

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
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
		displayInfoSets(flopInfoSets, trainer.InfoSetActionOptions)
	}

	// Display turn strategies
	if len(turnInfoSets) > 0 {
		fmt.Println("\n=== TURN STRATEGIES ===")
		displayInfoSets(turnInfoSets, trainer.InfoSetActionOptions)
	}

	// Display river strategies
	if len(riverInfoSets) > 0 {
		fmt.Println("\n=== RIVER STRATEGIES ===")
		displayInfoSets(riverInfoSets, trainer.InfoSetActionOptions)
	}
}

// displayInfoSets prints strategies for a group of information sets
func displayInfoSets(infoSets map[string]*InformationSet, actionOptionsMap map[string][]EnumActionType) {
	// Sort keys for consistent output
	keys := make([]string, 0, len(infoSets))
	for key := range infoSets {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	actionNames := map[EnumActionType]string{
		Check:      "Check",
		Call:       "Call",
		Raise33:    "Raise33%",
		Raise50:    "Raise50%",
		Raise75:    "Raise75%",
		Raise100:   "Raise100%",
		RaiseAllIn: "AllIn",
		Fold:       "Fold",
	}

	for _, key := range keys {
		infoSet := infoSets[key]
		strategy := GetFinalStrategy(infoSet)
		actionOptions := actionOptionsMap[key]

		fmt.Printf("\nInfo Set: %s\n", formatInfoSetKey(key))
		fmt.Println("Strategy:")

		// Display each action's probability using the actual action options
		for i, prob := range strategy {
			if prob > 0.001 { // Only show actions with >0.1% probability
				actionName := "Unknown"
				if i < len(actionOptions) {
					actionName = actionNames[actionOptions[i]]
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
			(part[0] >= '0' && part[0] <= '9' || part[0] == '1' ||
				part[0] == 'A' || part[0] == 'K' || part[0] == 'Q' || part[0] == 'J') {
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
	case strings.HasPrefix(actionCode, "rAllIn"):
		// Format: rAllIn100.00 -> AllIn(100.00)
		return "AllIn(" + actionCode[6:] + ")"
	case strings.HasPrefix(actionCode, "r"):
		// Format: r25.00|50 -> raise(50%, 25.00)
		rest := actionCode[1:]
		parts := strings.Split(rest, "|")
		if len(parts) == 2 {
			return fmt.Sprintf("raise(%s%%, %s)", parts[1], parts[0])
		}
		return "raise(" + rest + ")"
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

// WriteStrategiesToFile writes the strategies to a readable text file
func WriteStrategiesToFile(trainer *VanillaCFRTrainer, numBoardCards int, numIterations int, avgUtility float64, filename string) error {
	// If no filename provided, generate one with timestamp
	if filename == "" {
		filename = fmt.Sprintf("strategies_%s.txt", time.Now().Format("2006-01-02_15-04-05"))
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write header
	fmt.Fprintf(file, "╔══════════════════════════════════════════════════════════════╗\n")
	fmt.Fprintf(file, "║                    CFR SOLVER RESULTS                        ║\n")
	fmt.Fprintf(file, "╚══════════════════════════════════════════════════════════════╝\n\n")

	// Write summary
	fmt.Fprintf(file, "Training Summary:\n")
	fmt.Fprintf(file, "─────────────────\n")
	fmt.Fprintf(file, "  Iterations:           %d\n", numIterations)
	fmt.Fprintf(file, "  P1 Avg Utility:       %.4f BB\n", avgUtility)
	fmt.Fprintf(file, "  Information Sets:     %d\n", len(trainer.InformationSetMap))
	fmt.Fprintf(file, "  Board Cards:          %d\n", numBoardCards)
	fmt.Fprintf(file, "  Generated:            %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	// Group info sets by street
	flopInfoSets := make(map[string]*InformationSet)
	turnInfoSets := make(map[string]*InformationSet)
	riverInfoSets := make(map[string]*InformationSet)

	for key, infoSet := range trainer.InformationSetMap {
		if numBoardCards >= 5 {
			riverInfoSets[key] = infoSet
		} else if numBoardCards >= 4 {
			turnInfoSets[key] = infoSet
		} else {
			flopInfoSets[key] = infoSet
		}
	}

	// Write strategies by street
	if len(flopInfoSets) > 0 {
		fmt.Fprintf(file, "╔══════════════════════════════════════════════════════════════╗\n")
		fmt.Fprintf(file, "║                      FLOP STRATEGIES                         ║\n")
		fmt.Fprintf(file, "╚══════════════════════════════════════════════════════════════╝\n\n")
		writeInfoSetsToFile(file, flopInfoSets, trainer.InfoSetActionOptions)
	}

	if len(turnInfoSets) > 0 {
		fmt.Fprintf(file, "\n╔══════════════════════════════════════════════════════════════╗\n")
		fmt.Fprintf(file, "║                      TURN STRATEGIES                         ║\n")
		fmt.Fprintf(file, "╚══════════════════════════════════════════════════════════════╝\n\n")
		writeInfoSetsToFile(file, turnInfoSets, trainer.InfoSetActionOptions)
	}

	if len(riverInfoSets) > 0 {
		fmt.Fprintf(file, "\n╔══════════════════════════════════════════════════════════════╗\n")
		fmt.Fprintf(file, "║                      RIVER STRATEGIES                        ║\n")
		fmt.Fprintf(file, "╚══════════════════════════════════════════════════════════════╝\n\n")
		writeInfoSetsToFile(file, riverInfoSets, trainer.InfoSetActionOptions)
	}

	fmt.Printf("Strategies written to: %s\n", filename)
	return nil
}

// writeInfoSetsToFile writes a group of info sets to a file
func writeInfoSetsToFile(file *os.File, infoSets map[string]*InformationSet, actionOptionsMap map[string][]EnumActionType) {
	keys := make([]string, 0, len(infoSets))
	for key := range infoSets {
		keys = append(keys, key)
	}
	sort.Strings(keys)

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

	for _, key := range keys {
		infoSet := infoSets[key]
		strategy := GetFinalStrategy(infoSet)
		actionOptions := actionOptionsMap[key]

		fmt.Fprintf(file, "┌─────────────────────────────────────────────────────────────┐\n")
		fmt.Fprintf(file, "│ %s\n", formatInfoSetKey(key))
		fmt.Fprintf(file, "├─────────────────────────────────────────────────────────────┤\n")

		for i, prob := range strategy {
			if prob > 0.001 {
				actionName := "Unknown"
				if i < len(actionOptions) {
					actionName = actionNames[actionOptions[i]]
				}
				// Create a simple bar chart
				barLength := int(prob * 30)
				bar := strings.Repeat("█", barLength) + strings.Repeat("░", 30-barLength)
				fmt.Fprintf(file, "│  %-12s %s %5.1f%%\n", actionName, bar, prob*100)
			}
		}
		fmt.Fprintf(file, "└─────────────────────────────────────────────────────────────┘\n\n")
	}
}
