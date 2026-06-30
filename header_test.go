package f5messageembed

import (
	"testing"
)

// TestBuildHeader_Format tests the header format: k (bits 24-31) + fileSize (bits 0-22).
func TestBuildHeader_Format(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		k           int
		messageSize int
		expectedK   int
		expectedLen int
	}{
		{"k=1 small message", 1, 100, 1, 100},
		{"k=4 medium message", 4, 1000, 4, 1000},
		{"k=8 large message", 8, 100000, 8, 100000},
		{"k=1 max message size", 1, MaxMessageSize, 1, MaxMessageSize},
		{"k=8 zero message", 8, 0, 8, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			header := BuildHeader(tc.k, tc.messageSize)

			// Extract k from bits 24-31
			extractedK := int((header >> 24) & 0xFF)
			if extractedK != tc.expectedK {
				t.Errorf("BuildHeader(%d, %d): extracted k = %d, want %d",
					tc.k, tc.messageSize, extractedK, tc.expectedK)
			}

			// Extract fileSize from bits 0-22
			extractedLen := int(header & 0x7FFFFF)
			if extractedLen != tc.expectedLen {
				t.Errorf("BuildHeader(%d, %d): extracted fileSize = %d, want %d",
					tc.k, tc.messageSize, extractedLen, tc.expectedLen)
			}
		})
	}
}

// TestBuildHeader_KParameter tests that k parameter is stored correctly in bits 24-31.
func TestBuildHeader_KParameter(t *testing.T) {
	t.Parallel()

	// Test all valid k values (1-8)
	for k := 1; k <= 8; k++ {
		t.Run("k parameter storage", func(t *testing.T) {
			t.Parallel()
			header := BuildHeader(k, 1000)

			// k should be in bits 24-31
			extractedK := int((header >> 24) & 0xFF)
			if extractedK != k {
				t.Errorf("BuildHeader(%d, 1000): k = %d, want %d", k, extractedK, k)
			}

			// fileSize should not be affected
			extractedLen := int(header & 0x7FFFFF)
			if extractedLen != 1000 {
				t.Errorf("BuildHeader(%d, 1000): fileSize = %d, want 1000", k, extractedLen)
			}
		})
	}
}

// TestXORHeaderWithPRNG_XORsWith4Bytes tests that header is XORed with 4 PRNG bytes.
func TestXORHeaderWithPRNG_XORsWith4Bytes(t *testing.T) {
	t.Parallel()

	password := "test_header_xor"
	prng := InitializePRNG(password)

	// Build a header
	header := BuildHeader(4, 1000)

	// Get 4 PRNG bytes manually to calculate expected result
	prng2 := InitializePRNG(password)
	xorBytes := prng2.NextBytes(4)

	// Calculate expected XOR result
	// Header is XORed byte-by-byte with PRNG bytes
	// Java order: byte 0 XOR header bits 0-7, byte 1 XOR header bits 8-15, etc.
	expectedHeader := header
	expectedHeader ^= int32(int8(xorBytes[0]))       // bits 0-7
	expectedHeader ^= int32(int8(xorBytes[1])) << 8  // bits 8-15
	expectedHeader ^= int32(int8(xorBytes[2])) << 16 // bits 16-23
	expectedHeader ^= int32(int8(xorBytes[3])) << 24 // bits 24-31

	// XOR header using function
	result := XORHeaderWithPRNG(header, prng)

	if result != expectedHeader {
		t.Errorf("XORHeaderWithPRNG: got 0x%08X, want 0x%08X", result, expectedHeader)
	}
}

// TestXORHeaderWithPRNG_Deterministic tests that the same password produces the same XOR result.
func TestXORHeaderWithPRNG_Deterministic(t *testing.T) {
	t.Parallel()

	password := "deterministic_test"
	header := BuildHeader(5, 5000)

	// XOR with first PRNG
	prng1 := InitializePRNG(password)
	result1 := XORHeaderWithPRNG(header, prng1)

	// XOR with second PRNG (same password)
	prng2 := InitializePRNG(password)
	result2 := XORHeaderWithPRNG(header, prng2)

	if result1 != result2 {
		t.Errorf("XORHeaderWithPRNG not deterministic: 0x%08X vs 0x%08X", result1, result2)
	}
}

