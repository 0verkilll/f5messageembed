package f5messageembed

import (
	"github.com/0verkilll/f5prng"
	"github.com/0verkilll/logger"
)

// Embed embeds a secret message into JPEG DCT coefficients using the F5 algorithm.
//
// The F5 algorithm (Westfeld, 2001) provides high embedding efficiency through
// matrix encoding and distributes changes uniformly via permutative straddling.
// This implementation produces byte-identical output to the Java reference
// implementation when using the same password and coefficients.
//
// The embedding process:
//  1. Validate inputs (coefficients, password, message)
//  2. Initialize PRNG with password and generate permutation
//  3. Calculate capacity and auto-select optimal k parameter
//  4. Build and XOR the 32-bit header
//  5. Embed header using simple LSB embedding
//  6. Embed message using matrix encoding with shrinkage handling
//
// Coefficients are modified in-place for memory efficiency. The returned
// EmbedResult contains the same slice reference along with embedding metadata.
//
// Parameters:
//   - coefficients: JPEG DCT coefficients to modify (modified in-place)
//   - password: Password for PRNG seeding (must not be empty)
//   - message: Message bytes to embed (max 8,388,607 bytes)
//
// Returns:
//   - *EmbedResult: Embedding results including metadata
//   - error: Validation or embedding errors
//
// Example:
//
//	result, err := Embed(coefficients, "secret password", []byte("hidden message"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Embedded %d bytes with k=%d\n", result.BytesEmbedded, result.KParameter)
func Embed(coefficients []int16, password string, message []byte) (*EmbedResult, error) {
	return EmbedWithOptions(coefficients, password, message, EmbedOptions{})
}

// EmbedWithOptions embeds a secret message with additional configuration options.
//
// This function provides the same functionality as Embed but allows:
//   - Logger: Inject a logger for debugging and monitoring
//   - ForceK: Override automatic k selection with a specific value (1-8)
//
// When ForceK is 0 (default), the optimal k is automatically selected based on
// message size and available coefficient capacity, preferring higher k values
// for better embedding efficiency.
//
// Internally, this function creates a PRNG using InitializePRNG(password) and
// delegates to EmbedWithRandomSource for the actual embedding. The PRNG is
// securely cleared after embedding completes.
//
// Parameters:
//   - coefficients: JPEG DCT coefficients to modify (modified in-place)
//   - password: Password for PRNG seeding (must not be empty)
//   - message: Message bytes to embed (max 8,388,607 bytes)
//   - opts: Configuration options (Logger, ForceK)
//
// Returns:
//   - *EmbedResult: Embedding results including metadata
//   - error: Validation or embedding errors
//
// Example:
//
//	opts := EmbedOptions{
//	    Logger: myLogger,
//	    ForceK: 4, // Use k=4 instead of auto-selection
//	}
//	result, err := EmbedWithOptions(coefficients, password, message, opts)
func EmbedWithOptions(coefficients []int16, password string, message []byte, opts EmbedOptions) (*EmbedResult, error) {
	// Validate password before creating PRNG
	if err := ValidatePassword(password); err != nil {
		return nil, err
	}

	// Initialize PRNG with password
	prng := InitializePRNG(password)
	defer prng.Clear()

	// Delegate to EmbedWithRandomSource for actual embedding
	return EmbedWithRandomSource(coefficients, prng, message, opts)
}

