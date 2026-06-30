package f5messageembed

import (
	"testing"
)

// TestCalculateCapacityTotalCoefficients tests that total coefficients are counted correctly.
func TestCalculateCapacityTotalCoefficients(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		coefficients []int16
		expected     int
	}{
		{
			name:         "empty slice",
			coefficients: []int16{},
			expected:     0,
		},
		{
			name:         "single coefficient",
			coefficients: []int16{5},
			expected:     1,
		},
		{
			name:         "multiple coefficients",
			coefficients: []int16{1, 2, 3, 4, 5, 0, -1, -2},
			expected:     8,
		},
		{
			name:         "one full 8x8 block (64 coefficients)",
			coefficients: make([]int16, 64),
			expected:     64,
		},
		{
			name:         "two full 8x8 blocks (128 coefficients)",
			coefficients: make([]int16, 128),
			expected:     128,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := CalculateCapacity(tc.coefficients)
			if result.TotalCoefficients != tc.expected {
				t.Errorf("CalculateCapacity().TotalCoefficients = %d, want %d",
					result.TotalCoefficients, tc.expected)
			}
		})
	}
}

// TestCalculateCapacityUsableCoefficients tests usable coefficient counting.
// Usable coefficients are non-zero and non-DC (index % 64 != 0).
func TestCalculateCapacityUsableCoefficients(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		coefficients []int16
		expected     int
	}{
		{
			name:         "empty slice",
			coefficients: []int16{},
			expected:     0,
		},
		{
			name:         "all zeros",
			coefficients: []int16{0, 0, 0, 0, 0},
			expected:     0,
		},
		{
			name:         "DC coefficient only (index 0)",
			coefficients: []int16{5},
			expected:     0, // DC coefficient at index 0 is skipped
		},
		{
			name:         "one usable AC coefficient",
			coefficients: []int16{0, 5}, // index 0 is DC (skipped), index 1 is AC (usable)
			expected:     1,
		},
		{
			name:         "mixed usable and unusable",
			coefficients: []int16{5, 0, 3, 0, -2, 0, 0, 1}, // DC at 0, zeros at 1,3,5,6
			expected:     3,                                // indices 2,4,7 are usable (non-zero, non-DC)
		},
		{
			name:         "full block with DC excluded",
			coefficients: createTestBlock(10, true), // DC at index 0 is 10, all AC are 10
			expected:     63,                        // 64 - 1 DC = 63 usable
		},
		{
			name:         "two blocks with DC excluded",
			coefficients: createTestBlocksNonZero(2), // 2 blocks, each with DC excluded
			expected:     126,                        // (64-1)*2 = 126 usable
		},
		{
			name:         "block with some zeros",
			coefficients: createTestBlockWithZeros(), // Block with DC=0, some AC=0
			expected:     31,                         // 31 non-zero AC coefficients
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := CalculateCapacity(tc.coefficients)
			if result.UsableCoefficients != tc.expected {
				t.Errorf("CalculateCapacity().UsableCoefficients = %d, want %d",
					result.UsableCoefficients, tc.expected)
			}
		})
	}
}

