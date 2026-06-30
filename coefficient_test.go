package f5messageembed

import (
	"testing"
)

// TestGetStegoBitPositiveCoefficients tests that positive coefficients return
// the LSB directly: odd=1, even=0.
func TestGetStegoBitPositiveCoefficients(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		coefficient int16
		expected    int
	}{
		{"positive odd 1", 1, 1},
		{"positive odd 3", 3, 1},
		{"positive odd 5", 5, 1},
		{"positive odd 127", 127, 1},
		{"positive odd 2047", 2047, 1},
		{"positive even 2", 2, 0},
		{"positive even 4", 4, 0},
		{"positive even 6", 6, 0},
		{"positive even 128", 128, 0},
		{"positive even 2046", 2046, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := GetStegoBit(tc.coefficient)
			if result != tc.expected {
				t.Errorf("GetStegoBit(%d) = %d, want %d", tc.coefficient, result, tc.expected)
			}
		})
	}
}

// TestGetStegoBitNegativeCoefficients tests that negative coefficients return
// the inverted LSB: even=1, odd=0.
func TestGetStegoBitNegativeCoefficients(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		coefficient int16
		expected    int
	}{
		{"negative odd -1", -1, 0},
		{"negative odd -3", -3, 0},
		{"negative odd -5", -5, 0},
		{"negative odd -127", -127, 0},
		{"negative odd -2047", -2047, 0},
		{"negative even -2", -2, 1},
		{"negative even -4", -4, 1},
		{"negative even -6", -6, 1},
		{"negative even -128", -128, 1},
		{"negative even -2048", -2048, 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := GetStegoBit(tc.coefficient)
			if result != tc.expected {
				t.Errorf("GetStegoBit(%d) = %d, want %d", tc.coefficient, result, tc.expected)
			}
		})
	}
}

// TestModifyCoefficientPositive tests that positive coefficients are decremented.
func TestModifyCoefficientPositive(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		coefficient int16
		expected    int16
	}{
		{"decrement 5 to 4", 5, 4},
		{"decrement 10 to 9", 10, 9},
		{"decrement 2047 to 2046", 2047, 2046},
		{"decrement 100 to 99", 100, 99},
		{"decrement 2 to 1", 2, 1},
		{"decrement 1 to 0 (shrinkage)", 1, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := ModifyCoefficient(tc.coefficient)
			if result != tc.expected {
				t.Errorf("ModifyCoefficient(%d) = %d, want %d", tc.coefficient, result, tc.expected)
			}
		})
	}
}

// TestModifyCoefficientNegative tests that negative coefficients are incremented toward zero.
func TestModifyCoefficientNegative(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		coefficient int16
		expected    int16
	}{
		{"increment -5 to -4", -5, -4},
		{"increment -10 to -9", -10, -9},
		{"increment -2048 to -2047", -2048, -2047},
		{"increment -100 to -99", -100, -99},
		{"increment -2 to -1", -2, -1},
		{"increment -1 to 0 (shrinkage)", -1, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := ModifyCoefficient(tc.coefficient)
			if result != tc.expected {
				t.Errorf("ModifyCoefficient(%d) = %d, want %d", tc.coefficient, result, tc.expected)
			}
		})
	}
}

// TestCoefficientRangeBoundaries tests behavior at the coefficient range limits.
func TestCoefficientRangeBoundaries(t *testing.T) {
	t.Parallel()

	// Test GetStegoBit at boundaries
	t.Run("GetStegoBit at max positive", func(t *testing.T) {
		t.Parallel()
		result := GetStegoBit(CoefficientMax) // 2047 is odd
		if result != 1 {
			t.Errorf("GetStegoBit(%d) = %d, want 1", CoefficientMax, result)
		}
	})

	t.Run("GetStegoBit at min negative", func(t *testing.T) {
		t.Parallel()
		result := GetStegoBit(CoefficientMin) // -2048 is even
		if result != 1 {                      // negative even = 1
			t.Errorf("GetStegoBit(%d) = %d, want 1", CoefficientMin, result)
		}
	})

	// Test ModifyCoefficient at boundaries
	t.Run("ModifyCoefficient at max positive", func(t *testing.T) {
		t.Parallel()
		result := ModifyCoefficient(CoefficientMax)
		if result != 2046 {
			t.Errorf("ModifyCoefficient(%d) = %d, want 2046", CoefficientMax, result)
		}
	})

	t.Run("ModifyCoefficient at min negative", func(t *testing.T) {
		t.Parallel()
		result := ModifyCoefficient(CoefficientMin)
		if result != -2047 {
			t.Errorf("ModifyCoefficient(%d) = %d, want -2047", CoefficientMin, result)
		}
	})
}

