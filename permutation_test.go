package f5messageembed

import (
	"fmt"
	"testing"

	"github.com/0verkilll/f5prng"
	"github.com/0verkilll/fisheryates"
)

// TestInitializePRNG_DeterministicFromPassword verifies that the same password
// always produces the same PRNG state.
func TestInitializePRNG_DeterministicFromPassword(t *testing.T) {
	password := "test_password_123"

	// Initialize PRNG twice with the same password
	prng1 := InitializePRNG(password)
	prng2 := InitializePRNG(password)

	// Generate bytes from both - should be identical
	bytes1 := prng1.NextBytes(20)
	bytes2 := prng2.NextBytes(20)

	if len(bytes1) != len(bytes2) {
		t.Fatalf("expected same length, got %d and %d", len(bytes1), len(bytes2))
	}

	for i := range bytes1 {
		if bytes1[i] != bytes2[i] {
			t.Fatalf("byte mismatch at position %d: got %d and %d", i, bytes1[i], bytes2[i])
		}
	}
}

// TestInitializePRNG_DifferentPasswords verifies that different passwords
// produce different PRNG output.
func TestInitializePRNG_DifferentPasswords(t *testing.T) {
	prng1 := InitializePRNG("password1")
	prng2 := InitializePRNG("password2")

	bytes1 := prng1.NextBytes(20)
	bytes2 := prng2.NextBytes(20)

	// At least some bytes should differ
	sameCount := 0
	for i := range bytes1 {
		if bytes1[i] == bytes2[i] {
			sameCount++
		}
	}

	// Extremely unlikely for all bytes to be identical with different passwords
	if sameCount == len(bytes1) {
		t.Error("different passwords produced identical PRNG output")
	}
}

// TestGeneratePermutation_SizeMatchesCoefficients verifies that the generated
// permutation has exactly the requested size.
func TestGeneratePermutation_SizeMatchesCoefficients(t *testing.T) {
	testCases := []int{0, 1, 64, 100, 1000, 10000}

	for _, size := range testCases {
		t.Run(fmt.Sprintf("size_%d", size), func(t *testing.T) {
			prng := InitializePRNG("test_password")
			perm, err := GeneratePermutation(prng, size)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(perm) != size {
				t.Errorf("expected permutation size %d, got %d", size, len(perm))
			}
		})
	}
}

// TestGeneratePermutation_DeterministicFromPRNG verifies that the same PRNG
// state produces the same permutation.
func TestGeneratePermutation_DeterministicFromPRNG(t *testing.T) {
	password := "test_password_xyz"
	size := 1000

	// Initialize two PRNGs with the same password
	prng1 := InitializePRNG(password)
	prng2 := InitializePRNG(password)

	// Generate permutations
	perm1, err1 := GeneratePermutation(prng1, size)
	perm2, err2 := GeneratePermutation(prng2, size)

	if err1 != nil || err2 != nil {
		t.Fatalf("unexpected errors: %v, %v", err1, err2)
	}

	// Permutations should be identical
	for i := range perm1 {
		if perm1[i] != perm2[i] {
			t.Fatalf("permutation mismatch at position %d: got %d and %d", i, perm1[i], perm2[i])
		}
	}
}

// TestGeneratePermutation_ContainsAllIndices verifies that the permutation
// contains each index exactly once.
func TestGeneratePermutation_ContainsAllIndices(t *testing.T) {
	prng := InitializePRNG("test_password")
	size := 100
	perm, err := GeneratePermutation(prng, size)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that each index 0 to size-1 appears exactly once
	seen := make(map[int]int)
	for i, v := range perm {
		if v < 0 || v >= size {
			t.Errorf("index out of range at position %d: %d", i, v)
		}
		seen[v]++
	}

	if len(seen) != size {
		t.Errorf("expected %d unique values, got %d", size, len(seen))
	}

	for i := 0; i < size; i++ {
		if seen[i] != 1 {
			t.Errorf("index %d appears %d times (expected 1)", i, seen[i])
		}
	}
}

// TestGeneratePermutation_ExceedsMaxSize verifies that requesting a permutation
// larger than MaxPermutationSize returns an error.
func TestGeneratePermutation_ExceedsMaxSize(t *testing.T) {
	prng := InitializePRNG("test_password")

	// Request a permutation larger than the maximum
	_, err := GeneratePermutation(prng, fisheryates.MaxPermutationSize+1)

	if err == nil {
		t.Error("expected error for size exceeding MaxPermutationSize, got nil")
	}
}

// TestApplyDeZigZag_CorrectTransformation verifies the de-zigzag transformation
// produces the expected results based on the actual deZigZag table.
func TestApplyDeZigZag_CorrectTransformation(t *testing.T) {
	// Build test cases from the actual deZigZag table
	testCases := []struct {
		name           string
		shuffledIndex  int
		expectedZigzag int
	}{
		// First 8x8 block - use actual deZigZag values
		{"block0_pos0", 0, deZigZag[0]},
		{"block0_pos1", 1, deZigZag[1]},
		{"block0_pos2", 2, deZigZag[2]},
		{"block0_pos8", 8, deZigZag[8]},
		{"block0_pos63", 63, deZigZag[63]},

		// Second 8x8 block (indices 64-127): blockBase + deZigZag[posInBlock]
		{"block1_pos64", 64, 64 + deZigZag[0]},
		{"block1_pos65", 65, 64 + deZigZag[1]},
		{"block1_pos72", 72, 64 + deZigZag[8]},

		// Third 8x8 block (indices 128-191)
		{"block2_pos128", 128, 128 + deZigZag[0]},
		{"block2_pos130", 130, 128 + deZigZag[2]},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ApplyDeZigZag(tc.shuffledIndex)
			if result != tc.expectedZigzag {
				t.Errorf("ApplyDeZigZag(%d): expected %d, got %d",
					tc.shuffledIndex, tc.expectedZigzag, result)
			}
		})
	}
}