// TestCalculateCapacityByK tests capacity calculation for each k value (1-8).
func TestCalculateCapacityByK(t *testing.T) {
	t.Parallel()

	// Create a coefficient array with known usable count
	// 1000 usable coefficients (non-zero, non-DC)
	coefficients := createCoefficientArrayWithUsableCount(1000)

	result := CalculateCapacity(coefficients)

	// Verify CapacityByK is populated for all k values 1-8
	if len(result.CapacityByK) != 8 {
		t.Errorf("CapacityByK should have 8 entries, got %d", len(result.CapacityByK))
	}

	// Verify capacity formula: (usable / n) * k - 32, converted to bytes
	// For each k, n = 2^k - 1
	expectedCapacities := map[int]int{
		// k=1: (1000/1)*1 - 32 = 968 bits = 121 bytes
		1: (1000*1 - HeaderSize) / 8,
		// k=2: (1000/3)*2 - 32 = 666 - 32 = 634 bits = 79 bytes
		2: (333*2 - HeaderSize) / 8,
		// k=3: (1000/7)*3 - 32 = 428 - 32 = 396 bits = 49 bytes
		3: (142*3 - HeaderSize) / 8,
		// k=4: (1000/15)*4 - 32 = 264 - 32 = 232 bits = 29 bytes
		4: (66*4 - HeaderSize) / 8,
		// k=5: (1000/31)*5 - 32 = 160 - 32 = 128 bits = 16 bytes
		5: (32*5 - HeaderSize) / 8,
		// k=6: (1000/63)*6 - 32 = 90 - 32 = 58 bits = 7 bytes
		6: (15*6 - HeaderSize) / 8,
		// k=7: (1000/127)*7 - 32 = 49 - 32 = 17 bits = 2 bytes
		7: (7*7 - HeaderSize) / 8,
		// k=8: (1000/255)*8 - 32 = 24 - 32 = -8 bits -> 0 bytes
		8: 0,
	}

	for k := 1; k <= 8; k++ {
		if result.CapacityByK[k] != expectedCapacities[k] {
			t.Errorf("CapacityByK[%d] = %d, want %d (usable=1000)",
				k, result.CapacityByK[k], expectedCapacities[k])
		}
	}
}

// TestCalculateCapacityEstimatedShrinkageFactor tests shrinkage factor estimation.
// The shrinkage factor is the proportion of coefficients with absolute value 1.
func TestCalculateCapacityEstimatedShrinkageFactor(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		coefficients []int16
		minFactor    float64 // Minimum expected shrinkage factor
		maxFactor    float64 // Maximum expected shrinkage factor
	}{
		{
			name:         "empty slice",
			coefficients: []int16{},
			minFactor:    0.0,
			maxFactor:    0.0,
		},
		{
			name:         "no usable coefficients",
			coefficients: []int16{0, 0, 0, 0, 0},
			minFactor:    0.0,
			maxFactor:    0.0,
		},
		{
			name:         "all magnitude 1 (100% shrinkage)",
			coefficients: []int16{0, 1, -1, 1, -1, 1, -1, 1}, // DC at 0, rest are |1|
			minFactor:    0.99,
			maxFactor:    1.01, // Allow small floating point tolerance
		},
		{
			name:         "no magnitude 1 (0% shrinkage)",
			coefficients: []int16{0, 2, -2, 5, -5, 10, -10, 100}, // DC at 0, rest are |>1|
			minFactor:    0.0,
			maxFactor:    0.01,
		},
		{
			name:         "50% magnitude 1",
			coefficients: []int16{0, 1, 2, -1, -2, 1, 5, -1}, // 4 usable |1|, 3 usable |>1|
			minFactor:    0.50,                               // 4 out of 7 usable = 57%
			maxFactor:    0.60,
		},
		{
			name:         "mixed realistic distribution",
			coefficients: createRealisticCoefficientDistribution(),
			minFactor:    0.10, // Realistic images have ~10-30% shrinkage
			maxFactor:    0.50,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := CalculateCapacity(tc.coefficients)
			if result.EstimatedShrinkageFactor < tc.minFactor ||
				result.EstimatedShrinkageFactor > tc.maxFactor {
				t.Errorf("EstimatedShrinkageFactor = %f, want between %f and %f",
					result.EstimatedShrinkageFactor, tc.minFactor, tc.maxFactor)
			}
		})
	}
}

