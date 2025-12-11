package action_tree

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// TrainingMetrics tracks data over iterations for analysis and graphing
type TrainingMetrics struct {
	// Exploitability per iteration
	Exploitability []float64

	// Strategy tracking: map[infoSetKey] -> []iterationData
	// Each iteration data contains strategy probabilities
	CurrentStrategyHistory map[string][][]float64
	AverageStrategyHistory map[string][][]float64

	// Info sets to track (populated by user)
	TrackedInfoSets []string

	// Only track first action (no prior actions on current street)
	FirstActionOnly bool
}

// NewTrainingMetrics creates a new metrics tracker
func NewTrainingMetrics() *TrainingMetrics {
	return &TrainingMetrics{
		Exploitability:         make([]float64, 0),
		CurrentStrategyHistory: make(map[string][][]float64),
		AverageStrategyHistory: make(map[string][][]float64),
		TrackedInfoSets:        make([]string, 0),
		FirstActionOnly:        false,
	}
}

// isFirstActionInfoSet checks if the info set key represents the first action
// on the current street (no action codes like x, c, f, r after the last card)
func isFirstActionInfoSet(key string) bool {
	parts := strings.Split(key, "_")
	if len(parts) < 2 {
		return false
	}

	// The last part should be a card (not an action)
	// Cards look like: 2d, Ks, Ah, etc.
	// Actions look like: x, c, f, r25.00|50, rAllIn100.00
	lastPart := parts[len(parts)-1]

	// If it's an action code, this is not a first action info set
	if lastPart == "x" || lastPart == "c" || lastPart == "f" {
		return false
	}
	if strings.HasPrefix(lastPart, "r") {
		return false
	}

	// It's a card (first action on this street)
	return true
}

// TrackInfoSet adds an info set key pattern to track
// Pattern can be partial (e.g., "12s_12h" to track QQ hands)
func (m *TrainingMetrics) TrackInfoSet(pattern string) {
	m.TrackedInfoSets = append(m.TrackedInfoSets, pattern)
}

// RecordIteration records metrics for one iteration
func (m *TrainingMetrics) RecordIteration(trainer *VanillaCFRTrainer, exploitability float64) {
	m.Exploitability = append(m.Exploitability, exploitability)

	// Record strategies for tracked info sets
	for key, infoSet := range trainer.InformationSetMap {
		// Skip if FirstActionOnly is set and this isn't a first action
		if m.FirstActionOnly && !isFirstActionInfoSet(key) {
			continue
		}

		// Check if this key matches any tracked pattern
		for _, pattern := range m.TrackedInfoSets {
			if strings.Contains(key, pattern) {
				// Record current strategy
				currentStrategy := GetStrategy(infoSet)
				strategyCopy := make([]float64, len(currentStrategy))
				copy(strategyCopy, currentStrategy)

				if _, exists := m.CurrentStrategyHistory[key]; !exists {
					m.CurrentStrategyHistory[key] = make([][]float64, 0)
				}
				m.CurrentStrategyHistory[key] = append(m.CurrentStrategyHistory[key], strategyCopy)

				// Record average strategy
				avgStrategy := GetFinalStrategy(infoSet)
				avgStrategyCopy := make([]float64, len(avgStrategy))
				copy(avgStrategyCopy, avgStrategy)

				if _, exists := m.AverageStrategyHistory[key]; !exists {
					m.AverageStrategyHistory[key] = make([][]float64, 0)
				}
				m.AverageStrategyHistory[key] = append(m.AverageStrategyHistory[key], avgStrategyCopy)
				break
			}
		}
	}
}

// WriteExploitabilityCSV writes exploitability data to CSV
func (m *TrainingMetrics) WriteExploitabilityCSV(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header
	writer.Write([]string{"iteration", "exploitability"})

	// Data
	for i, exp := range m.Exploitability {
		writer.Write([]string{
			strconv.Itoa(i),
			fmt.Sprintf("%.6f", exp),
		})
	}

	fmt.Printf("Exploitability data written to: %s\n", filename)
	return nil
}

