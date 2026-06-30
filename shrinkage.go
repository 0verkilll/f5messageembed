package f5messageembed

import (
	"github.com/0verkilll/f5coefficient"
)

// DetectShrinkage determines if modifying a coefficient will cause shrinkage.
// This is an alias to f5coefficient.IsShrinkageCandidate for backward compatibility.
// See f5coefficient.IsShrinkageCandidate for full documentation.
func DetectShrinkage(coefficient int16) bool {
	return f5coefficient.IsShrinkageCandidate(coefficient)
}

// HandleShrinkage applies a coefficient modification and detects shrinkage.
//
// This function combines the modification operation with shrinkage detection,
// providing a single entry point for the shrinkage handling workflow in the
// F5 embedding process.
//
// The function:
//  1. Applies ModifyCoefficient to decrement the absolute value
//  2. Detects if the original coefficient was 1 or -1 (shrinkage case)
//  3. Returns both the shrinkage status and the new coefficient value
//
// The codeWord and position parameters are included for interface consistency
// with the matrix encoding workflow, where shrinkage handling may need context
// about which code word and position triggered the shrinkage. In the current
// implementation, these parameters are reserved for future use.
//
// Parameters:
//   - codeWord: The current code word being processed (reserved for future use)
//   - position: The position within the code word (reserved for future use)
//   - coefficient: The coefficient to modify
//
// Returns:
//   - shrunk: true if the modification caused the coefficient to become 0
//   - newCoeff: the coefficient value after modification
//
// Example:
//
//	shrunk, newCoeff := HandleShrinkage(codeWord, 1, int16(5))
//	// shrunk = false, newCoeff = 4
//
//	shrunk, newCoeff = HandleShrinkage(codeWord, 1, int16(1))
//	// shrunk = true, newCoeff = 0
//
//	shrunk, newCoeff = HandleShrinkage(codeWord, 1, int16(-1))
//	// shrunk = true, newCoeff = 0
func HandleShrinkage(codeWord []int, position int, coefficient int16) (shrunk bool, newCoeff int16) {
	// Detect shrinkage before modification
	// (coefficient of 1 or -1 will become 0)
	shrunk = f5coefficient.IsShrinkageCandidate(coefficient)

	// Apply the modification (decrement absolute value)
	newCoeff = f5coefficient.ModifyCoefficient(coefficient)

	// Unused parameters are reserved for future matrix encoding integration
	_ = codeWord
	_ = position

	return shrunk, newCoeff
}
