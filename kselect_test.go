package f5messageembed

import (
	"errors"
	"testing"
)

// TestCalculateEmbeddingRateFormula tests that R(k) = k/(2^k - 1) is calculated correctly.
func TestCalculateEmbeddingRateFormula(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		k        int
		expected float64
	}{
		// k=1: R(1) = 1/(2^1 - 1) = 1/1 = 1.0
		{1, 1.0},
		// k=2: R(2) = 2/(2^2 - 1) = 2/3 = 0.666...
		{2, 2.0 / 3.0},
		// k=3: R(3) = 3/(2^3 - 1) = 3/7 = 0.4285...
		{3, 3.0 / 7.0},
		// k=4: R(4) = 4/(2^4 - 1) = 4/15 = 0.2666...
		{4, 4.0 / 15.0},
		// k=5: R(5) = 5/(2^5 - 1) = 5/31 = 0.1612...
		{5, 5.0 / 31.0},
		// k=6: R(6) = 6/(2^6 - 1) = 6/63 = 0.0952...
		{6, 6.0 / 63.0},
		// k=7: R(7) = 7/(2^7 - 1) = 7/127 = 0.0551...
		{7, 7.0 / 127.0},
		// k=8: R(8) = 8/(2^8 - 1) = 8/255 = 0.0313...
		{8, 8.0 / 255.0},
	}

	for _, tc := range testCases {
		result := CalculateEmbeddingRate(tc.k)
		// Use a small epsilon for floating-point comparison
		epsilon := 0.0001
		diff := result - tc.expected
		if diff < 0 {
			diff = -diff
		}
		if diff > epsilon {
			t.Errorf("CalculateEmbeddingRate(%d) = %f, want %f", tc.k, result, tc.expected)
		}
	}
}

// TestSelectOptimalKSmallMessage tests that higher k is selected for small messages
// (better efficiency when capacity is abundant).
func TestSelectOptimalKSmallMessage(t *testing.T) {
	t.Parallel()

	// Large capacity with small message should select the highest embeddable k
	// for efficiency. Embedding is capped at k=7.
	// With 10000 usable coefficients and only 100 bits of message:
	// Need 100 + 32 = 132 bits total
	// For k=7 (n=127): capacity = (10000/127)*7 = 78*7 = 546 bits >= 132
	// k=7 should work since it has ample capacity
	usableCoeffs := 10000
	messageBits := 100

	k, err := SelectOptimalK(usableCoeffs, messageBits)
	if err != nil {
		t.Fatalf("SelectOptimalK returned unexpected error: %v", err)
	}

	// Should select k=7 (the highest embeddable k that fits)
	if k != 7 {
		t.Errorf("SelectOptimalK(%d, %d) = %d, want 7 (highest k for small message)",
			usableCoeffs, messageBits, k)
	}
}

// TestSelectOptimalKLargeMessage tests that lower k is selected for large messages
// (need more capacity even at cost of efficiency).
func TestSelectOptimalKLargeMessage(t *testing.T) {
	t.Parallel()

	// With limited capacity, we need lower k for higher capacity
	// For 500 usable coefficients and 400 bits (need 400 + 32 = 432 bits):
	// k=1 (n=1): capacity = (500/1)*1 = 500 bits >= 432 - OK
	// k=2 (n=3): capacity = (500/3)*2 = 166*2 = 332 bits < 432 - FAILS
	usableCoeffs := 500
	messageBits := 400

	k, err := SelectOptimalK(usableCoeffs, messageBits)
	if err != nil {
		t.Fatalf("SelectOptimalK returned unexpected error: %v", err)
	}

	// Should select k=1 (the only k that fits)
	if k != 1 {
		t.Errorf("SelectOptimalK(%d, %d) = %d, want 1 (lower k for large message)",
			usableCoeffs, messageBits, k)
	}
}

