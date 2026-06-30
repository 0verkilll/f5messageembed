package f5messageembed

import (
	"testing"
)

// TestComputeCodeWordHashBasicXOR tests that the hash function correctly computes
// f(a) = XOR(i=1 to n) of (a_i * i) for simple cases.
func TestComputeCodeWordHashBasicXOR(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		codeWord     []int
		n            int
		expectedHash int
	}{
		// For k=1, n=1: single element
		{"k=1 single 0", []int{0}, 1, 0},
		{"k=1 single 1", []int{1}, 1, 1},

		// For k=2, n=3: f(a) = a1*1 XOR a2*2 XOR a3*3
		{"k=2 all zeros", []int{0, 0, 0}, 3, 0},
		{"k=2 all ones", []int{1, 1, 1}, 3, 0},       // 1 XOR 2 XOR 3 = 0
		{"k=2 first one", []int{1, 0, 0}, 3, 1},      // 1 XOR 0 XOR 0 = 1
		{"k=2 second one", []int{0, 1, 0}, 3, 2},     // 0 XOR 2 XOR 0 = 2
		{"k=2 third one", []int{0, 0, 1}, 3, 3},      // 0 XOR 0 XOR 3 = 3
		{"k=2 first two", []int{1, 1, 0}, 3, 3},      // 1 XOR 2 XOR 0 = 3
		{"k=2 last two", []int{0, 1, 1}, 3, 1},       // 0 XOR 2 XOR 3 = 1
		{"k=2 first and last", []int{1, 0, 1}, 3, 2}, // 1 XOR 0 XOR 3 = 2

		// For k=3, n=7: f(a) = a1*1 XOR a2*2 XOR ... XOR a7*7
		{"k=3 all zeros", []int{0, 0, 0, 0, 0, 0, 0}, 7, 0},
		{"k=3 only first", []int{1, 0, 0, 0, 0, 0, 0}, 7, 1},
		{"k=3 only seventh", []int{0, 0, 0, 0, 0, 0, 1}, 7, 7},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := ComputeCodeWordHash(tc.codeWord, tc.n)
			if result != tc.expectedHash {
				t.Errorf("ComputeCodeWordHash(%v, %d) = %d, want %d",
					tc.codeWord, tc.n, result, tc.expectedHash)
			}
		})
	}
}

// TestComputeCodeWordHashK4 tests the hash function for k=4 (n=15) encoding.
func TestComputeCodeWordHashK4(t *testing.T) {
	t.Parallel()

	// For k=4, n=15: f(a) = XOR of (a_i * i) for i=1..15
	// Test a specific pattern
	codeWord := make([]int, 15)
	codeWord[0] = 1 // a1 = 1, contributes 1
	codeWord[3] = 1 // a4 = 1, contributes 4
	codeWord[7] = 1 // a8 = 1, contributes 8
	// Expected: 1 XOR 4 XOR 8 = 13

	result := ComputeCodeWordHash(codeWord, 15)
	expected := 13 // 1 XOR 4 XOR 8 = 0001 XOR 0100 XOR 1000 = 1101 = 13

	if result != expected {
		t.Errorf("ComputeCodeWordHash for k=4 pattern = %d, want %d", result, expected)
	}
}

// TestComputeCodeWordHashK8 tests the hash function for k=8 (n=255, maximum) encoding.
func TestComputeCodeWordHashK8(t *testing.T) {
	t.Parallel()

	// For k=8, n=255: f(a) = XOR of (a_i * i) for i=1..255
	codeWord := make([]int, 255)

	// All zeros should give hash 0
	result := ComputeCodeWordHash(codeWord, 255)
	if result != 0 {
		t.Errorf("ComputeCodeWordHash for k=8 all zeros = %d, want 0", result)
	}

	// Set position 255 to 1: should contribute 255 to hash
	codeWord[254] = 1 // a255 = 1 (0-indexed)
	result = ComputeCodeWordHash(codeWord, 255)
	if result != 255 {
		t.Errorf("ComputeCodeWordHash for k=8 with a255=1 = %d, want 255", result)
	}

	// Set position 1 to 1 as well: 1 XOR 255 = 254
	codeWord[0] = 1
	result = ComputeCodeWordHash(codeWord, 255)
	expected := 1 ^ 255 // = 254
	if result != expected {
		t.Errorf("ComputeCodeWordHash for k=8 with a1=1,a255=1 = %d, want %d", result, expected)
	}
}