// EmbedWithRandomSource embeds a secret message using a pre-seeded RandomSource.
//
// This function provides the core embedding functionality and is designed for
// callers who want to manage their own PRNG lifecycle, such as when implementing
// dependency injection patterns or when the same PRNG needs to be used across
// multiple operations.
//
// The RandomSource must be pre-seeded before calling this function. The function
// consumes PRNG state in the following order, matching the Java F5 reference
// implementation exactly:
//  1. GeneratePermutation (consumes PRNG state proportional to coefficient count)
//  2. XORHeaderWithPRNG (consumes 4 PRNG bytes)
//  3. Message XOR during embedding (consumes 1 PRNG byte per message byte)
//
// Unlike EmbedWithOptions, this function does NOT call prng.Clear() after
// embedding. The caller is responsible for clearing the PRNG when done if
// security-sensitive state needs to be wiped.
//
// Parameters:
//   - coefficients: JPEG DCT coefficients to modify (modified in-place)
//   - random: A pre-seeded RandomSource (must not be nil)
//   - message: Message bytes to embed (max 8,388,607 bytes)
//   - opts: Configuration options (Logger, ForceK)
//
// Returns:
//   - *EmbedResult: Embedding results including metadata
//   - error: Validation or embedding errors
//
// Example:
//
//	prng := InitializePRNG(password)
//	defer prng.Clear()
//	result, err := EmbedWithRandomSource(coefficients, prng, message, EmbedOptions{})
//
// PRNG Consumption Order (Java F5 compatible):
//
//	The RandomSource state is consumed in this exact order:
//	1. GeneratePermutation: Uses NextInt() for Fisher-Yates shuffle
//	2. Header XOR: Uses NextBytes(4) for header obfuscation
//	3. Message XOR: Uses NextBytes(1) per message byte during embedding
func EmbedWithRandomSource(coefficients []int16, random RandomSource, message []byte, opts EmbedOptions) (*EmbedResult, error) {
	log := opts.Logger

	// Step 1: Validate RandomSource (must not be nil)
	if random == nil {
		return nil, newValidationError(ErrKeyNilRandomSource)
	}

	// Step 2: Validate ForceK if specified.
	// Embedding is capped at k=7: k=8 (255 coeffs / 8 bits) buys negligible
	// efficiency over k=7 while sharply raising the per-codeword coefficient
	// demand, so the encoder never produces it. (The extractor still accepts
	// k up to 31 to stay compatible with payloads from other F5 tools.)
	if opts.ForceK != 0 && (opts.ForceK < 1 || opts.ForceK > 7) {
		return nil, newValidationError(ErrKeyInvalidForceK)
	}

	// Step 3: Validate coefficients
	if err := ValidateCoefficients(coefficients); err != nil {
		return nil, err
	}

	// Step 4: Calculate capacity to validate message size
	capacityResult := CalculateCapacity(coefficients)
	usableCount := capacityResult.UsableCoefficients
	// h(1) = exact integer count of |c|=1 coefficients = shrinkage-eligible
	// pool. Read directly from CapacityResult rather than reconstructing the
	// count from EstimatedShrinkageFactor (which round-trips the integer
	// through float64 and could lose ±1).
	magnitudeOneCount := capacityResult.MagnitudeOneCount

	logDebug(log, "Capacity analysis complete",
		"total", capacityResult.TotalCoefficients,
		"usable", usableCount,
		"magnitudeOne", magnitudeOneCount,
		"shrinkageFactor", capacityResult.EstimatedShrinkageFactor,
		"effectiveCapacity", EffectiveCapacityCoefficients(usableCount, magnitudeOneCount))

	// Step 5: Select k parameter using f5.jar's effective-capacity
	// `_expected` = (P − h(1)) + ⌊0.49·h(1)⌋. This reserves space for the expected
	// shrinkage cascade (when |c|=1 coefficients are decremented to 0,
	// the same code word must be re-embedded in the next block, consuming
	// extra coefficients that the naive P/n accounting ignores). Without
	// this reservation, the selector picks an over-large k for certain
	// message sizes and embedding fails partway through with
	// "insufficient coefficients during embedding".
	messageBits := len(message) * 8
	var k int
	var kSelectErr error

	if opts.ForceK != 0 {
		k = opts.ForceK
		// Validate that forced k has enough EFFECTIVE capacity. Falling
		// back to non-shrinkage capacity here would re-introduce the
		// silent mid-embed failure mode that ForceK callers were trying
		// to avoid.
		capacityBits := CalculateCapacityForKWithShrinkage(usableCount, magnitudeOneCount, k)
		if messageBits > capacityBits {
			return nil, newValidationError(ErrKeyInsufficientCapacity)
		}
		logDebug(log, "Using forced k parameter", "k", k)
	} else {
		k, kSelectErr = SelectOptimalKWithShrinkage(usableCount, magnitudeOneCount, messageBits)
		if kSelectErr != nil {
			// Convert to validation error for consistency
			return nil, newValidationError(ErrKeyInsufficientCapacity)
		}
		logDebug(log, "Auto-selected k parameter", "k", k, "messageBits", messageBits)
	}

	// Step 6: Validate message against EFFECTIVE capacity for selected k.
	// We compute this from the shrinkage-aware formula rather than reading
	// capacityResult.CapacityByK[k] (which is non-shrinkage and would
	// over-state capacity for the same reason SelectOptimalK does).
	capacityBytes := CalculateCapacityForKWithShrinkage(usableCount, magnitudeOneCount, k) / 8
	if msgErr := ValidateMessage(message, capacityBytes); msgErr != nil {
		return nil, msgErr
	}

	logInfo(log, "Starting F5 embedding",
		"messageSize", len(message),
		"k", k,
		"usableCoefficients", usableCount)

	// Step 7: Generate permutation (consumes PRNG state)
	permutation, permErr := GeneratePermutation(random, len(coefficients))
	if permErr != nil {
		return nil, ValidationErrorf(ErrKeyPermutationFailed, permErr)
	}

	// Step 8: Build header (k parameter + message size)
	header := BuildHeader(k, len(message))

	// Step 9: XOR header with PRNG bytes (consumes 4 PRNG bytes)
	xoredHeader := XORHeaderWithPRNG(header, random)

	// Step 10: Embed header using simple LSB embedding
	nextIndex, headerErr := EmbedHeader(coefficients, permutation, xoredHeader)
	if headerErr != nil {
		return nil, ValidationErrorf(ErrKeyHeaderEmbedFailed, headerErr)
	}

	logDebug(log, "Header embedded", "nextIndex", nextIndex)

	// Step 11: Embed message using matrix encoding
	shrinkageCount, embedErr := embedMessageWithMatrix(
		coefficients,
		permutation,
		nextIndex,
		message,
		k,
		random,
		log,
	)
	if embedErr != nil {
		return nil, ValidationErrorf(ErrKeyMessageEmbedFailed, embedErr)
	}

	logInfo(log, "F5 embedding complete",
		"bytesEmbedded", len(message),
		"shrinkageCount", shrinkageCount)

	return &EmbedResult{
		Coefficients:       coefficients,
		KParameter:         k,
		BytesEmbedded:      len(message),
		ShrinkageCount:     shrinkageCount,
		UsableCoefficients: usableCount,
	}, nil
}