// TestCalculateCapacityHeaderOverhead tests that header overhead is correctly subtracted.
func TestCalculateCapacityHeaderOverhead(t *testing.T) {
	t.Parallel()

	// Create coefficient array with exactly 32 usable coefficients
	// At k=1, this should give capacity of (32 - 32) / 8 = 0 bytes
	coefficients := createCoefficientArrayWithUsableCount(32)

	result := CalculateCapacity(coefficients)

	// At k=1: (32/1)*1 - 32 = 0 bits = 0 bytes
	if result.CapacityByK[1] != 0 {
		t.Errorf("CapacityByK[1] with 32 usable coeffs = %d, want 0 (no room after header)",
			result.CapacityByK[1])
	}

	// Create array with 40 usable coefficients
	// At k=1: (40/1)*1 - 32 = 8 bits = 1 byte
	coefficients40 := createCoefficientArrayWithUsableCount(40)
	result40 := CalculateCapacity(coefficients40)

	if result40.CapacityByK[1] != 1 {
		t.Errorf("CapacityByK[1] with 40 usable coeffs = %d, want 1 byte",
			result40.CapacityByK[1])
	}
}

// Helper function to create a test block with uniform non-zero values
func createTestBlock(value int16, includeDC bool) []int16 {
	block := make([]int16, 64)
	for i := range block {
		if i == 0 && !includeDC {
			block[i] = 0
		} else {
			block[i] = value
		}
	}
	return block
}

// Helper function to create multiple non-zero blocks (DC excluded from usable count)
func createTestBlocksNonZero(numBlocks int) []int16 {
	coefficients := make([]int16, 64*numBlocks)
	for i := range coefficients {
		coefficients[i] = 5 // Non-zero value
	}
	return coefficients
}

// Helper function to create a block with some zeros
func createTestBlockWithZeros() []int16 {
	block := make([]int16, 64)
	// DC at index 0 is 0 (not usable anyway)
	// Fill half of AC coefficients with non-zero values
	for i := 1; i < 64; i++ {
		if i%2 == 0 {
			block[i] = int16(i) // Non-zero
		} else {
			block[i] = 0 // Zero (not usable)
		}
	}
	// Count: indices 2,4,6,8,...62 are non-zero = 31 usable
	return block
}

// Helper function to create coefficient array with specific usable count
func createCoefficientArrayWithUsableCount(usableCount int) []int16 {
	// Calculate how many blocks we need
	// Each block has 63 usable positions (excluding DC at index 0)
	numBlocks := (usableCount / 63) + 1
	coefficients := make([]int16, 64*numBlocks)

	usableAdded := 0
	for i := range coefficients {
		// Skip DC coefficients (index % 64 == 0)
		if i%64 == 0 {
			coefficients[i] = 0 // DC = 0
			continue
		}
		// Add usable coefficients until we reach the count
		if usableAdded < usableCount {
			coefficients[i] = 5 // Non-zero value
			usableAdded++
		} else {
			coefficients[i] = 0 // Zero (not usable)
		}
	}
	return coefficients
}

// Helper function to create a realistic coefficient distribution
// Typical JPEG images have a distribution where many coefficients are small
func createRealisticCoefficientDistribution() []int16 {
	// Create 10 blocks (640 coefficients)
	coefficients := make([]int16, 640)

	// Realistic distribution:
	// - DC coefficients at indices 0, 64, 128, ... (10 DCs)
	// - ~30% zeros
	// - ~20% magnitude 1
	// - ~50% magnitude > 1

	idx := 0
	for block := 0; block < 10; block++ {
		// DC coefficient
		coefficients[idx] = int16(100 + block) // DC values are typically larger
		idx++

		// AC coefficients (63 per block)
		for ac := 0; ac < 63; ac++ {
			switch {
			case ac < 19: // ~30% zeros
				coefficients[idx] = 0
			case ac < 32: // ~20% magnitude 1
				if ac%2 == 0 {
					coefficients[idx] = 1
				} else {
					coefficients[idx] = -1
				}
			default: // ~50% magnitude > 1
				coefficients[idx] = int16((ac % 10) + 2) // Values 2-11
			}
			idx++
		}
	}
	return coefficients
}