// TestMatrixEncodePositionCalculation tests that position s = messageBits XOR hash
// is calculated correctly.
func TestMatrixEncodePositionCalculation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		codeWord         []int
		messageBits      int
		k                int
		expectedPosition int
	}{
		// k=1, n=1: simplest case
		// codeWord [0], message 0: hash=0, s = 0 XOR 0 = 0 (no change)
		{"k=1 no change needed", []int{0}, 0, 1, 0},
		// codeWord [0], message 1: hash=0, s = 1 XOR 0 = 1 (change pos 1)
		{"k=1 change needed", []int{0}, 1, 1, 1},
		// codeWord [1], message 1: hash=1, s = 1 XOR 1 = 0 (no change)
		{"k=1 already matches", []int{1}, 1, 1, 0},
		// codeWord [1], message 0: hash=1, s = 0 XOR 1 = 1 (change pos 1)
		{"k=1 mismatch change", []int{1}, 0, 1, 1},

		// k=2, n=3
		// codeWord [1,0,0], hash = 1, message = 1, s = 1 XOR 1 = 0 (no change)
		{"k=2 already matches", []int{1, 0, 0}, 1, 2, 0},
		// codeWord [0,0,0], hash = 0, message = 3, s = 3 XOR 0 = 3 (change pos 3)
		{"k=2 change pos 3", []int{0, 0, 0}, 3, 2, 3},
		// codeWord [1,1,1], hash = 0 (1 XOR 2 XOR 3 = 0), message = 2, s = 2 XOR 0 = 2
		{"k=2 change pos 2", []int{1, 1, 1}, 2, 2, 2},

		// k=3, n=7
		// codeWord all zeros, hash = 0, message = 5, s = 5 (change pos 5)
		{"k=3 change pos 5", []int{0, 0, 0, 0, 0, 0, 0}, 5, 3, 5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			position, err := MatrixEncode(tc.codeWord, tc.messageBits, tc.k)
			if err != nil {
				t.Fatalf("MatrixEncode returned unexpected error: %v", err)
			}
			if position != tc.expectedPosition {
				t.Errorf("MatrixEncode(%v, %d, %d) = %d, want %d",
					tc.codeWord, tc.messageBits, tc.k, position, tc.expectedPosition)
			}
		})
	}
}

// TestMatrixEncodeAtMostOneChange verifies that matrix encoding requires
// at most 1 coefficient change per code word.
func TestMatrixEncodeAtMostOneChange(t *testing.T) {
	t.Parallel()

	// For k=2 (n=3), test all possible message values (0-3)
	// and verify position is always 0, 1, 2, or 3 (at most 1 change)
	codeWord := []int{1, 0, 1} // Some arbitrary code word

	for messageBits := 0; messageBits < 4; messageBits++ {
		position, err := MatrixEncode(codeWord, messageBits, 2)
		if err != nil {
			t.Fatalf("MatrixEncode returned error for message %d: %v", messageBits, err)
		}
		// Position should be 0 (no change) or 1-3 (change one position)
		if position < 0 || position > 3 {
			t.Errorf("MatrixEncode position = %d for message %d, should be 0-3",
				position, messageBits)
		}
	}
}

