package f5messageembed

// CalculateEmbeddingRate computes the embedding rate R(k) for a given k parameter.
//
// The embedding rate represents the efficiency of the matrix encoding:
// R(k) = k / (2^k - 1)
//
// As k increases, the embedding rate decreases but embedding efficiency improves
// (fewer changes per embedded bit). The optimal k balances capacity needs with
// statistical detectability.
//
// Embedding rates for common k values:
//   - k=1: R(1) = 1.000 (1 bit per coefficient, lowest efficiency)
//   - k=2: R(2) = 0.667 (2 bits per 3 coefficients)
//   - k=3: R(3) = 0.429 (3 bits per 7 coefficients)
//   - k=4: R(4) = 0.267 (4 bits per 15 coefficients)
//   - k=5: R(5) = 0.161 (5 bits per 31 coefficients)
//   - k=6: R(6) = 0.095 (6 bits per 63 coefficients)
//   - k=7: R(7) = 0.055 (7 bits per 127 coefficients)
//   - k=8: R(8) = 0.031 (8 bits per 255 coefficients, highest efficiency)
//
// Parameters:
//   - k: The matrix encoding parameter (typically 1-8)
//
// Returns:
//   - The embedding rate as a float64 value
//   - Returns 0 for invalid k values (k <= 0)
//
// Example:
//
//	rate := CalculateEmbeddingRate(4) // Returns 0.267 (4/15)
func CalculateEmbeddingRate(k int) float64 {
	// Handle invalid k values to prevent division by zero
	if k <= 0 {
		return 0
	}

	// n = 2^k - 1 (code word length)
	n := (1 << k) - 1

	// Guard against n being 0 (shouldn't happen for k > 0, but be safe)
	if n == 0 {
		return 0
	}

	// R(k) = k / n = k / (2^k - 1)
	return float64(k) / float64(n)
}

// SelectOptimalK determines the best k parameter for embedding a message.
//
// F5 uses (1, n, k) matrix encoding where n = 2^k - 1. Higher k values provide
// better embedding efficiency (fewer coefficient changes) but require more
// coefficients per embedded bit chunk. This function selects the largest k
// that can accommodate the given message plus the 32-bit header.
//
// The selection algorithm:
//  1. Try k from 8 down to 1 (highest efficiency first)
//  2. For each k, calculate capacity: (usableCoeffs / n) * k
//  3. Select the first k where capacity >= messageBits + HeaderSize
//  4. Return error if no k can accommodate the message
//
// The capacity formula accounts for:
//   - Code word length n = 2^k - 1
//   - Each code word embeds k bits
//   - The 32-bit header overhead (HeaderSize = 32)
//
// Parameters:
//   - usableCoeffCount: Number of non-zero, non-DC coefficients available
//   - messageBits: Number of bits in the message to embed
//
// Returns:
//   - k: The optimal encoding parameter (1-8)
//   - err: ErrInsufficientCapacity if message cannot fit
//
// Example:
//
//	// With 10000 usable coefficients and 1000 message bits:
//	k, err := SelectOptimalK(10000, 1000)
//	// Returns k=8 (highest efficiency that fits)
//
//	// With limited coefficients, a lower k may be selected:
//	k, err = SelectOptimalK(500, 400)
//	// Returns k=1 (only option with enough capacity)
func SelectOptimalK(usableCoeffCount, messageBits int) (k int, err error) {
	// Total bits needed = message bits + 32-bit header
	totalBitsNeeded := messageBits + HeaderSize

	// Try k from 7 down to 1 (prefer higher efficiency). Embedding is capped
	// at k=7; see the ForceK validation in embed.go for the rationale.
	for k = 7; k >= 1; k-- {
		// n = 2^k - 1 (code word length)
		n := (1 << k) - 1

		// Calculate capacity in bits: (usableCoeffs / n) * k
		// Using integer division: capacity = (usableCoeffs / n) * k
		// Note: We use integer division which truncates, giving a conservative estimate
		codeWords := usableCoeffCount / n
		capacityBits := codeWords * k

		// Check if this k has enough capacity
		if capacityBits >= totalBitsNeeded {
			return k, nil
		}
	}

	// No k value could accommodate the message
	return 0, ValidationErrorf(ErrKeyCapacityDetails, totalBitsNeeded, usableCoeffCount)
}