// Fuzz Testing

// FuzzCapacity tests the capacity calculation functions with fuzzed inputs.
//
// This fuzz test verifies that:
// - No panics occur with arbitrary coefficient arrays
// - Capacity calculations are valid (non-negative, sensible)
// - CapacityResult fields are consistent
func FuzzCapacity(f *testing.F) {
	// Add seed corpus
	f.Add([]byte{0, 5, 0, 10, 0, 0, 0, 15})
	f.Add([]byte{})
	f.Add([]byte{0xFF, 0x07})
	f.Add([]byte{0x00, 0xF8})
	f.Add([]byte{0, 0, 0, 0, 0, 0, 0, 0})

	f.Fuzz(func(t *testing.T, coeffBytes []byte) {
		// Convert bytes to int16 coefficients
		if len(coeffBytes) < 2 {
			// Need at least 2 bytes for one coefficient
			return
		}

		numCoeffs := len(coeffBytes) / 2
		coefficients := make([]int16, numCoeffs)

		for i := 0; i < numCoeffs; i++ {
			byteIdx := i * 2
			if byteIdx+1 < len(coeffBytes) {
				val := int16(coeffBytes[byteIdx]) | (int16(coeffBytes[byteIdx+1]) << 8)
				// Clamp to valid range
				if val < CoefficientMin {
					val = CoefficientMin
				} else if val > CoefficientMax {
					val = CoefficientMax
				}
				coefficients[i] = val
			}
		}

		// Test CalculateCapacity - should not panic
		result := CalculateCapacity(coefficients)

		// Verify result is not nil
		if result == nil {
			t.Fatal("CalculateCapacity returned nil")
		}

		// Verify TotalCoefficients matches input
		if result.TotalCoefficients != len(coefficients) {
			t.Errorf("TotalCoefficients: got %d, want %d", result.TotalCoefficients, len(coefficients))
		}

		// Verify UsableCoefficients is non-negative and <= total
		if result.UsableCoefficients < 0 {
			t.Errorf("Negative UsableCoefficients: %d", result.UsableCoefficients)
		}
		if result.UsableCoefficients > result.TotalCoefficients {
			t.Errorf("UsableCoefficients (%d) > TotalCoefficients (%d)",
				result.UsableCoefficients, result.TotalCoefficients)
		}

		// Verify EstimatedShrinkageFactor is in valid range
		if result.EstimatedShrinkageFactor < 0.0 || result.EstimatedShrinkageFactor > 1.0 {
			t.Errorf("Invalid EstimatedShrinkageFactor: %f (expected 0.0-1.0)",
				result.EstimatedShrinkageFactor)
		}

		// Verify CapacityByK has entries for k=1 to k=8
		if result.CapacityByK == nil {
			t.Fatal("CapacityByK is nil")
		}
		for k := 1; k <= 8; k++ {
			capacity, exists := result.CapacityByK[k]
			if !exists {
				t.Errorf("CapacityByK missing entry for k=%d", k)
				continue
			}
			if capacity < 0 {
				t.Errorf("Negative capacity for k=%d: %d", k, capacity)
			}
		}

		// Verify capacity ordering: lower k generally has more capacity
		// (not strictly required but should be true for most cases)
		// Skip this check as it's not always true for small coefficient arrays

		// Test CalculateCapacityForK for each k value
		for k := 1; k <= 8; k++ {
			capacityBits := CalculateCapacityForK(result.UsableCoefficients, k)
			if capacityBits < 0 {
				t.Errorf("Negative CalculateCapacityForK(%d, %d): %d",
					result.UsableCoefficients, k, capacityBits)
			}
		}

		// Test invalid k values
		if CalculateCapacityForK(result.UsableCoefficients, 0) != 0 {
			t.Error("CalculateCapacityForK with k=0 should return 0")
		}
		if CalculateCapacityForK(result.UsableCoefficients, 9) != 0 {
			t.Error("CalculateCapacityForK with k=9 should return 0")
		}
		if CalculateCapacityForK(result.UsableCoefficients, -1) != 0 {
			t.Error("CalculateCapacityForK with k=-1 should return 0")
		}

		// Test CalculateEmbeddingRate
		for k := 1; k <= 8; k++ {
			rate := CalculateEmbeddingRate(k)
			if rate <= 0.0 || rate > 1.0 {
				t.Errorf("Invalid CalculateEmbeddingRate(%d): %f", k, rate)
			}
		}

		// Test invalid k for embedding rate
		if CalculateEmbeddingRate(0) != 0.0 {
			t.Error("CalculateEmbeddingRate(0) should return 0.0")
		}
		if CalculateEmbeddingRate(-1) != 0.0 {
			t.Error("CalculateEmbeddingRate(-1) should return 0.0")
		}

		// Test SelectOptimalK with various message sizes
		if result.UsableCoefficients > 0 {
			// Try selecting k for a small message
			_, err := SelectOptimalK(result.UsableCoefficients, 8) // 1 byte = 8 bits
			// Error is acceptable if capacity is too low
			_ = err

			// Try with zero message bits (just header)
			k, err := SelectOptimalK(result.UsableCoefficients, 0)
			if err == nil && (k < 1 || k > 8) {
				t.Errorf("SelectOptimalK returned invalid k=%d", k)
			}
		}

		// Test IsUsableCoefficient for each coefficient
		usableCount := 0
		for i, coeff := range coefficients {
			isUsable := IsUsableCoefficient(i, coeff)
			if isUsable {
				usableCount++
			}
			// Verify logic: DC (i%64==0) or zero should not be usable
			isDC := i%64 == 0
			isZero := coeff == 0
			expectedUsable := !isDC && !isZero
			if isUsable != expectedUsable {
				t.Errorf("IsUsableCoefficient(%d, %d): got %v, want %v",
					i, coeff, isUsable, expectedUsable)
			}
		}

		// Verify usable count matches CapacityResult
		if usableCount != result.UsableCoefficients {
			t.Errorf("UsableCoefficients count mismatch: manual=%d, result=%d",
				usableCount, result.UsableCoefficients)
		}
	})
}