// TestSelectOptimalKBoundaryConditions tests edge cases for k selection.
func TestSelectOptimalKBoundaryConditions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		usableCoeffs int
		messageBits  int
		expectedK    int
		expectError  bool
	}{
		// Ample capacity with a small message selects the max embeddable k=7.
		// For k=7 (n=127): (1275/127)*7 = 10*7 = 70 >= 8 + 32 = 40 bits.
		{
			name:         "small message selects max k=7",
			usableCoeffs: 1275,
			messageBits:  8,
			expectedK:    7,
			expectError:  false,
		},
		// Zero message bits with ample capacity also selects max k=7.
		// For k=7 (n=127): (1020/127)*7 = 8*7 = 56 >= 32 bits.
		{
			name:         "zero message bits selects max k=7",
			usableCoeffs: 1020,
			messageBits:  0,
			expectedK:    7,
			expectError:  false,
		},
		// k=7 is the highest embeddable k and still fits here.
		// k=7 (n=127): (635/127)*7 = 5*7 = 35 >= 32 - OK
		{
			name:         "k=7 at the embeddable ceiling",
			usableCoeffs: 635, // (635/127)*7 = 35 >= 32
			messageBits:  0,
			expectedK:    7,
			expectError:  false,
		},
		// Minimum viable capacity (just enough for header with k=1)
		{
			name:         "minimum for header only with k=1",
			usableCoeffs: 32, // k=1: (32/1)*1 = 32 >= 32 (exact fit)
			messageBits:  0,
			expectedK:    1,
			expectError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			k, err := SelectOptimalK(tc.usableCoeffs, tc.messageBits)
			if tc.expectError {
				if err == nil {
					t.Errorf("SelectOptimalK(%d, %d) should return error",
						tc.usableCoeffs, tc.messageBits)
				}
			} else {
				if err != nil {
					t.Fatalf("SelectOptimalK returned unexpected error: %v", err)
				}
				if k != tc.expectedK {
					t.Errorf("SelectOptimalK(%d, %d) = %d, want %d",
						tc.usableCoeffs, tc.messageBits, k, tc.expectedK)
				}
			}
		})
	}
}

// TestSelectOptimalKMessageTooLarge tests error when message cannot fit at any k.
func TestSelectOptimalKMessageTooLarge(t *testing.T) {
	t.Parallel()

	// Very small capacity that cannot accommodate even the smallest message
	// For k=1 (n=1): capacity = (coeffs/1)*1
	// If we have only 10 coefficients: capacity = 10 bits
	// Need 100 + 32 = 132 bits, which exceeds 10
	usableCoeffs := 10
	messageBits := 100 // Way too large for 10 coefficients

	_, err := SelectOptimalK(usableCoeffs, messageBits)
	if err == nil {
		t.Error("SelectOptimalK should return error when message too large for capacity")
	}

	// Verify it returns the expected error type (ValidationError with capacity details key)
	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Errorf("SelectOptimalK should return *ValidationError, got: %T", err)
	} else if valErr.Key != ErrKeyCapacityDetails {
		t.Errorf("SelectOptimalK should return error with key %q, got: %q", ErrKeyCapacityDetails, valErr.Key)
	}
}

// TestSelectOptimalKInsufficientForHeader tests error when capacity cannot fit header.
func TestSelectOptimalKInsufficientForHeader(t *testing.T) {
	t.Parallel()

	// Not even enough for the 32-bit header at k=1
	// k=1: capacity = (coeffs/1)*1 >= 32
	// Need at least 32 coefficients for just the header
	usableCoeffs := 20 // Less than 32
	messageBits := 0

	_, err := SelectOptimalK(usableCoeffs, messageBits)
	if err == nil {
		t.Error("SelectOptimalK should return error when capacity insufficient for header")
	}

	// Verify it returns the expected error type (ValidationError with capacity details key)
	var valErr2 *ValidationError
	if !errors.As(err, &valErr2) {
		t.Errorf("SelectOptimalK should return *ValidationError, got: %T", err)
	} else if valErr2.Key != ErrKeyCapacityDetails {
		t.Errorf("SelectOptimalK should return error with key %q, got: %q", ErrKeyCapacityDetails, valErr2.Key)
	}
}

