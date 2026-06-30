package f5messageembed

import (
	"github.com/0verkilll/f5coefficient"
)

// GetStegoBit extracts the steganographic bit value from a DCT coefficient.
// This is an alias to f5coefficient.GetStegoBit for backward compatibility.
// See f5coefficient.GetStegoBit for full documentation.
func GetStegoBit(coefficient int16) int {
	return f5coefficient.GetStegoBit(coefficient)
}

// ModifyCoefficient decrements the absolute value of a coefficient.
// This is an alias to f5coefficient.ModifyCoefficient for backward compatibility.
// See f5coefficient.ModifyCoefficient for full documentation.
func ModifyCoefficient(coefficient int16) int16 {
	return f5coefficient.ModifyCoefficient(coefficient)
}

// IsUsableCoefficient determines if a coefficient can carry steganographic data.
// This is an alias to f5coefficient.IsUsableCoefficient for backward compatibility.
// See f5coefficient.IsUsableCoefficient for full documentation.
func IsUsableCoefficient(index int, coefficient int16) bool {
	return f5coefficient.IsUsableCoefficient(index, coefficient)
}