// CalculateCapacityForK computes the embedding capacity in bits for a specific k value.
//
// This helper function calculates how many message bits can be embedded
// using a specific k parameter, accounting for the header overhead.
//
// Parameters:
//   - usableCoeffCount: Number of non-zero, non-DC coefficients
//   - k: The matrix encoding parameter (1-8)
//
// Returns:
//   - The message capacity in bits (excluding header)
//
// Example:
//
//	// With 1000 coefficients and k=4 (n=15):
//	capacity := CalculateCapacityForK(1000, 4)
//	// Returns (1000/15)*4 - 32 = 66*4 - 32 = 232 bits
//
// NOTE: This function does not account for shrinkage. For accuracy that
// matches F5.jar's actual k-selection on real images, prefer
// CalculateCapacityForKWithShrinkage which uses f5.jar's `_expected`
// effective pool (P − h(1)) + ⌊0.49·h(1)⌋ before computing capacity.
func CalculateCapacityForK(usableCoeffCount, k int) int {
	if k <= 0 || k > 8 {
		return 0
	}

	n := (1 << k) - 1
	codeWords := usableCoeffCount / n
	totalCapacity := codeWords * k

	// Subtract header overhead
	messageCapacity := totalCapacity - HeaderSize
	if messageCapacity < 0 {
		return 0
	}

	return messageCapacity
}

// EffectiveCapacityCoefficients returns the F5 effective number of
// coefficients available for matrix encoding (f5.jar's `_expected`), which
// reserves space for the expected shrinkage cascade.
//
// It reproduces f5.jar's shipped calculation BYTE-EXACTLY
// (james/JpegEncoder.java:281-282):
//
//	_large    = coeffCount - _zero - _one - coeffCount/64;  // = P - h(1)
//	_expected = _large + (int) (0.49 * _one);               // INTEGER TRUNCATION
//
// where P = usableCoeffCount (non-zero, non-DC) and _one = h(1) =
// magnitudeOneCount. Re-expressed against P:
//
//	_expected = (P − h(1)) + ⌊0.49·h(1)⌋
//
// This is NOT the same as P − round(0.51·h(1)). For even h(1) the two
// differ by exactly 1 — the algebraic identity
// (P−h1)+⌊0.49·h1⌋ = P − ⌈0.51·h1⌉ holds, and ⌈0.51·h1⌉ exceeds
// round(0.51·h1) by 1 whenever h1 is even. A 1-coefficient error in
// `_expected` flips the selected k at a code-word bucket boundary
// (selectOptimalKJavaExact's `usable < byteToEmbed+4` test), producing
// byte-divergent, un-extractable stego. The earlier round(0.51·h1) form
// matched f5.jar only on the ~half of covers with odd h(1); this exact
// truncation matches on all of them.
//
// Go's int() conversion truncates toward zero for the non-negative product
// 0.49·h(1), matching Java's (int) cast. Source: Westfeld 2001 §3,
// Liu 2020 Eq.(1), and f5.jar james/JpegEncoder.java.
//
// Returns 0 if usableCoeffCount is non-positive or the reserve consumes the
// entire pool.
func EffectiveCapacityCoefficients(usableCoeffCount, magnitudeOneCount int) int {
	if usableCoeffCount <= 0 {
		return 0
	}
	if magnitudeOneCount < 0 {
		magnitudeOneCount = 0
	}
	// f5.jar JpegEncoder.java:281-282, byte-exact. int() truncates toward
	// zero just like Java's (int) cast on the non-negative product.
	large := usableCoeffCount - magnitudeOneCount
	effective := large + int(0.49*float64(magnitudeOneCount))
	if effective < 0 {
		return 0
	}
	return effective
}

// CalculateCapacityForKWithShrinkage is the shrinkage-aware variant of
// CalculateCapacityForK. It computes the message-byte capacity using the
// f5.jar effective-capacity `_expected` = (P − h(1)) + ⌊0.49·h(1)⌋, which
// reserves space for the expected shrinkage cascade.
//
// Parameters mirror CalculateCapacityForK plus magnitudeOneCount, the count
// of usable coefficients with |value| = 1. CapacityResult.EstimatedShrinkageFactor
// times CapacityResult.UsableCoefficients gives the same value.
//
// Use this for embed-time k-selection when capacity fidelity to F5.jar's
// published formula matters more than the simpler/faster non-shrinkage form.
func CalculateCapacityForKWithShrinkage(usableCoeffCount, magnitudeOneCount, k int) int {
	if k <= 0 || k > 8 {
		return 0
	}
	effective := EffectiveCapacityCoefficients(usableCoeffCount, magnitudeOneCount)
	return CalculateCapacityForK(effective, k)
}