// TestSelectOptimalKDecrementalSearch verifies that k is tried from 7 down to 1,
// selecting the largest k that provides sufficient capacity.
func TestSelectOptimalKDecrementalSearch(t *testing.T) {
	t.Parallel()

	// Design a case where k=5 is the largest k that fits
	// Need 200 + 32 = 232 bits
	// k=5 (n=31): capacity = (coeffs/31)*5
	// k=6 (n=63): capacity = (coeffs/63)*6
	//
	// For k=5 to be exactly sufficient: (coeffs/31)*5 >= 232
	// coeffs >= 232*31/5 = 1438.4, so need 1439 or more
	// BUT with integer division: (1439/31)*5 = 46*5 = 230 < 232 - FAILS
	// Need: (coeffs/31) >= 47, so coeffs >= 47*31 = 1457
	// Check: (1457/31)*5 = 47*5 = 235 >= 232 - OK
	//
	// Verify k=6 doesn't work: (1457/63)*6 = 23*6 = 138 < 232 - FAILS

	usableCoeffs := 1457
	messageBits := 200

	k, err := SelectOptimalK(usableCoeffs, messageBits)
	if err != nil {
		t.Fatalf("SelectOptimalK returned unexpected error: %v", err)
	}

	if k != 5 {
		t.Errorf("SelectOptimalK(%d, %d) = %d, want 5", usableCoeffs, messageBits, k)
	}

	// Also verify that k=4 would have worked (higher capacity)
	// k=4 (n=15): (1457/15)*4 = 97*4 = 388 >= 232 - OK
	// Since we try from k=7 down, k=5 should be selected (not k=4)
}

// TestCalculateEmbeddingRateInvalidK tests embedding rate for edge case k values.
func TestCalculateEmbeddingRateInvalidK(t *testing.T) {
	t.Parallel()

	// k=0 would cause division by zero (2^0 - 1 = 0)
	// The function should handle this gracefully
	result := CalculateEmbeddingRate(0)
	if result != 0 {
		t.Errorf("CalculateEmbeddingRate(0) = %f, want 0 (or safe default)", result)
	}

	// Negative k should also be handled
	result = CalculateEmbeddingRate(-1)
	if result != 0 {
		t.Errorf("CalculateEmbeddingRate(-1) = %f, want 0 (or safe default)", result)
	}
}

// TestCalculateCapacityForK tests the helper function that computes capacity for a given k.
func TestCalculateCapacityForK(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		usableCoeffs int
		k            int
		expected     int
	}{
		// k=1 (n=1): capacity = (coeffs/1)*1 - 32
		{"k=1 with 100 coeffs", 100, 1, 68}, // 100 - 32 = 68
		// k=2 (n=3): capacity = (coeffs/3)*2 - 32
		{"k=2 with 100 coeffs", 100, 2, 34}, // (33)*2 - 32 = 66 - 32 = 34
		// k=4 (n=15): capacity = (coeffs/15)*4 - 32
		{"k=4 with 1000 coeffs", 1000, 4, 232}, // (66)*4 - 32 = 264 - 32 = 232
		// k=8 (n=255): capacity = (coeffs/255)*8 - 32
		{"k=8 with 10000 coeffs", 10000, 8, 280}, // (39)*8 - 32 = 312 - 32 = 280
		// Not enough for header
		{"k=1 insufficient for header", 20, 1, 0}, // 20 - 32 < 0 -> 0
		// Invalid k
		{"k=0 invalid", 1000, 0, 0},
		{"k=9 invalid", 1000, 9, 0},
		{"k=-1 invalid", 1000, -1, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := CalculateCapacityForK(tc.usableCoeffs, tc.k)
			if result != tc.expected {
				t.Errorf("CalculateCapacityForK(%d, %d) = %d, want %d",
					tc.usableCoeffs, tc.k, result, tc.expected)
			}
		})
	}
}

// Coverage Tests

// TestCalculateCapacityBytes_InvalidK tests calculateCapacityBytes with k > 8.
func TestCalculateCapacityBytes_InvalidK(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		k            int
		usableCoeffs int
		expected     int
	}{
		{"k=9 invalid", 9, 1000, 0},
		{"k=10 invalid", 10, 1000, 0},
		{"k=100 invalid", 100, 1000, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := calculateCapacityBytes(tc.usableCoeffs, tc.k)
			if result != tc.expected {
				t.Errorf("calculateCapacityBytes(%d, %d) = %d, want %d",
					tc.usableCoeffs, tc.k, result, tc.expected)
			}
		})
	}
}

// TestItoa_NegativeNumbers tests the itoa helper with negative numbers.
func TestItoa_NegativeNumbers(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		expected string
		input    int
	}{
		{expected: "0", input: 0},
		{expected: "1", input: 1},
		{expected: "-1", input: -1},
		{expected: "-123", input: -123},
		{expected: "-2048", input: -2048},
		{expected: "123456", input: 123456},
		{expected: "-987654", input: -987654},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			t.Parallel()
			result := itoa(tc.input)
			if result != tc.expected {
				t.Errorf("itoa(%d) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}