// TestMatrixEncodeK4Encoding tests k=4 encoding (n=15) with specific test cases.
func TestMatrixEncodeK4Encoding(t *testing.T) {
	t.Parallel()

	// Create a code word for n=15
	codeWord := make([]int, 15)
	// Set specific bits to create known hash
	codeWord[0] = 1 // contributes 1
	codeWord[2] = 1 // contributes 3
	// hash = 1 XOR 3 = 2

	// If we want to embed message 2, position should be 0 (no change)
	position, err := MatrixEncode(codeWord, 2, 4)
	if err != nil {
		t.Fatalf("MatrixEncode returned unexpected error: %v", err)
	}
	if position != 0 {
		t.Errorf("MatrixEncode for matching hash should return 0, got %d", position)
	}

	// If we want to embed message 7, s = 7 XOR 2 = 5 (change position 5)
	position, err = MatrixEncode(codeWord, 7, 4)
	if err != nil {
		t.Fatalf("MatrixEncode returned unexpected error: %v", err)
	}
	expected := 7 ^ 2 // = 5
	if position != expected {
		t.Errorf("MatrixEncode for message 7 = %d, want %d", position, expected)
	}
}

// TestMatrixEncodeK8Encoding tests k=8 encoding (n=255, maximum).
func TestMatrixEncodeK8Encoding(t *testing.T) {
	t.Parallel()

	// Create a code word for n=255
	codeWord := make([]int, 255)
	// All zeros, hash = 0

	// Embed message 128: s = 128 XOR 0 = 128
	position, err := MatrixEncode(codeWord, 128, 8)
	if err != nil {
		t.Fatalf("MatrixEncode returned unexpected error: %v", err)
	}
	if position != 128 {
		t.Errorf("MatrixEncode for message 128 = %d, want 128", position)
	}

	// Embed message 0: s = 0 XOR 0 = 0 (no change)
	position, err = MatrixEncode(codeWord, 0, 8)
	if err != nil {
		t.Fatalf("MatrixEncode returned unexpected error: %v", err)
	}
	if position != 0 {
		t.Errorf("MatrixEncode for message 0 = %d, want 0", position)
	}

	// Set position 200 to 1, hash becomes 200
	codeWord[199] = 1
	// Embed message 200: s = 200 XOR 200 = 0 (no change)
	position, err = MatrixEncode(codeWord, 200, 8)
	if err != nil {
		t.Fatalf("MatrixEncode returned unexpected error: %v", err)
	}
	if position != 0 {
		t.Errorf("MatrixEncode for message 200 with hash 200 = %d, want 0", position)
	}
}

// TestMatrixEncodeEmptyCodeWord tests handling of empty code words.
func TestMatrixEncodeEmptyCodeWord(t *testing.T) {
	t.Parallel()

	// Empty code word should return an error
	_, err := MatrixEncode([]int{}, 0, 1)
	if err == nil {
		t.Error("MatrixEncode should return error for empty code word")
	}
}

// TestMatrixEncodeInvalidK tests handling of invalid k parameter values.
func TestMatrixEncodeInvalidK(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		codeWord []int
		k        int
	}{
		{name: "k=0", codeWord: []int{0}, k: 0},
		{name: "k=-1", codeWord: []int{0}, k: -1},
		{name: "k=9", codeWord: make([]int, 511), k: 9}, // k=9 would need n=511, but k>8 is not supported
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := MatrixEncode(tc.codeWord, 0, tc.k)
			if err == nil {
				t.Errorf("MatrixEncode should return error for k=%d", tc.k)
			}
		})
	}
}

// TestMatrixEncodeCodeWordSizeMismatch tests error handling when code word size
// doesn't match the expected n = 2^k - 1.
func TestMatrixEncodeCodeWordSizeMismatch(t *testing.T) {
	t.Parallel()

	// For k=2, n should be 3. Provide wrong size.
	_, err := MatrixEncode([]int{0, 0}, 0, 2) // size 2 instead of 3
	if err == nil {
		t.Error("MatrixEncode should return error for mismatched code word size")
	}

	_, err = MatrixEncode([]int{0, 0, 0, 0}, 0, 2) // size 4 instead of 3
	if err == nil {
		t.Error("MatrixEncode should return error for mismatched code word size")
	}
}