// TestXORHeaderWithPRNG_SignedByteArithmetic tests Java-compatible signed byte arithmetic.
func TestXORHeaderWithPRNG_SignedByteArithmetic(t *testing.T) {
	t.Parallel()

	// Use a password that produces bytes with high bit set (negative when signed)
	password := "signed_byte_test_password_xyz"
	header := BuildHeader(3, 3000)

	prng1 := InitializePRNG(password)
	result := XORHeaderWithPRNG(header, prng1)

	// Verify by manually computing with signed byte conversion
	prng2 := InitializePRNG(password)
	bytes := prng2.NextBytes(4)

	// Java-compatible signed byte XOR
	expected := header
	expected ^= int32(int8(bytes[0]))
	expected ^= int32(int8(bytes[1])) << 8
	expected ^= int32(int8(bytes[2])) << 16
	expected ^= int32(int8(bytes[3])) << 24

	if result != expected {
		t.Errorf("signed byte arithmetic mismatch: got 0x%08X, want 0x%08X", result, expected)
	}
}

// TestEmbedHeader_SimpleLSB tests that header is embedded using simple LSB (not matrix encoding).
func TestEmbedHeader_SimpleLSB(t *testing.T) {
	t.Parallel()

	// Create test coefficients with known values
	// Need enough non-zero, non-DC coefficients to embed 32 bits
	coefficients := make([]int16, 256)
	for i := range coefficients {
		if i%64 != 0 { // Not DC coefficient
			coefficients[i] = 5 // Non-zero, odd (stego bit = 1)
		}
	}

	// Create a permutation (identity for simplicity in testing)
	permutation := make([]int, len(coefficients))
	for i := range permutation {
		permutation[i] = i
	}

	// Build and embed header with all zero bits
	header := int32(0) // All stego bits should become 0

	nextIndex, err := EmbedHeader(coefficients, permutation, header)
	if err != nil {
		t.Fatalf("EmbedHeader failed: %v", err)
	}

	// Verify that 32 bits were embedded
	bitsEmbedded := 0
	for i := 0; i < nextIndex; i++ {
		shuffled := permutation[i]
		if shuffled%64 == 0 {
			continue // Skip DC
		}
		zigzag := ApplyDeZigZag(shuffled)
		if coefficients[zigzag] != 0 {
			bitsEmbedded++
		}
	}

	// We should have used 32 coefficients for 32 bits
	if bitsEmbedded < 32 {
		t.Errorf("expected at least 32 bits embedded, got %d", bitsEmbedded)
	}
}

// TestEmbedHeader_SkipsDCAndZero tests that DC and zero coefficients are skipped.
func TestEmbedHeader_SkipsDCAndZero(t *testing.T) {
	t.Parallel()

	// Create coefficients with DC and zeros that should be skipped
	coefficients := make([]int16, 256)
	for i := range coefficients {
		switch {
		case i%64 == 0:
			coefficients[i] = 100 // DC coefficient - should be skipped
		case i%3 == 0:
			coefficients[i] = 0 // Zero - should be skipped
		default:
			coefficients[i] = 5 // Usable
		}
	}

	// Identity permutation
	permutation := make([]int, len(coefficients))
	for i := range permutation {
		permutation[i] = i
	}

	// Pattern with alternating bits using negative value to fit in int32
	// 0xAAAAAAAA as signed int32 is -1431655766
	header := int32(-1431655766)

	_, err := EmbedHeader(coefficients, permutation, header)
	if err != nil {
		t.Fatalf("EmbedHeader failed: %v", err)
	}

	// Verify DC coefficients were NOT modified
	for i := 0; i < len(coefficients); i += 64 {
		if coefficients[i] != 100 {
			t.Errorf("DC coefficient at index %d was modified: got %d, want 100", i, coefficients[i])
		}
	}
}

