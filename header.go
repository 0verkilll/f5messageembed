package f5messageembed

import "github.com/0verkilll/f5prng"

// BuildHeader constructs a 32-bit F5 header from k parameter and message size.
//
// The classic F5 header format is:
//   - Bits 24-31 (8 bits): k parameter value (matrix encoding parameter)
//   - Bits 0-22 (23 bits): message file size in bytes
//
// This header format allows messages up to 8,388,607 bytes (2^23 - 1) and
// k values from 0-255 (though only 1-8 are typically used).
//
// Parameters:
//   - k: The matrix encoding parameter (typically 1-8)
//   - messageSize: The message size in bytes (max MaxMessageSize)
//
// Returns:
//   - A 32-bit header value ready for XOR and embedding
//
// Example:
//
//	header := BuildHeader(4, 1000)
//	// header bits 24-31 = 4 (k parameter)
//	// header bits 0-22 = 1000 (message size)
func BuildHeader(k, messageSize int) int32 {
	// k goes in bits 24-31 (shift by 24)
	// messageSize goes in bits 0-22 (mask to 23 bits for safety)
	// #nosec G115 -- Intentional int->int32 conversion for header format, values are bounded by F5 spec
	return (int32(k) << 24) | (int32(messageSize) & 0x7FFFFF) //nolint:gosec // G115: values bounded by F5 spec
}

// XORHeaderWithPRNG XORs a header with 4 consecutive PRNG bytes.
//
// The F5 algorithm XORs the header with PRNG bytes to add another layer
// of password-dependent obfuscation. This must be done in the exact same
// order as the Java implementation for byte-identical compatibility.
//
// PRNG consumption order (Java-compatible):
//   - Byte 0: XORed with header bits 0-7
//   - Byte 1: XORed with header bits 8-15
//   - Byte 2: XORed with header bits 16-23
//   - Byte 3: XORed with header bits 24-31
//
// Important: Uses signed byte arithmetic to match Java's behavior.
// Go's bytes are unsigned [0, 255], but Java's bytes are signed [-128, 127].
// The int(int8(byte)) conversion ensures sign extension matches Java.
//
// Parameters:
//   - header: The 32-bit header value from BuildHeader
//   - random: A seeded RandomSource (PRNG state after permutation generation)
//
// Returns:
//   - The XORed header value ready for embedding
//
// Example:
//
//	prng := InitializePRNG(password)
//	_, _ = GeneratePermutation(prng, coeffCount) // consumes PRNG for permutation
//	header := BuildHeader(k, len(message))
//	xoredHeader := XORHeaderWithPRNG(header, prng) // consumes 4 more PRNG bytes
func XORHeaderWithPRNG(header int32, random RandomSource) int32 {
	// Fetch 4 consecutive PRNG bytes. Prefer the zero-alloc NextBytesInto fast
	// path when the concrete PRNG supports it; otherwise fall back to the
	// NextBytes(4) allocating path. Both paths consume PRNG state identically,
	// preserving Java F5 parity.
	var xorBytes [4]byte
	if into, ok := random.(f5prng.RandomSourceWithBytesInto); ok {
		// Error intentionally discarded: failures are retained on the PRNG's
		// LastError and surface on the next draw, matching the NextBytes(4)
		// fallback path which has no error-return surface at all.
		_ = into.NextBytesInto(xorBytes[:]) //nolint:errcheck // see comment above
	} else {
		copy(xorBytes[:], random.NextBytes(4))
	}

	// Signed byte conversion is load-bearing here: Java's byte is signed
	// [-128, 127] and the XOR mixes the sign-extended value into the int32.
	// Rewriting as unsigned would produce a different output stream and break
	// byte-for-byte parity with the Java reference implementation.
	result := header
	result ^= int32(int8(xorBytes[0]))       //nolint:gosec // G115: Java-compat signed byte (bits 0-7)
	result ^= int32(int8(xorBytes[1])) << 8  //nolint:gosec // G115: Java-compat signed byte (bits 8-15)
	result ^= int32(int8(xorBytes[2])) << 16 //nolint:gosec // G115: Java-compat signed byte (bits 16-23)
	result ^= int32(int8(xorBytes[3])) << 24 //nolint:gosec // G115: Java-compat signed byte (bits 24-31)

	return result
}

// EmbedHeader embeds a 32-bit header into coefficients using simple LSB embedding.
//
// Unlike the message payload which uses matrix encoding, the F5 header is embedded
// using simple LSB modification. This is because the header must be extracted first
// to determine the k parameter needed for matrix decoding.
//
// The function:
//  1. Iterates through permuted coefficient indices
//  2. Skips DC coefficients (index % 64 == 0)
//  3. Applies de-zigzag transformation
//  4. Skips zero coefficients
//  5. Embeds one header bit per usable coefficient
//  6. Modifies coefficient if current stego bit doesn't match desired bit
//
// Parameters:
//   - coefficients: The DCT coefficient array to modify (modified in-place)
//   - permutation: The Fisher-Yates permutation of coefficient indices
//   - header: The 32-bit XORed header value to embed
//
// Returns:
//   - nextIndex: The permutation index to resume from for message embedding
//   - err: An error if insufficient coefficients for header embedding
//
// Example:
//
//	nextIndex, err := EmbedHeader(coefficients, permutation, xoredHeader)
//	if err != nil {
//	    return err
//	}
//	// Continue embedding message starting at permutation[nextIndex]
func EmbedHeader(coefficients []int16, permutation []int, header int32) (nextIndex int, err error) {
	bitsEmbedded := 0
	coeffIndex := 0

	for bitsEmbedded < HeaderSize && coeffIndex < len(permutation) {
		// Get the permuted index
		shuffled := permutation[coeffIndex]
		coeffIndex++

		// Skip DC coefficients (first coefficient of each 8x8 block). Using
		// `shuffled & 63` — BlockSize is a power of two, asserted at compile
		// time in f5coefficient — avoids IDIV in the per-bit header loop.
		if shuffled&63 == 0 {
			continue
		}

		// Apply de-zigzag transformation
		zigzag := ApplyDeZigZag(shuffled)

		// Bounds check
		if zigzag >= len(coefficients) {
			continue
		}

		// Get the coefficient
		coefficient := coefficients[zigzag]

		// Skip zero coefficients
		if coefficient == 0 {
			continue
		}

		// Extract the bit we want to embed (LSB-first order)
		desiredBit := int((header >> bitsEmbedded) & 1)

		// Get current stego bit of the coefficient
		currentBit := GetStegoBit(coefficient)

		// If current bit doesn't match desired bit, modify coefficient
		if currentBit != desiredBit {
			newCoeff := ModifyCoefficient(coefficient)

			// Handle shrinkage: if coefficient became zero, skip it and don't count the bit
			// We need to re-embed this bit in the next coefficient
			if newCoeff == 0 {
				coefficients[zigzag] = 0
				continue // Don't increment bitsEmbedded, try next coefficient
			}

			coefficients[zigzag] = newCoeff
		}

		bitsEmbedded++
	}

	if bitsEmbedded < HeaderSize {
		return 0, ValidationErrorf(ErrKeyInsufficientHeader, bitsEmbedded)
	}

	return coeffIndex, nil
}

// itoa is a simple integer to string conversion to avoid importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	// Maximum int64 has 19 digits, plus sign
	buf := make([]byte, 20)
	i := len(buf)

	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}

	if negative {
		i--
		buf[i] = '-'
	}

	return string(buf[i:])
}