// embeddingState tracks the state during message embedding.
type embeddingState struct {
	// Message data after XOR with PRNG
	xoredMessage []byte
	// Current bit position in the message
	bitPos int
	// Total number of bits to embed
	totalBits int
}

// newEmbeddingState creates a new embedding state by XORing message with PRNG.
//
// The XOR stream is allocated in bulk (one make + one PRNG call) rather than
// one NextBytes(1) call per message byte. When the injected PRNG satisfies
// f5prng.RandomSourceWithBytesInto, the stream is written directly into the
// caller buffer — zero allocation on the hot path. Otherwise we fall back to a
// single NextBytes(len(message)) call. Either path produces the exact same
// byte stream as the old per-byte loop (PRNG state advances identically), so
// Java F5 parity is preserved.
func newEmbeddingState(message []byte, prng RandomSource) *embeddingState {
	xoredMessage := make([]byte, len(message))
	if len(message) == 0 {
		return &embeddingState{
			xoredMessage: xoredMessage,
			bitPos:       0,
			totalBits:    0,
		}
	}

	// Bulk-fill the XOR stream. Prefer the zero-alloc NextBytesInto API when
	// the concrete PRNG supports it; otherwise fall back to NextBytes(n).
	var xorStream []byte
	if into, ok := prng.(f5prng.RandomSourceWithBytesInto); ok {
		// Reuse xoredMessage's backing array as scratch for the XOR stream:
		// we overwrite each byte below during the XOR loop, so borrowing the
		// buffer here saves a second allocation. After NextBytesInto runs,
		// xoredMessage holds the raw PRNG bytes; we then XOR them with
		// message in place.
		xorStream = xoredMessage
		// Error intentionally discarded: any PRNG failure is retained on the
		// source itself (LastError) and surfaces on the next draw. This path
		// also mirrors the NextBytes(n) fallback which cannot report errors.
		_ = into.NextBytesInto(xorStream) //nolint:errcheck // see comment above
	} else {
		// Fallback: single NextBytes call (still one allocation, not len(message)).
		xorStream = prng.NextBytes(len(message))
	}

	// XOR each message byte with PRNG byte using Java-compatible signed
	// arithmetic (int8 sign extension matches the Java reference). The two
	// gosec G115 warnings here are intentional: Java bytes are signed, so the
	// int8(uint8) reinterpretation is by design, and the final byte(int^int)
	// wraps back into the byte domain with identical bit pattern to Java.
	for i := range message {
		xorByte := int8(xorStream[i])                          //nolint:gosec // G115: Java-compat signed byte
		xoredMessage[i] = byte(int(message[i]) ^ int(xorByte)) //nolint:gosec // G115: XOR result wraps back to byte
	}

	return &embeddingState{
		xoredMessage: xoredMessage,
		bitPos:       0,
		totalBits:    len(message) * 8,
	}
}

