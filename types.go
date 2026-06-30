package f5messageembed

import (
	"github.com/0verkilll/f5core"
	"github.com/0verkilll/logger"
)

// Re-export constants from f5core for backward compatibility.
const (
	// MaxMessageSize is the maximum message size in bytes that can be embedded.
	// See f5core.MaxMessageSize for documentation.
	MaxMessageSize = f5core.MaxMessageSize

	// CoefficientMin is the minimum valid JPEG quantized DCT coefficient value.
	// See f5core.CoefficientMin for documentation.
	CoefficientMin = f5core.CoefficientMin

	// CoefficientMax is the maximum valid JPEG quantized DCT coefficient value.
	// See f5core.CoefficientMax for documentation.
	CoefficientMax = f5core.CoefficientMax

	// HeaderSize is the number of bits in the F5 message header.
	// See f5core.HeaderSize for documentation.
	HeaderSize = f5core.HeaderSize
)

// deZigZag is an alias to f5core.DeZigZag for package-internal use.
// See f5core.DeZigZag for full documentation.
var deZigZag = f5core.DeZigZag[:]

// EmbedResult contains the results of an F5 embedding operation.
//
// The Coefficients field contains the modified coefficient slice. Note that
// embedding modifies coefficients in-place, so this is the same slice reference
// passed to the Embed function.
type EmbedResult struct {
	// Coefficients is the modified coefficient slice after embedding.
	// This is the same slice reference passed to Embed (modified in-place).
	Coefficients []int16

	// KParameter is the matrix encoding parameter used (1-8).
	// Higher k values provide better efficiency but require more coefficients.
	KParameter int

	// BytesEmbedded is the number of message bytes successfully embedded.
	BytesEmbedded int

	// ShrinkageCount is the number of times shrinkage occurred during embedding.
	// Shrinkage happens when decrementing |1| or |-1| produces 0, requiring
	// the bit to be re-embedded using the next available coefficient.
	ShrinkageCount int

	// UsableCoefficients is the count of non-zero, non-DC coefficients used.
	UsableCoefficients int
}

// CapacityResult contains the capacity analysis results for a set of coefficients.
//
// Use this to determine if a message can be embedded before attempting
// the embedding operation.
type CapacityResult struct {
	// CapacityByK maps each k parameter (1-8) to the maximum message capacity
	// in bytes when using that k value. Higher k values generally have lower
	// capacity but better embedding efficiency.
	CapacityByK map[int]int

	// EstimatedShrinkageFactor is the estimated proportion of coefficients
	// that will cause shrinkage (coefficients with absolute value 1).
	// Range: 0.0 to 1.0
	EstimatedShrinkageFactor float64

	// TotalCoefficients is the total number of coefficients in the input.
	TotalCoefficients int

	// UsableCoefficients is the count of non-zero, non-DC coefficients.
	// Only usable coefficients can carry steganographic data.
	UsableCoefficients int

	// MagnitudeOneCount is the exact integer count of usable coefficients with
	// absolute value 1 (i.e. h(1), the shrinkage-eligible pool). This is the
	// raw integer counter the capacity scan computes; the value is exposed so
	// callers can use it without round-tripping
	// EstimatedShrinkageFactor × UsableCoefficients through float64 (which
	// can lose ±1 from the integer count). EstimatedShrinkageFactor is kept
	// for API compatibility but is now derived from this field.
	MagnitudeOneCount int
}

// EmbedOptions provides optional configuration for the embedding operation.
//
// All fields are optional. Zero values indicate the default behavior:
//   - Logger: nil (no logging)
//   - ForceK: 0 (auto-select optimal k)
type EmbedOptions struct {
	// Logger is an optional logger for debugging and monitoring.
	// When nil, no logging is performed.
	Logger logger.Logger

	// ForceK forces a specific k parameter (1-8) instead of auto-selection.
	// When 0, the optimal k is automatically selected based on message size
	// and available capacity.
	ForceK int
}

// EmbedOption is a functional option for configuring embedding behavior.
// Use these options with EmbedWithOptions or apply them directly to EmbedOptions.
type EmbedOption func(*EmbedOptions)

// WithLogger returns an EmbedOption that sets the logger for debugging and monitoring.
//
// When a logger is provided, the embedding process will log:
//   - Debug level: k selection, shrinkage events, coefficient changes, capacity analysis
//   - Info level: embedding start, embedding complete
//
// When nil is passed, no logging is performed (same as default behavior).
//
// Example:
//
//	opts := EmbedOptions{}
//	WithLogger(myLogger)(&opts)
//	result, err := EmbedWithOptions(coeffs, password, message, opts)
func WithLogger(log logger.Logger) EmbedOption {
	return func(opts *EmbedOptions) {
		opts.Logger = log
	}
}

// WithForceK returns an EmbedOption that forces a specific k parameter.
//
// The k parameter (1-8) controls the trade-off between embedding efficiency
// and capacity:
//   - Higher k: Better efficiency (fewer changes per bit), but less capacity
//   - Lower k: More capacity, but more changes required
//
// When k is 0 (default), the optimal k is automatically selected based on
// message size and available coefficient capacity.
//
// Example:
//
//	opts := EmbedOptions{}
//	WithForceK(4)(&opts)
//	result, err := EmbedWithOptions(coeffs, password, message, opts)
func WithForceK(k int) EmbedOption {
	return func(opts *EmbedOptions) {
		opts.ForceK = k
	}
}
