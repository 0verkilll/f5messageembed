package f5messageembed

import (
	"testing"
)

// TestDetectShrinkagePositiveOne tests that coefficient 1 is detected as causing shrinkage.
func TestDetectShrinkagePositiveOne(t *testing.T) {
	t.Parallel()

	result := DetectShrinkage(1)
	if !result {
		t.Errorf("DetectShrinkage(1) = false, want true (|1| becomes 0)")
	}
}

// TestDetectShrinkageNegativeOne tests that coefficient -1 is detected as causing shrinkage.
func TestDetectShrinkageNegativeOne(t *testing.T) {
	t.Parallel()

	result := DetectShrinkage(-1)
	if !result {
		t.Errorf("DetectShrinkage(-1) = false, want true (|-1| becomes 0)")
	}
}

// TestDetectShrinkageNoShrinkageLargerCoefficients tests that coefficients with
// absolute value >= 2 do not cause shrinkage.
func TestDetectShrinkageNoShrinkageLargerCoefficients(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		coefficient int16
	}{
		{"positive 2", 2},
		{"positive 3", 3},
		{"positive 10", 10},
		{"positive 100", 100},
		{"positive max", CoefficientMax},
		{"negative -2", -2},
		{"negative -3", -3},
		{"negative -10", -10},
		{"negative -100", -100},
		{"negative min", CoefficientMin},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := DetectShrinkage(tc.coefficient)
			if result {
				t.Errorf("DetectShrinkage(%d) = true, want false (|coeff| > 1 should not shrink)", tc.coefficient)
			}
		})
	}
}

// TestDetectShrinkageZeroCoefficient tests that zero does not cause shrinkage.
// Zero coefficients should not be passed to this function in practice,
// but we verify consistent behavior.
func TestDetectShrinkageZeroCoefficient(t *testing.T) {
	t.Parallel()

	result := DetectShrinkage(0)
	if result {
		t.Errorf("DetectShrinkage(0) = true, want false (zero is not 1 or -1)")
	}
}

// TestHandleShrinkageReturnsCorrectNewCoefficient tests that HandleShrinkage
// correctly applies the modification and returns the new coefficient value.
func TestHandleShrinkageReturnsCorrectNewCoefficient(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		coefficient   int16
		expectedCoeff int16
		expectShrunk  bool
	}{
		// Shrinkage cases
		{"positive 1 shrinks to 0", 1, 0, true},
		{"negative -1 shrinks to 0", -1, 0, true},
		// Non-shrinkage cases
		{"positive 2 becomes 1", 2, 1, false},
		{"positive 5 becomes 4", 5, 4, false},
		{"positive max becomes max-1", CoefficientMax, CoefficientMax - 1, false},
		{"negative -2 becomes -1", -2, -1, false},
		{"negative -5 becomes -4", -5, -4, false},
		{"negative min becomes min+1", CoefficientMin, CoefficientMin + 1, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// We're testing the coefficient modification and shrinkage detection
			// The codeWord and position parameters are used for context in the full
			// embedding loop but HandleShrinkage focuses on the coefficient itself
			codeWord := []int{0, 1, 0} // Dummy code word
			position := 1              // Dummy position

			shrunk, newCoeff := HandleShrinkage(codeWord, position, tc.coefficient)

			if shrunk != tc.expectShrunk {
				t.Errorf("HandleShrinkage(codeWord, %d, %d) shrunk = %v, want %v",
					position, tc.coefficient, shrunk, tc.expectShrunk)
			}
			if newCoeff != tc.expectedCoeff {
				t.Errorf("HandleShrinkage(codeWord, %d, %d) newCoeff = %d, want %d",
					position, tc.coefficient, newCoeff, tc.expectedCoeff)
			}
		})
	}
}