// getNextKBits extracts the next k bits from the XORed message and advances
// the internal bit cursor by k. Callers should check s.bitPos / s.totalBits
// themselves to decide whether to keep iterating — the older signature also
// returned a hasMore bool but no caller consulted it (the embed loop tracks
// bits via bitsEmbedded vs totalBits and the trailing zero-bit chunk flag),
// so the second return was dropped to avoid foot-gunning future refactors.
func (s *embeddingState) getNextKBits(k int) int {
	if s.bitPos >= s.totalBits {
		return 0
	}

	bits := 0
	for i := 0; i < k && s.bitPos+i < s.totalBits; i++ {
		byteIdx := (s.bitPos + i) / 8
		bitInByte := (s.bitPos + i) % 8
		bit := (int(s.xoredMessage[byteIdx]) >> bitInByte) & 1
		bits |= bit << i
	}

	s.bitPos += k
	return bits
}

// embedMessageWithMatrix embeds the message bytes using matrix encoding.
//
// This function implements the core F5 matrix encoding loop:
//  1. XOR message bytes with PRNG stream
//  2. For each k-bit chunk, collect n coefficients (n = 2^k - 1)
//  3. Compute hash of current code word
//  4. Apply matrix encoding to determine which coefficient to modify
//  5. Handle shrinkage by re-embedding when |1| or |-1| becomes 0
//
// The PRNG consumption order matches the Java F5 reference implementation:
// - One byte per message byte for XOR operation (consumed before embedding)
//
// Parameters:
//   - coefficients: The coefficient array to modify
//   - permutation: The Fisher-Yates permutation of indices
//   - startIndex: The permutation index to start from (after header)
//   - message: The message bytes to embed
//   - k: The matrix encoding parameter (1-8)
//   - prng: The PRNG for message XOR
//   - log: Optional logger for debugging
//
// Returns:
//   - shrinkageCount: Total number of shrinkage events
//   - error: Any embedding errors
func embedMessageWithMatrix(
	coefficients []int16,
	permutation []int,
	startIndex int,
	message []byte,
	k int,
	prng RandomSource,
	log logger.Logger,
) (shrinkageCount int, err error) {
	// Calculate code word length: n = 2^k - 1
	n := CodeWordLength(k)

	// k=1 (n=1) takes Java's "default code" branch (JpegEncoder.java
	// lines 497-547) — plain LSB embed that simply walks coefficients
	// in permutation order, embedding one bit per non-zero coefficient.
	// It does NOT do the trailing zero-bit matrix chunk that the n>1
	// do-while produces. With an empty message there are no message
	// bits to embed and the loop exits without touching any coefficient.
	if k <= 1 && len(message) == 0 {
		return 0, nil
	}

	// Create embedding state with XORed message
	state := newEmbeddingState(message, prng)

	// Current position in permutation
	permIndex := startIndex

	// Total bits to embed
	totalBits := len(message) * 8
	bitsEmbedded := 0

	// Hoist per-codeword scratch buffers out of the inner loop. The matrix
	// encoding step collects exactly n stego-bits plus n coefficient indices
	// per k-bit chunk; allocating both slices inside the loop grew the
	// embedding alloc count linearly with the number of codewords. We reuse
	// the same backing arrays by resetting length to 0 each iteration.
	codeWord := make([]int, 0, n)
	indices := make([]int, 0, n)

	// Java parity: the reference embed loop is a do-while that always runs
	// at least one matrix-encode iteration, and only exits AFTER an iteration
	// that consumed zero new bits (i.e. the bit-source was already empty when
	// the k-bit refill started). With the previous `for bitsEmbedded < totalBits`
	// loop we exited one iteration too early whenever 8*len(message) was a
	// multiple of k (including len(message)==0), producing one fewer coefficient
	// change than f5.jar. See /tmp/f5jar_extract/james/JpegEncoder.java
	// lines 421-496 (the `embeddingLoop: do { ... } while (!isLastByte)` block):
	// kBitsToEmbed is initialised to 0 each iteration, the refill `for` may
	// break early with isLastByte=true and kBitsToEmbed still 0, and the
	// matrix-encode step always executes before the while-condition is
	// re-evaluated.
	// k=1 follows Java's default-code path and exits when bits run out
	// (no trailing zero-bit chunk). k>=2 follows the matrix-encode do-while
	// and always runs one more iteration after bits are exhausted.
	hasTrailingChunk := k >= 2
	isLastChunk := false
	for {
		// Get the next k bits to embed. When the bit-source is exhausted at
		// the START of this chunk, Java (n>1 path) embeds messageBits=0 once
		// more and then exits. When the bit-source runs dry MID-chunk, Java
		// embeds the partial bits accumulated so far (LSBs only, upper bits
		// zero) and also exits. The bits-tracking below mirrors both cases.
		var messageBits int
		if bitsEmbedded >= totalBits {
			if !hasTrailingChunk {
				return shrinkageCount, nil
			}
			// Java path: availableBitsToEmbed==0 at i==0, isLastByte=true,
			// kBitsToEmbed=0.
			messageBits = 0
			isLastChunk = true
		} else {
			messageBits = state.getNextKBits(k)
			bitsRemaining := totalBits - bitsEmbedded
			if bitsRemaining >= k {
				bitsEmbedded += k
				// If we just consumed the LAST k bits exactly, Java will
				// still do one more iteration with messageBits=0; defer
				// isLastChunk to the next pass.
			} else {
				// Partial-fill: Java sets isLastByte=true mid-refill and
				// runs matrix-encode with the partial bits. No further
				// iteration after this one.
				bitsEmbedded += bitsRemaining
				isLastChunk = true
			}
		}

		// Java-faithful shrinkage cascade (JpegEncoder.java:410-488). For each
		// symbol Java runs an inner `do { j = startOfN; collect n usable coeffs
		// from j; hash; apply change } while (changed coeff == 0)`: on shrinkage
		// it RESTARTS codeword collection from startOfN — the just-zeroed
		// coefficient is then skipped (coeff==0) so the window naturally extends
		// by one — and recomputes the hash from scratch. The previous Go strategy
		// (shift-the-buffer + append-one) produced a different change position
		// than f5.jar in the tail-of-permutation shrinkage cases (the ≤6-coeff
		// drift). startOfN advances only after the symbol is embedded.
		startOfN := permIndex
		for {
			permIndex = startOfN
			codeWord = codeWord[:0]
			indices = indices[:0]

			for len(codeWord) < n && permIndex < len(permutation) {
				shuffled := permutation[permIndex]
				permIndex++

				// Skip DC coefficients. BlockSize is 64 (power of two, compile-
				// time asserted in f5coefficient), so `shuffled & 63` matches
				// `shuffled % 64` and avoids IDIV in the hot loop.
				if shuffled&63 == 0 {
					continue
				}
				zigzag := ApplyDeZigZag(shuffled)
				if zigzag >= len(coefficients) {
					continue
				}
				coeff := coefficients[zigzag]
				if coeff == 0 {
					continue
				}
				codeWord = append(codeWord, GetStegoBit(coeff))
				indices = append(indices, zigzag)
			}

			// Capacity exhausted: Java `break label323` (silent stop). On the
			// trailing zero-bit chunk this is benign; otherwise it is an error.
			if len(codeWord) < n {
				if isLastChunk && messageBits == 0 {
					return shrinkageCount, nil
				}
				return shrinkageCount, ValidationErrorf(ErrKeyInsufficientCoeffs, n, len(codeWord))
			}

			changePos, encErr := MatrixEncode(codeWord, messageBits, k)
			if encErr != nil {
				return shrinkageCount, ValidationErrorf(ErrKeyMatrixEncodeFailed, encErr)
			}
			if changePos == 0 {
				// hash already matches the message symbol — no change needed.
				break
			}

			coeffIdx := indices[changePos-1]
			shrunk, newCoeff := HandleShrinkage(codeWord, changePos, coefficients[coeffIdx])
			coefficients[coeffIdx] = newCoeff
			if !shrunk {
				// Embedded this symbol without shrinkage; permIndex is where the
				// collection ended → next symbol's startOfN.
				break
			}
			// Shrinkage: the changed |1| coefficient became 0. Java loops with
			// j = startOfN and re-collects, which now skips the zeroed coeff.
			shrinkageCount++
		}

		if isLastChunk {
			return shrinkageCount, nil
		}
	}
}

// logDebug logs a debug message if logger is available.
// Uses the logger.Logger interface with WithFields for structured logging.
// When logger is nil, no operation is performed (zero overhead).
func logDebug(log logger.Logger, msg string, keyvals ...any) {
	if log != nil {
		log.WithFields(keyvals...).Debug(msg)
	}
}

// logInfo logs an info message if logger is available.
// Uses the logger.Logger interface with WithFields for structured logging.
// When logger is nil, no operation is performed (zero overhead).
func logInfo(log logger.Logger, msg string, keyvals ...any) {
	if log != nil {
		log.WithFields(keyvals...).Info(msg)
	}
}
