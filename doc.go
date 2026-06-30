// Package f5messageembed implements the F5 steganographic algorithm for embedding
// hidden messages into JPEG DCT coefficients.
//
// The F5 algorithm was developed by Andreas Westfeld (2001) and is designed
// to resist both visual and statistical detection attacks while maintaining
// high embedding capacity. This implementation produces byte-identical output
// to the Java reference implementation for interoperability.
//
// # Algorithm Overview
//
// F5 builds upon earlier steganographic techniques:
//   - Jsteg: LSB replacement (detectable via chi-square attack)
//   - F3: Decrements instead of overwrites (detectable via even coefficient surplus)
//   - F4: Sign-based encoding (resistant to statistical attacks)
//   - F5: F4 + matrix encoding + permutative straddling (optimal efficiency)
//
// # Key Features
//
//   - Matrix Encoding: Uses (1, n, k) codes where n = 2^k - 1. Embeds k message
//     bits in n coefficients with at most 1 change per code word.
//   - Permutative Straddling: Shuffles coefficient indices before embedding to
//     distribute changes uniformly and prevent localized statistical analysis.
//   - Shrinkage Handling: When decrementing |1| or |-1| produces 0, the bit is
//     re-embedded using the next available coefficient.
//   - Optimal K Selection: Automatically selects the best k parameter (1-8) based
//     on message length and usable coefficient capacity.
//
// # Security Warning
//
// This package uses SHA1PRNG for pseudorandom number generation to maintain
// compatibility with the Java reference implementation. SHA1PRNG is NOT
// cryptographically secure by modern standards. The algorithm relies on SHA-1
// which has known collision vulnerabilities.
//
// DO NOT use this package for security-critical applications requiring
// cryptographic randomness. The PRNG is deterministic and predictable given
// the password/seed.
//
// This package is intended for:
//   - Compatibility with existing F5-encoded images (e.g., PixelKnot)
//   - Research and educational purposes
//   - Applications where statistical undetectability is prioritized over
//     cryptographic security
//
// # Basic Usage
//
// Embedding a message into JPEG DCT coefficients:
//
//	import "github.com/0verkilll/f5messageembed"
//
//	// Coefficients from JPEG DCT (obtained via jpeg package)
//	coefficients := []int16{/* ... */}
//	password := "secret"
//	message := []byte("Hello, World!")
//
//	result, err := f5messageembed.Embed(coefficients, password, message)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Embedded %d bytes using k=%d\n", result.BytesEmbedded, result.KParameter)
//
// # Capacity Calculation
//
// Before embedding, you can calculate the maximum message capacity:
//
//	capacity := f5messageembed.CalculateCapacity(coefficients)
//	fmt.Printf("Usable coefficients: %d\n", capacity.UsableCoefficients)
//	fmt.Printf("Max capacity at k=4: %d bytes\n", capacity.CapacityByK[4])
//
// # Advanced Options
//
// Using options for logging and custom k parameter:
//
//	result, err := f5messageembed.EmbedWithOptions(
//	    coefficients,
//	    password,
//	    message,
//	    f5messageembed.EmbedOptions{
//	        Logger: myLogger,
//	        ForceK: 4, // Force k=4 instead of auto-selection
//	    },
//	)
//
// # Research Functions
//
// The package exports low-level functions for research purposes:
//
//	// Get steganographic bit value from coefficient
//	bit := f5messageembed.GetStegoBit(coefficient)
//
//	// Modify coefficient to flip its steganographic bit
//	newCoeff := f5messageembed.ModifyCoefficient(coefficient)
//
//	// Matrix encoding for a code word
//	changePos, err := f5messageembed.MatrixEncode(codeWord, messageBits, k)
//
//	// Select optimal k parameter
//	k, err := f5messageembed.SelectOptimalK(usableCoeffCount, messageBits)
//
// # Dependencies
//
// This package depends on the following sibling packages:
//   - f5coefficient: DCT coefficient bit operations (GetStegoBit, ModifyCoefficient, IsShrinkageCandidate)
//   - f5core: Shared constants (MaxMessageSize, HeaderSize, CoefficientMin/Max, DeZigZag table)
//   - f5matrix: Matrix encoding for (1,n,k) code words
//   - f5prng: Unified PRNG interfaces and SHA1PRNG implementation for deterministic randomness
//   - fisheryates: Fisher-Yates shuffle for coefficient permutation
//   - logger: Optional structured logging support
//
// Indirect dependencies (via f5prng):
//   - sha1: SHA-1 hasher for PRNG seeding
//   - i18n: Internationalized error messages
//
// # Reference
//
// Westfeld, A. (2001). F5 - A Steganographic Algorithm: High Capacity Despite
// Better Steganalysis. Lecture Notes in Computer Science, 2137, 289-302.
package f5messageembed