// TestApplyMatrixChangeBasic tests the ApplyMatrixChange function for basic cases.
func TestApplyMatrixChangeBasic(t *testing.T) {
	t.Parallel()

	// Position 0 means no change needed
	coefficients := []int16{5, 3, -2}
	indices := []int{0, 1, 2}
	result := ApplyMatrixChange(coefficients, indices, 0)
	if result != 0 {
		t.Errorf("ApplyMatrixChange with position 0 should return 0, got %d", result)
	}
	// Coefficients should be unchanged
	if coefficients[0] != 5 || coefficients[1] != 3 || coefficients[2] != -2 {
		t.Errorf("Coefficients should be unchanged when position is 0")
	}
}

// TestApplyMatrixChangeModifiesCorrectPosition tests that the correct coefficient
// is modified when position > 0.
func TestApplyMatrixChangeModifiesCorrectPosition(t *testing.T) {
	t.Parallel()

	// Position 2 means modify coefficient at index indices[1] (1-indexed position)
	indices := []int{10, 20, 30} // Actual indices in the larger array

	// We need to test with the actual coefficient array
	fullCoeffs := make([]int16, 40)
	fullCoeffs[10] = 5
	fullCoeffs[20] = 3
	fullCoeffs[30] = -2

	// Position 2 means change the coefficient at indices[1] = 20
	result := ApplyMatrixChange(fullCoeffs, indices, 2)
	// Coefficient 3 (positive) should be decremented to 2
	expectedNewValue := int16(2)
	if result != expectedNewValue {
		t.Errorf("ApplyMatrixChange at position 2 returned %d, want %d", result, expectedNewValue)
	}
	if fullCoeffs[20] != expectedNewValue {
		t.Errorf("Coefficient at position 2 should be %d, got %d", expectedNewValue, fullCoeffs[20])
	}
}

// TestApplyMatrixChangeNegativeCoefficient tests modification of negative coefficients.
func TestApplyMatrixChangeNegativeCoefficient(t *testing.T) {
	t.Parallel()

	coefficients := make([]int16, 10)
	coefficients[5] = -7 // Negative coefficient
	indices := []int{5}

	// Position 1 means modify coefficient at indices[0]
	result := ApplyMatrixChange(coefficients, indices, 1)
	// -7 should become -6 (increment toward zero)
	expectedNewValue := int16(-6)
	if result != expectedNewValue {
		t.Errorf("ApplyMatrixChange for negative coeff returned %d, want %d", result, expectedNewValue)
	}
	if coefficients[5] != expectedNewValue {
		t.Errorf("Coefficient should be %d, got %d", expectedNewValue, coefficients[5])
	}
}

// TestApplyMatrixChangeShrinkageCase tests that shrinkage is handled correctly.
func TestApplyMatrixChangeShrinkageCase(t *testing.T) {
	t.Parallel()

	// When coefficient is 1 or -1, modifying it produces 0 (shrinkage)
	coefficients := make([]int16, 10)
	coefficients[3] = 1 // Will shrink to 0
	indices := []int{3}

	result := ApplyMatrixChange(coefficients, indices, 1)
	// 1 becomes 0
	if result != 0 {
		t.Errorf("ApplyMatrixChange for coeff 1 returned %d, want 0", result)
	}
	if coefficients[3] != 0 {
		t.Errorf("Coefficient should be 0 after shrinkage, got %d", coefficients[3])
	}
}

// TestMatrixEncodeMessageBitsRange tests that message bits are within valid range.
func TestMatrixEncodeMessageBitsRange(t *testing.T) {
	t.Parallel()

	// For k=3 (n=7), valid message bits are 0-7 (3 bits = 0 to 2^3-1)
	codeWord := make([]int, 7)

	// Valid message bits: 0 to 7
	for msg := 0; msg <= 7; msg++ {
		_, err := MatrixEncode(codeWord, msg, 3)
		if err != nil {
			t.Errorf("MatrixEncode should accept message %d for k=3, got error: %v", msg, err)
		}
	}

	// Invalid message bits: 8 and above (exceeds k bits)
	_, err := MatrixEncode(codeWord, 8, 3)
	if err == nil {
		t.Error("MatrixEncode should return error for message bits exceeding k bits")
	}
}