// selectOptimalKJavaExact replicates f5.jar's exact k-selection arithmetic
// from JpegEncoder.java:337-349:
//
//	for (i = 1; i < 8; i++) {
//	    n = (1 << i) - 1;
//	    usable = _expected * i / n - _expected * i / n % n;
//	    usable /= 8;
//	    if (usable == 0) break;
//	    if (usable < byteToEmbed + 4) break;
//	}
//	k = i - 1;
//
// This produces k values byte-identical to f5.jar's choice, including the
// bucket-boundary cases where the more-standard floor(P/n)*k formula would
// pick a different k. The quirky `usable - usable%n` clamps the bit count
// to a multiple of n before converting to bytes — at large n (e.g. n=127
// for i=7) this can return 0 bits where the standard formula would compute
// non-zero capacity, forcing k = i - 1.
//
// effectiveCoeffs = (P − h(1)) + ⌊0.49·h(1)⌋, f5.jar's `_expected` (see
// EffectiveCapacityCoefficients). messageBytes = message length in bytes (Java's
// byteToEmbed). The Java loop's check `usable < byteToEmbed + 4` is the
// 4-byte (32-bit) F5 header overhead.
//
// Returns k in [1, 7]; returns 0 if no k fits.
func selectOptimalKJavaExact(effectiveCoeffs, messageBytes int) int {
	if effectiveCoeffs <= 0 || messageBytes < 0 {
		return 0
	}
	i := 1
	for ; i < 8; i++ {
		n := (1 << i) - 1
		q := effectiveCoeffs * i / n
		usableBits := q - (q % n)
		usableBytes := usableBits / 8
		if usableBytes == 0 {
			break
		}
		if usableBytes < messageBytes+4 {
			break
		}
	}
	return i - 1
}

// SelectOptimalKWithShrinkage is the shrinkage-aware k-selector. It
// computes the f5.jar effective capacity (P − h(1)) + ⌊0.49·h(1)⌋ and then
// walks the same iterative k-selection loop f5.jar uses (JpegEncoder.java:337-
// 349, see selectOptimalKJavaExact), so it returns the SAME k value
// f5.jar would pick for the same (cover, message) inputs.
//
// Prior to 2026-05-23 this delegated to SelectOptimalK (a "largest k that
// fits" walk using the standard floor(P/n)*k capacity formula). That gave
// the WRONG k at certain bucket-boundary covers — empirically a 64×64
// cover with 1081 usable coefficients and 1-byte message: f5.jar picks k=6
// (its quirky formula returns 0 bits at i=7, forcing break and k = i-1);
// the old SelectOptimalK picked k=7 because the standard formula gives 56
// non-zero bits. The mismatched k produced byte-divergent stego on roughly
// half of (cover, Q) combinations — see project_f5jar_parity.md and the
// reproducer at /tmp/mcu_parity_test/.
//
// Use this in place of SelectOptimalK whenever the caller has access to
// h(1) (the magnitudeOneCount). The non-shrinkage SelectOptimalK is
// retained for API compatibility but does NOT match f5.jar's k choice.
func SelectOptimalKWithShrinkage(usableCoeffCount, magnitudeOneCount, messageBits int) (int, error) {
	effective := EffectiveCapacityCoefficients(usableCoeffCount, magnitudeOneCount)
	// Java's byteToEmbed is in bytes. messageBits is byte-aligned at every
	// caller (embed.go:170 sets `messageBits := len(message) * 8`), so the
	// /8 is exact for our use; rounding up is the safe behaviour for any
	// non-aligned future caller.
	messageBytes := (messageBits + 7) / 8
	k := selectOptimalKJavaExact(effective, messageBytes)
	if k <= 0 {
		totalBitsNeeded := messageBits + HeaderSize
		return 0, ValidationErrorf(ErrKeyCapacityDetails, totalBitsNeeded, usableCoeffCount)
	}
	return k, nil
}
