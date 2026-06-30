package f5messageembed

import (
	"github.com/0verkilll/f5coefficient"
	"github.com/0verkilll/f5matrix"
)

// Re-export matrix encoding errors from f5matrix for backward compatibility.
// These are part of the public API and intentionally exported for callers.
var (
	ErrEmptyCodeWord        = f5matrix.ErrEmptyCodeWord
	ErrInvalidKParameter    = f5matrix.ErrInvalidKParameter
	ErrCodeWordSizeMismatch = f5matrix.ErrCodeWordSizeMismatch
	ErrMessageBitsExceedK   = f5matrix.ErrMessageBitsExceedK
)

// Ensure exported errors are recognized as intentionally public API.
var (
	_ = ErrEmptyCodeWord
	_ = ErrInvalidKParameter
	_ = ErrCodeWordSizeMismatch
	_ = ErrMessageBitsExceedK
)

// ComputeCodeWordHash computes the hash function for matrix encoding.
// This is an alias to f5matrix.ComputeCodeWordHash for backward compatibility.
// See f5matrix.ComputeCodeWordHash for full documentation.
func ComputeCodeWordHash(codeWord []int, n int) int {
	// Note: local signature included n parameter for historical reasons
	// f5matrix version doesn't need n since it uses len(codeWord)
	_ = n
	return f5matrix.ComputeCodeWordHash(codeWord)
}

// MatrixEncode performs (1, n, k) matrix encoding to determine which coefficient
// to modify in order to embed k message bits.
// This is an alias to f5matrix.MatrixEncode for backward compatibility.
// See f5matrix.MatrixEncode for full documentation.
func MatrixEncode(codeWord []int, messageBits, k int) (changePosition int, err error) {
	return f5matrix.MatrixEncode(codeWord, messageBits, k)
}

// ApplyMatrixChange applies the coefficient modification at the calculated position.
// This is an alias to f5matrix.ApplyMatrixChange for backward compatibility.
// See f5matrix.ApplyMatrixChange for full documentation.
func ApplyMatrixChange(coefficients []int16, indices []int, position int) int16 {
	return f5matrix.ApplyMatrixChange(coefficients, indices, position)
}

// CodeWordLength returns the code word length n for a given k parameter.
// This is an alias to f5matrix.CodeWordLength for backward compatibility.
// See f5matrix.CodeWordLength for full documentation.
func CodeWordLength(k int) int {
	return f5matrix.CodeWordLength(k)
}

// ExtractStegoBits extracts the steganographic bit values from a slice of coefficients.
// This is an alias to f5coefficient.ExtractStegoBits for backward compatibility.
// See f5coefficient.ExtractStegoBits for full documentation.
func ExtractStegoBits(coefficients []int16) []int {
	return f5coefficient.ExtractStegoBits(coefficients)
}