// TestHeaderRoundTrip_EmbedThenExtract tests that embedded header can be extracted correctly.
func TestHeaderRoundTrip_EmbedThenExtract(t *testing.T) {
	t.Parallel()

	// Create coefficients
	coefficients := make([]int16, 512)
	for i := range coefficients {
		if i%64 != 0 { // Not DC
			// Mix of odd and even values
			if i%2 == 0 {
				coefficients[i] = 4 // even = stego bit 0 (positive)
			} else {
				coefficients[i] = 3 // odd = stego bit 1 (positive)
			}
		}
	}

	// Create a permutation
	permutation := make([]int, len(coefficients))
	for i := range permutation {
		permutation[i] = i
	}

	// Build header: k=5, messageSize=12345
	originalHeader := BuildHeader(5, 12345)

	// Embed header
	nextIndex, err := EmbedHeader(coefficients, permutation, originalHeader)
	if err != nil {
		t.Fatalf("EmbedHeader failed: %v", err)
	}

	// Extract header (simulate extraction)
	extractedHeader := int32(0)
	bitsExtracted := 0
	coeffIndex := 0

	for bitsExtracted < 32 && coeffIndex < len(permutation) {
		shuffled := permutation[coeffIndex]
		coeffIndex++

		// Skip DC
		if shuffled%64 == 0 {
			continue
		}

		// Apply de-zigzag
		zigzag := ApplyDeZigZag(shuffled)
		if zigzag >= len(coefficients) {
			continue
		}

		coefficient := coefficients[zigzag]
		if coefficient == 0 {
			continue
		}

		// Extract bit (same as GetStegoBit)
		bit := GetStegoBit(coefficient)
		extractedHeader |= int32(bit) << bitsExtracted
		bitsExtracted++
	}

	if extractedHeader != originalHeader {
		t.Errorf("header round-trip failed: embedded 0x%08X, extracted 0x%08X",
			originalHeader, extractedHeader)
	}

	// Verify nextIndex is reasonable
	if nextIndex <= 32 {
		t.Logf("nextIndex = %d (consumed %d permutation entries for 32 bits)", nextIndex, nextIndex)
	}
}

// TestHeaderRoundTrip_WithRandomPermutation tests header embedding with shuffled permutation.
func TestHeaderRoundTrip_WithRandomPermutation(t *testing.T) {
	t.Parallel()

	password := "round_trip_test"

	// Create coefficients
	coefficients := make([]int16, 1024)
	for i := range coefficients {
		if i%64 != 0 { // Not DC
			coefficients[i] = int16((i%10)+1) * 2 // Even values (stego bit 0)
		}
	}

	// Generate real permutation
	prng := InitializePRNG(password)
	permutation, err := GeneratePermutation(prng, len(coefficients))
	if err != nil {
		t.Fatalf("GeneratePermutation failed: %v", err)
	}

	// Build and XOR header
	originalHeader := BuildHeader(3, 54321)
	xoredHeader := XORHeaderWithPRNG(originalHeader, prng)

	// Embed header
	nextIndex, err := EmbedHeader(coefficients, permutation, xoredHeader)
	if err != nil {
		t.Fatalf("EmbedHeader failed: %v", err)
	}

	// Extract header using same permutation
	extractedXoredHeader := int32(0)
	bitsExtracted := 0
	coeffIndex := 0

	for bitsExtracted < 32 && coeffIndex < len(permutation) {
		shuffled := permutation[coeffIndex]
		coeffIndex++

		if shuffled%64 == 0 {
			continue
		}

		zigzag := ApplyDeZigZag(shuffled)
		if zigzag >= len(coefficients) {
			continue
		}

		coefficient := coefficients[zigzag]
		if coefficient == 0 {
			continue
		}

		bit := GetStegoBit(coefficient)
		extractedXoredHeader |= int32(bit) << bitsExtracted
		bitsExtracted++
	}

	if extractedXoredHeader != xoredHeader {
		t.Errorf("header extraction failed: embedded 0x%08X, extracted 0x%08X",
			xoredHeader, extractedXoredHeader)
	}

	// Verify we advanced past the header
	if nextIndex != coeffIndex {
		t.Errorf("nextIndex mismatch: EmbedHeader returned %d, extraction used %d",
			nextIndex, coeffIndex)
	}
}

// TestEmbedHeader_InsufficientCoefficients tests error when not enough coefficients.
func TestEmbedHeader_InsufficientCoefficients(t *testing.T) {
	t.Parallel()

	// Create very small coefficient array
	coefficients := make([]int16, 32)
	for i := range coefficients {
		coefficients[i] = 5
	}

	permutation := make([]int, len(coefficients))
	for i := range permutation {
		permutation[i] = i
	}

	header := BuildHeader(4, 1000)

	// Should fail because many coefficients are DC (index%64==0) or we run out
	_, err := EmbedHeader(coefficients, permutation, header)

	// With only 32 coefficients and some being DC, this should fail
	// Actually index 0 is DC, so we have 31 usable. Need 32 bits.
	if err == nil {
		t.Error("expected error for insufficient coefficients, got nil")
	}
}
