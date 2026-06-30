package f5messageembed

import (
	"github.com/0verkilll/f5prng"
	"github.com/0verkilll/fisheryates"
)

// RandomSource is a type alias for f5prng.RandomSource to maintain backward compatibility.
// New code should use f5prng.RandomSource directly.
//
// Deprecated: Use f5prng.RandomSource directly. This type alias will be removed in a future version.
type RandomSource = f5prng.RandomSource

// InitializePRNG creates and initializes a SecureRandom instance for F5 embedding.
//
// Deprecated: Use EmbedWithRandomSource with a pre-seeded RandomSource instead.
// This function is kept for backward compatibility but new code should prefer
// the dependency injection pattern using EmbedWithRandomSource which accepts
// a pre-seeded RandomSource, allowing better testability and flexibility.
//
// The PRNG is seeded with the password bytes using SHA-1 hashing, matching
// Java's SecureRandom SHA1PRNG algorithm exactly. This ensures byte-identical
// output when using the same password as the Java F5 reference implementation.
//
// The returned RandomSource should be cleared when no longer needed to prevent
// sensitive password-derived state from remaining in memory:
//
//	prng := InitializePRNG(password)
//	defer prng.Clear()
//	// ... use prng
//
// Parameters:
//   - password: The password string used to seed the PRNG
//
// Returns:
//   - A seeded f5prng.RandomSource ready for permutation generation
//
// Migration Example:
//
// Before (deprecated):
//
//	result, err := EmbedWithOptions(coefficients, password, message, opts)
//
// After (recommended):
//
//	prng := InitializePRNG(password)  // or your own RandomSource implementation
//	defer prng.Clear()
//	result, err := EmbedWithRandomSource(coefficients, prng, message, opts)
func InitializePRNG(password string) f5prng.RandomSource {
	// Create the default factory for f5prng
	factory := f5prng.NewDefaultFactory()

	// Create a new PRNG instance
	random := factory.NewPRNG()

	// Seed with the RAW password bytes. Seed can only fail with ErrNilHasher / a
	// hasher error, and the factory just produced a PRNG with a live SHA-1
	// hasher, so this path is unreachable. Any real failure will surface via
	// the PRNG's LastError on the first draw.
	//
	// The real Westfeld f5.jar constructs its PRNG directly from the password:
	//
	//	this.random = new SecureRandom(password.getBytes());
	//
	// Java's SHA1PRNG engineSetSeed then derives state = SHA-1(password).
	// f5prng's SecureRandom.Seed maps 1:1 onto SHA1PRNG.setSeed (it applies that
	// same internal SHA-1), so passing the raw password bytes yields
	// state = SHA-1(password) — byte-parity with f5.jar.
	//
	// Do NOT pre-hash here (sha1.Sum(password)): that yields
	// state = SHA-1(SHA-1(password)), a double hash that matches neither real
	// f5.jar nor images it produced, and derives the wrong permutation.
	//
	// Note: []byte(password) encodes the string as UTF-8. The Java reference
	// uses password.getBytes() (platform default, typically UTF-8). For
	// ASCII-only passwords these are identical; characters above U+007F may
	// differ. ASCII passwords are strongly recommended for interoperability.
	_ = random.Seed([]byte(password)) //nolint:errcheck // see comment above

	return random
}

// GeneratePermutation generates a Fisher-Yates permutation of indices.
//
// The permutation is used for "permutative straddling" in F5, which shuffles
// coefficient indices before embedding to distribute changes uniformly across
// the image. This prevents localized statistical analysis attacks.
//
// The function uses the fisheryates package's GenerateInto method for zero
// allocation when reusing buffers. The permutation is deterministic based on
// the RandomSource state, ensuring the same password produces the same
// permutation for both embedding and extraction.
//
// PRNG Consumption: This function consumes PRNG state proportional to the
// size parameter. After calling GeneratePermutation, the PRNG state will
// have advanced exactly as the Java F5 implementation expects.
//
// Parameters:
//   - random: A seeded f5prng.RandomSource (typically from InitializePRNG)
//   - size: The number of elements in the permutation (coefficient count)
//
// Returns:
//   - A slice containing a permutation of [0, 1, ..., size-1]
//   - An error if size is negative or exceeds fisheryates.MaxPermutationSize
//
// Example:
//
//	prng := InitializePRNG("password")
//	defer prng.Clear()
//	perm, err := GeneratePermutation(prng, len(coefficients))
//	if err != nil {
//	    return err
//	}
//	// perm now contains shuffled indices for coefficient access
func GeneratePermutation(random f5prng.RandomSource, size int) ([]int, error) {
	// Create Fisher-Yates permutator
	permutator := fisheryates.NewFisherYates()

	// Generate permutation using zero-allocation method
	// We pass nil buffer since we don't have a reusable buffer
	// fisheryates.RandomSource interface is compatible with f5prng.RandomSource
	return permutator.GenerateInto(nil, size, random)
}

// ApplyDeZigZag transforms a shuffled coefficient index using the JPEG de-zigzag table.
//
// JPEG stores DCT coefficients in zigzag order to improve compression by
// grouping zero coefficients together. The F5 algorithm applies this
// transformation after permutation to access coefficients in their natural
// 8x8 block positions.
//
// Formula: zigzag = shuffled - shuffled%64 + deZigZag[shuffled%64]
//
// This can be understood as:
//   - (shuffled / 64) * 64: The base offset of the 8x8 block
//   - deZigZag[shuffled % 64]: The de-zigzag position within the block
//
// Parameters:
//   - shuffledIndex: The permuted coefficient index
//
// Returns:
//   - The de-zigzagged index for accessing the coefficient array
//
// Example:
//
//	shuffled := permutation[i]
//	zigzag := ApplyDeZigZag(shuffled)
//	coefficient := coefficients[zigzag]
func ApplyDeZigZag(shuffledIndex int) int {
	// BlockSize is 64 (power of two, compile-time asserted in f5coefficient),
	// so `&63` replaces `%64` and `&^63` replaces `x - x%64` on non-negative
	// indices — both eliminate IDIV in a loop that runs once per usable
	// coefficient.
	posInBlock := shuffledIndex & 63
	blockBase := shuffledIndex &^ 63
	return blockBase + deZigZag[posInBlock]
}
