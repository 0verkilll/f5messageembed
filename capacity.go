package f5messageembed

// CalculateCapacity analyzes a coefficient array and returns capacity information.
//
// This function counts total and usable coefficients, calculates the embedding
// capacity for each k value (1-8), and estimates the shrinkage factor based
// on the distribution of coefficient magnitudes.
//
// Usable coefficients are those that can carry steganographic data:
//   - Non-DC coefficients (index % 64 != 0)
//   - Non-zero coefficients
//
// The capacity for each k is calculated as:
//
//	capacityBits = (usableCount / n) * k - HeaderSize
//	capacityBytes = capacityBits / 8
//
// where n = 2^k - 1 is the code word length.
//
// The estimated shrinkage factor is the proportion of usable coefficients
// with absolute value 1, which will cause shrinkage when modified.
//
// Parameters:
//   - coefficients: The JPEG DCT coefficients to analyze
//
// Returns:
//   - A CapacityResult struct containing:
//   - TotalCoefficients: Total count of coefficients
//   - UsableCoefficients: Count of non-zero, non-DC coefficients
//   - CapacityByK: Map of k values (1-8) to capacity in bytes
//   - EstimatedShrinkageFactor: Proportion of |1| coefficients (0.0-1.0)
//
// Example:
//
//	result := CalculateCapacity(coefficients)
//	fmt.Printf("Total: %d, Usable: %d\n", result.TotalCoefficients, result.UsableCoefficients)
//	fmt.Printf("Capacity at k=4: %d bytes\n", result.CapacityByK[4])
//	fmt.Printf("Expected shrinkage: %.1f%%\n", result.EstimatedShrinkageFactor*100)
func CalculateCapacity(coefficients []int16) *CapacityResult {
	result := &CapacityResult{
		TotalCoefficients:        len(coefficients),
		UsableCoefficients:       0,
		CapacityByK:              make(map[int]int, 8),
		EstimatedShrinkageFactor: 0.0,
		MagnitudeOneCount:        0,
	}

	// Count usable coefficients and coefficients with magnitude 1 in a single
	// pass. MagnitudeOneCount = h(1), the exact integer count of |c|=1
	// coefficients (= shrinkage-eligible pool). Callers should prefer this
	// field over reconstructing the count from EstimatedShrinkageFactor,
	// which goes through float64 and can lose ±1 on the round trip.
	for i, coeff := range coefficients {
		if IsUsableCoefficient(i, coeff) {
			result.UsableCoefficients++
			if coeff == 1 || coeff == -1 {
				result.MagnitudeOneCount++
			}
		}
	}

	// Calculate estimated shrinkage factor (kept for API compatibility).
	if result.UsableCoefficients > 0 {
		result.EstimatedShrinkageFactor = float64(result.MagnitudeOneCount) / float64(result.UsableCoefficients)
	}

	// Calculate capacity for each k value (1-8)
	for k := 1; k <= 8; k++ {
		capacityBytes := calculateCapacityBytes(result.UsableCoefficients, k)
		result.CapacityByK[k] = capacityBytes
	}

	return result
}

// calculateCapacityBytes computes the message capacity in bytes for a given k value.
//
// The formula is:
//
//	n = 2^k - 1 (code word length)
//	codeWords = usableCount / n
//	totalCapacityBits = codeWords * k
//	messageCapacityBits = totalCapacityBits - HeaderSize
//	messageCapacityBytes = messageCapacityBits / 8
//
// If the capacity is negative (not enough for header), returns 0.
func calculateCapacityBytes(usableCoeffCount, k int) int {
	if k <= 0 || k > 8 {
		return 0
	}

	// n = 2^k - 1 (code word length)
	n := (1 << k) - 1

	// Calculate number of complete code words
	codeWords := usableCoeffCount / n

	// Total capacity in bits
	totalCapacityBits := codeWords * k

	// Subtract header overhead (32 bits)
	messageCapacityBits := totalCapacityBits - HeaderSize

	// If capacity is negative, return 0
	if messageCapacityBits < 0 {
		return 0
	}

	// Convert to bytes (integer division truncates)
	return messageCapacityBits / 8
}