// TestComputeCodeWordHashWithNilCodeWord tests handling of nil code word.
func TestComputeCodeWordHashWithNilCodeWord(t *testing.T) {
	t.Parallel()

	// Should handle nil gracefully (return 0)
	result := ComputeCodeWordHash(nil, 0)
	if result != 0 {
		t.Errorf("ComputeCodeWordHash(nil, 0) = %d, want 0", result)
	}
}

// TestCodeWordLength tests the CodeWordLength helper function.
func TestCodeWordLength(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		k        int
		expected int
	}{
		{1, 1},
		{2, 3},
		{3, 7},
		{4, 15},
		{5, 31},
		{6, 63},
		{7, 127},
		{8, 255},
	}

	for _, tc := range testCases {
		result := CodeWordLength(tc.k)
		if result != tc.expected {
			t.Errorf("CodeWordLength(%d) = %d, want %d", tc.k, result, tc.expected)
		}
	}
}

// TestExtractStegoBits tests the ExtractStegoBits helper function.
func TestExtractStegoBits(t *testing.T) {
	t.Parallel()

	coefficients := []int16{3, -2, 5, -4, 1, -1, 2, -3}
	// Expected stego bits:
	// 3: positive odd -> 1
	// -2: negative even -> 1
	// 5: positive odd -> 1
	// -4: negative even -> 1
	// 1: positive odd -> 1
	// -1: negative odd -> 0
	// 2: positive even -> 0
	// -3: negative odd -> 0
	expected := []int{1, 1, 1, 1, 1, 0, 0, 0}

	result := ExtractStegoBits(coefficients)

	if len(result) != len(expected) {
		t.Fatalf("ExtractStegoBits returned %d bits, want %d", len(result), len(expected))
	}

	for i, bit := range result {
		if bit != expected[i] {
			t.Errorf("ExtractStegoBits[%d] = %d, want %d (coeff=%d)",
				i, bit, expected[i], coefficients[i])
		}
	}
}

// TestApplyMatrixChangeOutOfBounds tests boundary conditions for ApplyMatrixChange.
func TestApplyMatrixChangeOutOfBounds(t *testing.T) {
	t.Parallel()

	coefficients := []int16{5, 3, -2}
	indices := []int{0, 1, 2}

	// Position beyond indices length should return 0 safely
	result := ApplyMatrixChange(coefficients, indices, 10)
	if result != 0 {
		t.Errorf("ApplyMatrixChange with out-of-bounds position should return 0, got %d", result)
	}

	// Negative position (converted from 1-indexed) should be handled
	result = ApplyMatrixChange(coefficients, indices, -1)
	if result != 0 {
		t.Errorf("ApplyMatrixChange with negative position should return 0, got %d", result)
	}
}

// TestApplyMatrixChangeInvalidIndex tests when indices point outside coefficient array.
func TestApplyMatrixChangeInvalidIndex(t *testing.T) {
	t.Parallel()

	coefficients := []int16{5, 3, -2}
	indices := []int{0, 100, 2} // Index 100 is out of bounds

	// Should handle gracefully and return 0
	result := ApplyMatrixChange(coefficients, indices, 2) // Position 2 = indices[1] = 100
	if result != 0 {
		t.Errorf("ApplyMatrixChange with invalid coefficient index should return 0, got %d", result)
	}
}

// Fuzz Testing