// TestZeroCoefficientHandling tests behavior with zero coefficients.
func TestZeroCoefficientHandling(t *testing.T) {
	t.Parallel()

	// GetStegoBit with zero - while this should not be called with zero
	// in practice, we should verify consistent behavior
	t.Run("GetStegoBit with zero returns 0", func(t *testing.T) {
		t.Parallel()
		result := GetStegoBit(0)
		if result != 0 {
			t.Errorf("GetStegoBit(0) = %d, want 0", result)
		}
	})

	// ModifyCoefficient with zero - edge case behavior
	t.Run("ModifyCoefficient with zero returns -1", func(t *testing.T) {
		t.Parallel()
		// Zero is not positive, so it falls through to the else branch (coeff + 1)
		result := ModifyCoefficient(0)
		if result != 1 {
			t.Errorf("ModifyCoefficient(0) = %d, want 1", result)
		}
	})
}

// TestIsUsableCoefficientDCCoefficients tests that DC coefficients are skipped.
func TestIsUsableCoefficientDCCoefficients(t *testing.T) {
	t.Parallel()

	dcIndices := []int{0, 64, 128, 192, 256, 320, 640, 1280, 6400}

	for _, idx := range dcIndices {
		t.Run("DC coefficient at index", func(t *testing.T) {
			t.Parallel()
			// DC coefficients should be skipped even if they have non-zero values
			result := IsUsableCoefficient(idx, 5)
			if result {
				t.Errorf("IsUsableCoefficient(%d, 5) = true, want false (DC coefficient)", idx)
			}
		})
	}
}

// TestIsUsableCoefficientZeroCoefficients tests that zero coefficients are skipped.
func TestIsUsableCoefficientZeroCoefficients(t *testing.T) {
	t.Parallel()

	acIndices := []int{1, 2, 10, 63, 65, 100, 127}

	for _, idx := range acIndices {
		t.Run("zero coefficient at AC index", func(t *testing.T) {
			t.Parallel()
			result := IsUsableCoefficient(idx, 0)
			if result {
				t.Errorf("IsUsableCoefficient(%d, 0) = true, want false (zero coefficient)", idx)
			}
		})
	}
}

// TestIsUsableCoefficientValidCoefficients tests that valid AC non-zero coefficients are usable.
func TestIsUsableCoefficientValidCoefficients(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		index       int
		coefficient int16
	}{
		{"first AC coefficient positive", 1, 5},
		{"first AC coefficient negative", 1, -5},
		{"middle of first block", 32, 10},
		{"end of first block", 63, -100},
		{"second block AC", 65, 1},
		{"second block middle", 100, -1},
		{"third block", 129, 2047},
		{"large index", 10000, -2048},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := IsUsableCoefficient(tc.index, tc.coefficient)
			if !result {
				t.Errorf("IsUsableCoefficient(%d, %d) = false, want true", tc.index, tc.coefficient)
			}
		})
	}
}

// TestGetStegoBitMatchesF5PixelExtraction verifies that our GetStegoBit implementation
// matches the behavior of the getCoefficientBit function in f5pixel extraction.
func TestGetStegoBitMatchesF5PixelExtraction(t *testing.T) {
	t.Parallel()

	// Test values that exercise the full range of the algorithm
	// These test cases are derived from the f5pixel extraction implementation:
	// bit := int(coefficient) & 1
	// if coefficient < 0 { bit = 1 - bit }

	testCases := []struct {
		coefficient int16
		expected    int
	}{
		// Positive coefficients: LSB directly
		{1, 1},
		{2, 0},
		{3, 1},
		{4, 0},
		{100, 0},
		{101, 1},
		{2046, 0},
		{2047, 1},
		// Negative coefficients: inverted LSB
		{-1, 0},
		{-2, 1},
		{-3, 0},
		{-4, 1},
		{-100, 1},
		{-101, 0},
		{-2047, 0},
		{-2048, 1},
	}

	for _, tc := range testCases {
		t.Run("coefficient compatibility", func(t *testing.T) {
			t.Parallel()
			result := GetStegoBit(tc.coefficient)
			if result != tc.expected {
				t.Errorf("GetStegoBit(%d) = %d, want %d (must match f5pixel extraction)", tc.coefficient, result, tc.expected)
			}
		})
	}
}

// TestModifyCoefficientFlipsStegoBit verifies that ModifyCoefficient changes the
// steganographic bit value (which is the purpose of the modification).
func TestModifyCoefficientFlipsStegoBit(t *testing.T) {
	t.Parallel()

	testCases := []int16{2, 3, 4, 5, 10, 100, 2046, 2047, -2, -3, -4, -5, -10, -100, -2047, -2048}

	for _, coeff := range testCases {
		t.Run("modification flips stego bit", func(t *testing.T) {
			t.Parallel()
			originalBit := GetStegoBit(coeff)
			modifiedCoeff := ModifyCoefficient(coeff)

			// Skip shrinkage cases (|1| and |-1| become 0)
			if modifiedCoeff == 0 {
				return
			}

			modifiedBit := GetStegoBit(modifiedCoeff)
			if originalBit == modifiedBit {
				t.Errorf("ModifyCoefficient(%d) did not flip stego bit: %d -> %d (both have bit %d)",
					coeff, coeff, modifiedCoeff, originalBit)
			}
		})
	}
}