// TestHandleShrinkageIntegration tests the complete shrinkage handling workflow
// by verifying that shrinkage detection and coefficient modification work together.
func TestHandleShrinkageIntegration(t *testing.T) {
	t.Parallel()

	// Test that the shrunk flag correctly indicates when the modified coefficient becomes 0
	testCases := []struct {
		name         string
		coefficient  int16
		expectShrunk bool
	}{
		{"shrinkage from 1", 1, true},
		{"shrinkage from -1", -1, true},
		{"no shrinkage from 2", 2, false},
		{"no shrinkage from -2", -2, false},
		{"no shrinkage from 100", 100, false},
		{"no shrinkage from -100", -100, false},
	}

	codeWord := []int{1, 0, 1}
	position := 0

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			shrunk, newCoeff := HandleShrinkage(codeWord, position, tc.coefficient)

			// Verify shrinkage detection is consistent with new coefficient being zero
			if tc.expectShrunk && newCoeff != 0 {
				t.Errorf("Expected shrinkage but newCoeff = %d (not 0)", newCoeff)
			}
			if !tc.expectShrunk && newCoeff == 0 {
				t.Errorf("Did not expect shrinkage but newCoeff = 0")
			}
			if shrunk != tc.expectShrunk {
				t.Errorf("HandleShrinkage shrunk = %v, want %v", shrunk, tc.expectShrunk)
			}
		})
	}
}

// TestShrinkageCounterTracking tests that shrinkage events can be tracked
// by verifying the shrunk return value across multiple operations.
func TestShrinkageCounterTracking(t *testing.T) {
	t.Parallel()

	// Simulate a sequence of coefficient modifications and count shrinkage events
	coefficients := []int16{5, 1, -3, -1, 10, 1, -2}
	expectedShrinkageCount := 3 // coefficients 1, -1, and the last 1 should shrink

	codeWord := []int{0, 1}
	position := 0
	shrinkageCount := 0

	for _, coeff := range coefficients {
		shrunk, _ := HandleShrinkage(codeWord, position, coeff)
		if shrunk {
			shrinkageCount++
		}
	}

	if shrinkageCount != expectedShrinkageCount {
		t.Errorf("Shrinkage count = %d, want %d", shrinkageCount, expectedShrinkageCount)
	}
}

// Coverage Tests

// TestApplyChangeWithShrinkageHandling_NoChange tests when changePos is 0.
// NOTE: the standalone helper applyChangeWithShrinkageHandling (a
// shift-the-buffer + append-one strategy) was removed when the matrix loop in
// embedMessageWithMatrix was rewritten to f5.jar's faithful startOfN-restart
// shrinkage cascade (JpegEncoder.java:440-488). The no-change behaviour it used
// to test (changePos==0 ⇒ no modification, permIndex untouched) is now covered
// by the `if changePos == 0 { break }` branch in embedMessageWithMatrix and by
// TestF5JarParity's byte-identical comparison against the real f5.jar.

// Fuzz Testing