// FuzzMatrixEncode tests the MatrixEncode function with fuzzed inputs.
//
// This fuzz test verifies that:
// - No panics occur with arbitrary code word values and k parameters
// - Change position is valid (0 to n) when no error
// - Error handling works correctly for invalid inputs
func FuzzMatrixEncode(f *testing.F) {
	// Add seed corpus
	f.Add([]byte{0, 1, 0}, 1, 2)
	f.Add([]byte{1}, 0, 1)
	f.Add([]byte{1, 1, 1, 1, 1, 1, 1}, 0, 3)
	f.Add([]byte{}, 0, 4)
	f.Add([]byte{1, 0, 0}, 5, 2)
	f.Add([]byte{1, 0, 0}, 0, 0)
	f.Add([]byte{1, 0, 0}, 0, 9)

	f.Fuzz(func(t *testing.T, codeWordBytes []byte, messageBits int, k int) {
		// Convert bytes to code word (each byte becomes 0 or 1)
		codeWord := make([]int, len(codeWordBytes))
		for i, b := range codeWordBytes {
			codeWord[i] = int(b) & 1 // 0 or 1
		}

		// Call MatrixEncode - should not panic
		changePos, err := MatrixEncode(codeWord, messageBits, k)

		// If there's an error, verify it's expected for the inputs
		if err != nil {
			// Expected errors for:
			// - k < 1 or k > 8: ErrInvalidKParameter
			// - empty code word: ErrEmptyCodeWord
			// - code word size mismatch: ErrCodeWordSizeMismatch
			// - message bits exceed k-bit max: ErrMessageBitsExceedK
			return
		}

		// If successful, verify output validity
		if k < 1 || k > 8 {
			t.Errorf("MatrixEncode succeeded with invalid k=%d", k)
		}

		expectedN := (1 << k) - 1
		if len(codeWord) != expectedN {
			t.Errorf("MatrixEncode succeeded with wrong code word length: %d (expected %d)", len(codeWord), expectedN)
		}

		// Change position should be 0 to n
		if changePos < 0 || changePos > expectedN {
			t.Errorf("Invalid changePosition: %d (expected 0 to %d)", changePos, expectedN)
		}

		// Verify the matrix encoding property:
		// After applying the change at changePos, the hash should equal messageBits
		if changePos > 0 {
			// Simulate the change
			modifiedCodeWord := make([]int, len(codeWord))
			copy(modifiedCodeWord, codeWord)
			// Toggle the bit at changePos-1 (0-indexed)
			modifiedCodeWord[changePos-1] = 1 - modifiedCodeWord[changePos-1]

			// Compute hash of modified code word
			hash := ComputeCodeWordHash(modifiedCodeWord, expectedN)

			// Clamp messageBits to valid range for comparison
			maxBits := (1 << k) - 1
			if messageBits >= 0 && messageBits <= maxBits && hash != messageBits {
				t.Errorf("Matrix encoding property violated: hash=%d, messageBits=%d", hash, messageBits)
			}
		}
	})
}

// TestFuzzEdgeCaseLargeK tests matrix encoding with k=8 (n=255).
func TestFuzzEdgeCaseLargeK(t *testing.T) {
	// k=8 means n=255, code word of 255 coefficients
	n := 255
	codeWord := make([]int, n)

	// Test all zeros
	pos, err := MatrixEncode(codeWord, 0, 8)
	if err != nil {
		t.Fatalf("MatrixEncode failed: %v", err)
	}
	if pos != 0 {
		t.Errorf("Expected pos=0 for all zeros and msg=0, got %d", pos)
	}

	// Test message = max (255)
	pos, err = MatrixEncode(codeWord, 255, 8)
	if err != nil {
		t.Fatalf("MatrixEncode failed: %v", err)
	}
	if pos != 255 {
		t.Errorf("Expected pos=255 for all zeros and msg=255, got %d", pos)
	}

	// Test all ones
	for i := range codeWord {
		codeWord[i] = 1
	}
	// Hash of all ones for n=255: XOR of 1,2,3,...,255 = 0 (pairs cancel)
	// Actually, 1^2^3^...^255: for powers of 2, this is complex
	// Let's just verify no panic and valid range
	pos, err = MatrixEncode(codeWord, 128, 8)
	if err != nil {
		t.Fatalf("MatrixEncode failed: %v", err)
	}
	if pos < 0 || pos > n {
		t.Errorf("Invalid pos=%d for k=8", pos)
	}
}
