package main

import (
	"fmt"

	"github.com/0verkilll/f5messageembed"
)

func main() {
	fmt.Println("F5 Capacity Calculation Example")
	fmt.Println("================================")
	fmt.Println()

	// Example 1: Small coefficient array
	fmt.Println("1. Small Coefficient Array (1,000 coefficients)")
	fmt.Println("------------------------------------------------")
	smallCoeffs := generateSyntheticCoefficients(1000)
	analyzeCapacity(smallCoeffs)
	fmt.Println()

	// Example 2: Medium coefficient array
	fmt.Println("2. Medium Coefficient Array (10,000 coefficients)")
	fmt.Println("-------------------------------------------------")
	mediumCoeffs := generateSyntheticCoefficients(10000)
	analyzeCapacity(mediumCoeffs)
	fmt.Println()

	// Example 3: Large coefficient array
	fmt.Println("3. Large Coefficient Array (100,000 coefficients)")
	fmt.Println("-------------------------------------------------")
	largeCoeffs := generateSyntheticCoefficients(100000)
	analyzeCapacity(largeCoeffs)
	fmt.Println()

	// Example 4: Sparse coefficients (many zeros)
	fmt.Println("4. Sparse Coefficient Array (50% zeros)")
	fmt.Println("---------------------------------------")
	sparseCoeffs := generateSparseCoefficients(10000)
	analyzeCapacity(sparseCoeffs)
	fmt.Println()

	// Example 5: Dense coefficients (few zeros)
	fmt.Println("5. Dense Coefficient Array (10% zeros)")
	fmt.Println("--------------------------------------")
	denseCoeffs := generateDenseCoefficients(10000)
	analyzeCapacity(denseCoeffs)
	fmt.Println()

	// Example 6: Demonstrating message size validation
	fmt.Println("6. Message Size Validation")
	fmt.Println("--------------------------")
	demonstrateValidation(mediumCoeffs)
}

// analyzeCapacity performs and displays capacity analysis.
func analyzeCapacity(coefficients []int16) {
	result := f5messageembed.CalculateCapacity(coefficients)

	fmt.Printf("Total coefficients:      %d\n", result.TotalCoefficients)
	fmt.Printf("Usable coefficients:     %d (%.1f%%)\n",
		result.UsableCoefficients,
		float64(result.UsableCoefficients)*100/float64(result.TotalCoefficients))
	fmt.Printf("Est. shrinkage factor:   %.1f%%\n", result.EstimatedShrinkageFactor*100)
	fmt.Println()

	fmt.Println("Capacity by k parameter:")
	fmt.Println("  k | n (code word) | Capacity (bytes) | Efficiency")
	fmt.Println("  --|---------------|------------------|----------")
	for k := 1; k <= 8; k++ {
		n := (1 << k) - 1
		capacity := result.CapacityByK[k]
		efficiency := float64(k) / float64(n)
		fmt.Printf("  %d | %13d | %16d | %.3f\n", k, n, capacity, efficiency)
	}
}

// generateSyntheticCoefficients creates coefficients with realistic JPEG distribution.
func generateSyntheticCoefficients(count int) []int16 {
	coefficients := make([]int16, count)

	for i := range coefficients {
		if i%64 == 0 {
			coefficients[i] = int16(100 + (i % 200))
			continue
		}

		switch {
		case i%7 == 0:
			coefficients[i] = 0
		case i%11 == 0:
			coefficients[i] = int16(10 + (i % 50))
		case i%13 == 0:
			coefficients[i] = int16(-(10 + (i % 50)))
		case i%5 == 0:
			coefficients[i] = int16(1 + (i % 3))
		case i%3 == 0:
			coefficients[i] = int16(-(1 + (i % 3)))
		default:
			if i%2 == 0 {
				coefficients[i] = int16(2 + (i % 10))
			} else {
				coefficients[i] = int16(-(2 + (i % 10)))
			}
		}
	}

	return coefficients
}

// generateSparseCoefficients creates coefficients with many zeros (50%).
func generateSparseCoefficients(count int) []int16 {
	coefficients := make([]int16, count)

	for i := range coefficients {
		if i%64 == 0 {
			coefficients[i] = int16(100 + (i % 200))
			continue
		}

		// 50% zeros
		if i%2 == 0 {
			coefficients[i] = 0
		} else {
			coefficients[i] = int16(1 + (i % 10))
		}
	}

	return coefficients
}

// generateDenseCoefficients creates coefficients with few zeros (10%).
func generateDenseCoefficients(count int) []int16 {
	coefficients := make([]int16, count)

	for i := range coefficients {
		if i%64 == 0 {
			coefficients[i] = int16(100 + (i % 200))
			continue
		}

		// Only 10% zeros
		if i%10 == 0 {
			coefficients[i] = 0
		} else {
			val := int16(1 + (i % 20))
			if i%2 == 0 {
				coefficients[i] = val
			} else {
				coefficients[i] = -val
			}
		}
	}

	return coefficients
}

// demonstrateValidation shows how to check if a message fits.
func demonstrateValidation(coefficients []int16) {
	capacity := f5messageembed.CalculateCapacity(coefficients)

	// Find optimal k for different message sizes
	testSizes := []int{10, 100, 500, 1000, 5000}

	fmt.Println("Message size validation:")
	fmt.Println("  Size (bytes) | Fits at k | Max capacity at k")
	fmt.Println("  -------------|-----------|------------------")

	for _, size := range testSizes {
		// Find the best k that can fit this message
		bestK := 0
		for k := 8; k >= 1; k-- {
			if capacity.CapacityByK[k] >= size {
				bestK = k
				break
			}
		}

		if bestK > 0 {
			fmt.Printf("  %12d | %9d | %d bytes\n", size, bestK, capacity.CapacityByK[bestK])
		} else {
			fmt.Printf("  %12d | Too large | Max: %d bytes at k=1\n", size, capacity.CapacityByK[1])
		}
	}
}