// FuzzShrinkageHandling tests shrinkage detection and handling with fuzzed inputs.
//
// This fuzz test verifies that:
// - No panics occur with arbitrary coefficient values
// - Shrinkage is correctly detected for |1| and |-1|
// - Modified coefficient values are correct
func FuzzShrinkageHandling(f *testing.F) {
	// Add seed corpus with edge cases
	f.Add(int16(1))
	f.Add(int16(-1))
	f.Add(int16(0))
	f.Add(int16(2))
	f.Add(int16(-2))
	f.Add(int16(2047))
	f.Add(int16(-2048))
	f.Add(int16(100))
	f.Add(int16(-100))

	f.Fuzz(func(t *testing.T, coefficient int16) {
		// Test DetectShrinkage - should not panic
		willShrink := DetectShrinkage(coefficient)

		// Verify shrinkage detection logic
		expectedShrink := coefficient == 1 || coefficient == -1
		if willShrink != expectedShrink {
			t.Errorf("DetectShrinkage(%d): got %v, want %v", coefficient, willShrink, expectedShrink)
		}

		// Test HandleShrinkage - should not panic
		// Use empty code word and position 0 since they're reserved for future use
		shrunk, newCoeff := HandleShrinkage([]int{}, 0, coefficient)

		// Verify shrinkage flag matches detection
		if shrunk != expectedShrink {
			t.Errorf("HandleShrinkage(%d) shrunk: got %v, want %v", coefficient, shrunk, expectedShrink)
		}

		// Verify new coefficient value based on sign
		var expectedNewCoeff int16
		switch {
		case coefficient > 0:
			expectedNewCoeff = coefficient - 1
		case coefficient < 0:
			expectedNewCoeff = coefficient + 1
		default:
			// coefficient == 0: ModifyCoefficient(0) returns 1
			// Looking at the code: if coefficient > 0: return coefficient - 1
			// else: return coefficient + 1
			// So for 0: it's not > 0, so it goes to else: return 0 + 1 = 1
			expectedNewCoeff = 1
		}

		if newCoeff != expectedNewCoeff {
			t.Errorf("HandleShrinkage(%d) newCoeff: got %d, want %d", coefficient, newCoeff, expectedNewCoeff)
		}

		// Test GetStegoBit - should not panic
		if coefficient != 0 {
			bit := GetStegoBit(coefficient)
			if bit != 0 && bit != 1 {
				t.Errorf("GetStegoBit(%d): got %d, want 0 or 1", coefficient, bit)
			}

			// Verify sign-based encoding
			lsb := int(coefficient) & 1
			if coefficient > 0 {
				if bit != lsb {
					t.Errorf("GetStegoBit(%d) positive: got %d, want %d", coefficient, bit, lsb)
				}
			} else {
				expectedBit := 1 - lsb
				if bit != expectedBit {
					t.Errorf("GetStegoBit(%d) negative: got %d, want %d", coefficient, bit, expectedBit)
				}
			}
		}

		// Test ModifyCoefficient - should not panic
		modifiedCoeff := ModifyCoefficient(coefficient)
		if coefficient > 0 {
			if modifiedCoeff != coefficient-1 {
				t.Errorf("ModifyCoefficient(%d): got %d, want %d", coefficient, modifiedCoeff, coefficient-1)
			}
		} else {
			if modifiedCoeff != coefficient+1 {
				t.Errorf("ModifyCoefficient(%d): got %d, want %d", coefficient, modifiedCoeff, coefficient+1)
			}
		}
	})
}

// TestFuzzEdgeCaseBoundaryCoefficients tests coefficients at range boundaries.
func TestFuzzEdgeCaseBoundaryCoefficients(t *testing.T) {
	testCases := []struct {
		name  string
		coeff int16
	}{
		{"max", CoefficientMax},
		{"min", CoefficientMin},
		{"max-1", CoefficientMax - 1},
		{"min+1", CoefficientMin + 1},
		{"one", 1},
		{"neg_one", -1},
		{"two", 2},
		{"neg_two", -2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// These should not panic
			_ = GetStegoBit(tc.coeff)
			_ = ModifyCoefficient(tc.coeff)
			_ = DetectShrinkage(tc.coeff)
			_, _ = HandleShrinkage([]int{0}, 1, tc.coeff)
		})
	}
}

// TestFuzzEdgeCaseOverflowProtection tests that int16 arithmetic doesn't overflow.
func TestFuzzEdgeCaseOverflowProtection(t *testing.T) {
	// Test that ModifyCoefficient handles boundary correctly
	// CoefficientMin = -2048
	// ModifyCoefficient(-2048) should return -2047 (increment toward zero)
	result := ModifyCoefficient(CoefficientMin)
	if result != CoefficientMin+1 {
		t.Errorf("ModifyCoefficient(%d): got %d, want %d", CoefficientMin, result, CoefficientMin+1)
	}

	// CoefficientMax = 2047
	// ModifyCoefficient(2047) should return 2046 (decrement)
	result = ModifyCoefficient(CoefficientMax)
	if result != CoefficientMax-1 {
		t.Errorf("ModifyCoefficient(%d): got %d, want %d", CoefficientMax, result, CoefficientMax-1)
	}
}