// TestApplyDeZigZag_Formula verifies the de-zigzag formula:
// zigzag = shuffled - shuffled%64 + deZigZag[shuffled%64]
func TestApplyDeZigZag_Formula(t *testing.T) {
	// Test across multiple blocks
	for blockNum := 0; blockNum < 5; blockNum++ {
		blockBase := blockNum * 64
		for posInBlock := 0; posInBlock < 64; posInBlock++ {
			shuffled := blockBase + posInBlock
			expected := blockBase + deZigZag[posInBlock]
			result := ApplyDeZigZag(shuffled)

			if result != expected {
				t.Errorf("block %d, pos %d: ApplyDeZigZag(%d) = %d, expected %d",
					blockNum, posInBlock, shuffled, result, expected)
			}
		}
	}
}

// TestApplyDeZigZag_BlockBoundaries verifies correct behavior at block boundaries.
func TestApplyDeZigZag_BlockBoundaries(t *testing.T) {
	// Test each block boundary
	boundaries := []int{0, 63, 64, 127, 128, 191, 192, 255}

	for _, idx := range boundaries {
		blockBase := idx - idx%64
		posInBlock := idx % 64
		expected := blockBase + deZigZag[posInBlock]
		result := ApplyDeZigZag(idx)

		if result != expected {
			t.Errorf("boundary index %d: ApplyDeZigZag = %d, expected %d",
				idx, result, expected)
		}
	}
}

// TestPRNGConsumptionOrder verifies that PRNG bytes are consumed in the
// correct order matching Java F5 implementation.
//
// Order: seed -> permutation -> header XOR bytes -> message XOR bytes
func TestPRNGConsumptionOrder(t *testing.T) {
	password := "test_order"
	coeffCount := 64 // Small coefficient count for testing

	// Initialize PRNG
	prng := InitializePRNG(password)

	// Step 1: Generate permutation - this consumes PRNG state
	_, err := GeneratePermutation(prng, coeffCount)
	if err != nil {
		t.Fatalf("failed to generate permutation: %v", err)
	}

	// Step 2: Get 4 bytes for header XOR - this is the state after permutation
	headerXorBytes := prng.NextBytes(4)

	// Now verify with a fresh PRNG that we get the same results
	prng2 := InitializePRNG(password)
	_, err = GeneratePermutation(prng2, coeffCount)
	if err != nil {
		t.Fatalf("failed to generate second permutation: %v", err)
	}
	headerXorBytes2 := prng2.NextBytes(4)

	for i := range headerXorBytes {
		if headerXorBytes[i] != headerXorBytes2[i] {
			t.Errorf("header XOR byte %d mismatch: %d vs %d",
				i, headerXorBytes[i], headerXorBytes2[i])
		}
	}
}

// TestPRNGConsumptionMatchesExtraction verifies that embedding PRNG consumption
// matches the extraction pattern using f5prng.
// This test validates that InitializePRNG produces the same PRNG behavior as
// using f5prng.NewDefaultFactory directly.
func TestPRNGConsumptionMatchesExtraction(t *testing.T) {
	password := "match_test_password"
	coeffCount := 1000

	// Create PRNG for embedding using InitializePRNG
	embedPRNG := InitializePRNG(password)

	// Generate permutation (same as extraction)
	perm1, err := GeneratePermutation(embedPRNG, coeffCount)
	if err != nil {
		t.Fatalf("failed to generate permutation: %v", err)
	}

	// Independently create PRNG using f5prng factory
	// (simulates what extraction would do with the same password). Like the
	// embed/extract code paths, seed with the RAW password bytes to match the
	// real Westfeld f5.jar (new SecureRandom(password.getBytes())); Seed applies
	// SHA-1 internally, giving state = SHA-1(password).
	factory := f5prng.NewDefaultFactory()
	random := factory.NewPRNG()
	if seedErr := random.Seed([]byte(password)); seedErr != nil {
		t.Fatalf("seed: %v", seedErr)
	}

	permutator := fisheryates.NewFisherYates()
	perm2, err := permutator.Generate(coeffCount, random)
	if err != nil {
		t.Fatalf("failed to generate comparison permutation: %v", err)
	}

	// Verify permutations match
	for i := range perm1 {
		if perm1[i] != perm2[i] {
			t.Fatalf("permutation mismatch at index %d: %d vs %d", i, perm1[i], perm2[i])
		}
	}

	// Verify PRNG state after permutation matches
	// Get 4 bytes from each PRNG (header XOR bytes)
	embedHeaderBytes := embedPRNG.NextBytes(4)
	extractHeaderBytes := random.NextBytes(4)

	for i := range embedHeaderBytes {
		if embedHeaderBytes[i] != extractHeaderBytes[i] {
			t.Errorf("PRNG state mismatch at byte %d after permutation: %d vs %d",
				i, embedHeaderBytes[i], extractHeaderBytes[i])
		}
	}
}