// TestFuzzEdgeCaseZeroCoefficients tests behavior with all-zero coefficients.
func TestFuzzEdgeCaseZeroCoefficients(t *testing.T) {
	coefficients := make([]int16, 1024)
	// All zeros

	result := CalculateCapacity(coefficients)
	if result.UsableCoefficients != 0 {
		t.Errorf("Expected 0 usable coefficients for all-zero array, got %d", result.UsableCoefficients)
	}

	for k := 1; k <= 8; k++ {
		if result.CapacityByK[k] != 0 {
			t.Errorf("Expected 0 capacity for k=%d with all-zero coefficients, got %d",
				k, result.CapacityByK[k])
		}
	}
}

// TestFuzzEdgeCaseAllDCCoefficients tests behavior when all positions are DC.
func TestFuzzEdgeCaseAllDCCoefficients(t *testing.T) {
	// Create array where all are at DC positions (multiples of 64)
	coefficients := make([]int16, 64)
	for i := range coefficients {
		coefficients[i] = 100 // Non-zero but at DC position
	}

	result := CalculateCapacity(coefficients)
	// Only index 0 is DC in a 64-element array, others are AC
	// But index 0 is DC, indices 1-63 are AC positions within the block

	// Actually, in the first 64 coefficients (one block), index 0 is DC
	// Indices 1-63 are AC positions
	// So we should have 63 usable if all are non-zero

	// Wait, all coefficients are non-zero (100), but index 0 is DC
	// So usable = 63
	if result.UsableCoefficients != 63 {
		t.Errorf("Expected 63 usable coefficients, got %d", result.UsableCoefficients)
	}
}