// WriteStrategyCSV writes strategy evolution data to CSV for a specific info set
func (m *TrainingMetrics) WriteStrategyCSV(infoSetKey string, actionOptions []EnumActionType, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	actionNames := map[EnumActionType]string{
		Check:      "Check",
		Call:       "Call",
		Raise33:    "Raise33",
		Raise50:    "Raise50",
		Raise75:    "Raise75",
		Raise100:   "Raise100",
		RaiseAllIn: "AllIn",
		Fold:       "Fold",
	}

	// Build header
	header := []string{"iteration"}
	for _, action := range actionOptions {
		header = append(header, "current_"+actionNames[action])
	}
	for _, action := range actionOptions {
		header = append(header, "average_"+actionNames[action])
	}
	writer.Write(header)

	// Get data for this info set
	currentHistory := m.CurrentStrategyHistory[infoSetKey]
	avgHistory := m.AverageStrategyHistory[infoSetKey]

	numIterations := len(currentHistory)
	if len(avgHistory) < numIterations {
		numIterations = len(avgHistory)
	}

	for i := 0; i < numIterations; i++ {
		row := []string{strconv.Itoa(i)}

		// Current strategy values
		for j := range actionOptions {
			if j < len(currentHistory[i]) {
				row = append(row, fmt.Sprintf("%.6f", currentHistory[i][j]))
			} else {
				row = append(row, "0")
			}
		}

		// Average strategy values
		for j := range actionOptions {
			if j < len(avgHistory[i]) {
				row = append(row, fmt.Sprintf("%.6f", avgHistory[i][j]))
			} else {
				row = append(row, "0")
			}
		}

		writer.Write(row)
	}

	fmt.Printf("Strategy data for %s written to: %s\n", infoSetKey, filename)
	return nil
}

// WriteAllTrackedStrategiesCSV writes all tracked info sets to separate CSV files
func (m *TrainingMetrics) WriteAllTrackedStrategiesCSV(trainer *VanillaCFRTrainer, prefix string) {
	for key := range m.CurrentStrategyHistory {
		actionOptions := trainer.InfoSetActionOptions[key]
		safeKey := strings.ReplaceAll(key, "|", "_")
		safeKey = strings.ReplaceAll(safeKey, ".", "_")
		filename := fmt.Sprintf("%s_strategy_%s.csv", prefix, safeKey)
		m.WriteStrategyCSV(key, actionOptions, filename)
	}
}

// GetBetFrequency calculates the betting frequency (non-check/non-fold actions) from a strategy
func GetBetFrequency(strategy []float64, actionOptions []EnumActionType) float64 {
	betFreq := 0.0
	for i, prob := range strategy {
		if i < len(actionOptions) {
			action := actionOptions[i]
			if action != Check && action != Fold && action != Call {
				betFreq += prob
			}
		}
	}
	return betFreq
}

// WriteBetFrequencyCSV writes bet frequency evolution to CSV
func (m *TrainingMetrics) WriteBetFrequencyCSV(trainer *VanillaCFRTrainer, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Build header dynamically based on tracked info sets
	header := []string{"iteration"}
	infoSetKeys := make([]string, 0)
	for key := range m.CurrentStrategyHistory {
		infoSetKeys = append(infoSetKeys, key)
		shortKey := shortenInfoSetKey(key)
		header = append(header, "current_bet_freq_"+shortKey)
		header = append(header, "average_bet_freq_"+shortKey)
	}
	writer.Write(header)

	// Find max iterations
	maxIter := 0
	for _, history := range m.CurrentStrategyHistory {
		if len(history) > maxIter {
			maxIter = len(history)
		}
	}

	// Write data
	for i := 0; i < maxIter; i++ {
		row := []string{strconv.Itoa(i)}

		for _, key := range infoSetKeys {
			actionOptions := trainer.InfoSetActionOptions[key]

			// Current bet frequency
			if i < len(m.CurrentStrategyHistory[key]) {
				betFreq := GetBetFrequency(m.CurrentStrategyHistory[key][i], actionOptions)
				row = append(row, fmt.Sprintf("%.6f", betFreq))
			} else {
				row = append(row, "")
			}

			// Average bet frequency
			if i < len(m.AverageStrategyHistory[key]) {
				betFreq := GetBetFrequency(m.AverageStrategyHistory[key][i], actionOptions)
				row = append(row, fmt.Sprintf("%.6f", betFreq))
			} else {
				row = append(row, "")
			}
		}

		writer.Write(row)
	}

	fmt.Printf("Bet frequency data written to: %s\n", filename)
	return nil
}

// shortenInfoSetKey creates a shorter readable key for headers
func shortenInfoSetKey(key string) string {
	parts := strings.Split(key, "_")
	if len(parts) >= 2 {
		return parts[0] + parts[1] // Just hole cards
	}
	return key
}

// FindInfoSetsByPattern returns all info set keys that match a pattern
func (trainer *VanillaCFRTrainer) FindInfoSetsByPattern(pattern string) []string {
	matches := make([]string, 0)
	for key := range trainer.InformationSetMap {
		if strings.Contains(key, pattern) {
			matches = append(matches, key)
		}
	}
	return matches
}
