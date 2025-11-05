package utils

func Normalize(values []float64) []float64 {
	total := 0.0
	for _, value := range values {
			total += value
	}

	normalized := make([]float64, len(values))
	if total > 0 {
			for i, value := range values {
					normalized[i] = value / total
			}
	}
	return normalized
}