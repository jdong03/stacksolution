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
	} else if len(values) > 0 {
		// Return uniform distribution when total is zero
		uniform := 1.0 / float64(len(values))
		for i := range normalized {
			normalized[i] = uniform
		}
	}
	return normalized
}