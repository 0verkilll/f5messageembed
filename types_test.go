package f5messageembed

import (
	"testing"
)

// TestEmbedResultStructFieldAccess tests that EmbedResult struct fields
// are accessible and store values correctly.
func TestEmbedResultStructFieldAccess(t *testing.T) {
	t.Parallel()

	coefficients := []int16{1, 2, 3, 4, 5}
	result := &EmbedResult{
		Coefficients:       coefficients,
		KParameter:         4,
		BytesEmbedded:      100,
		ShrinkageCount:     5,
		UsableCoefficients: 1000,
	}

	// Verify all fields are accessible and store correct values
	if len(result.Coefficients) != 5 {
		t.Errorf("Coefficients length = %d, want 5", len(result.Coefficients))
	}
	if result.Coefficients[0] != 1 {
		t.Errorf("Coefficients[0] = %d, want 1", result.Coefficients[0])
	}
	if result.KParameter != 4 {
		t.Errorf("KParameter = %d, want 4", result.KParameter)
	}
	if result.BytesEmbedded != 100 {
		t.Errorf("BytesEmbedded = %d, want 100", result.BytesEmbedded)
	}
	if result.ShrinkageCount != 5 {
		t.Errorf("ShrinkageCount = %d, want 5", result.ShrinkageCount)
	}
	if result.UsableCoefficients != 1000 {
		t.Errorf("UsableCoefficients = %d, want 1000", result.UsableCoefficients)
	}
}

// TestCapacityResultStructFieldAccess tests that CapacityResult struct fields
// are accessible and store values correctly.
func TestCapacityResultStructFieldAccess(t *testing.T) {
	t.Parallel()

	capacityByK := map[int]int{
		1: 1000,
		2: 800,
		3: 600,
		4: 500,
		5: 400,
		6: 300,
		7: 200,
		8: 100,
	}

	result := &CapacityResult{
		TotalCoefficients:        10000,
		UsableCoefficients:       8000,
		CapacityByK:              capacityByK,
		EstimatedShrinkageFactor: 0.05,
	}

	// Verify all fields are accessible and store correct values
	if result.TotalCoefficients != 10000 {
		t.Errorf("TotalCoefficients = %d, want 10000", result.TotalCoefficients)
	}
	if result.UsableCoefficients != 8000 {
		t.Errorf("UsableCoefficients = %d, want 8000", result.UsableCoefficients)
	}
	if len(result.CapacityByK) != 8 {
		t.Errorf("CapacityByK length = %d, want 8", len(result.CapacityByK))
	}
	if result.CapacityByK[4] != 500 {
		t.Errorf("CapacityByK[4] = %d, want 500", result.CapacityByK[4])
	}
	if result.EstimatedShrinkageFactor != 0.05 {
		t.Errorf("EstimatedShrinkageFactor = %f, want 0.05", result.EstimatedShrinkageFactor)
	}
}

// TestEmbedOptionsValidation tests EmbedOptions struct field access
// and verifies the optional fields work correctly.
func TestEmbedOptionsValidation(t *testing.T) {
	t.Parallel()

	// Test with empty options (default values)
	opts := EmbedOptions{}

	if opts.Logger != nil {
		t.Error("Default Logger should be nil")
	}
	if opts.ForceK != 0 {
		t.Errorf("Default ForceK = %d, want 0", opts.ForceK)
	}

	// Test with ForceK set
	optsWithK := EmbedOptions{
		ForceK: 4,
	}

	if optsWithK.ForceK != 4 {
		t.Errorf("ForceK = %d, want 4", optsWithK.ForceK)
	}
}

// TestConstantValues verifies package constants have correct values.
// These tests document the expected values and will fail if they change.
func TestConstantValues(t *testing.T) {
	t.Parallel()

	// Use a helper to avoid "always false" static analysis warnings
	// while still documenting and verifying expected constant values
	assertEqual := func(name string, got, want int) {
		if got != want {
			t.Errorf("%s = %d, want %d", name, got, want)
		}
	}

	// MaxMessageSize should be 2^23 - 1 = 8,388,607
	assertEqual("MaxMessageSize", MaxMessageSize, (1<<23)-1)

	// CoefficientMin should be -2048
	assertEqual("CoefficientMin", int(CoefficientMin), -2048)

	// CoefficientMax should be 2047
	assertEqual("CoefficientMax", int(CoefficientMax), 2047)

	// HeaderSize should be 32 bits
	assertEqual("HeaderSize", HeaderSize, 32)
}

// TestDeZigZagTableLength verifies the deZigZag table has exactly 64 entries.
func TestDeZigZagTableLength(t *testing.T) {
	t.Parallel()

	if len(deZigZag) != 64 {
		t.Errorf("deZigZag table length = %d, want 64", len(deZigZag))
	}
}

// TestDeZigZagTableMatchesF5PixelExtraction verifies that the deZigZag table
// matches the exact values used in the f5pixel extraction implementation.
func TestDeZigZagTableMatchesF5PixelExtraction(t *testing.T) {
	t.Parallel()

	// Expected values from f5pixel/internal/core/extraction/f5.go
	expected := []int{
		0, 1, 5, 6, 14, 15, 27, 28,
		2, 4, 7, 13, 16, 26, 29, 42,
		3, 8, 12, 17, 25, 30, 41, 43,
		9, 11, 18, 24, 31, 40, 44, 53,
		10, 19, 23, 32, 39, 45, 52, 54,
		20, 22, 33, 38, 46, 51, 55, 60,
		21, 34, 37, 47, 50, 56, 59, 61,
		35, 36, 48, 49, 57, 58, 62, 63,
	}

	if len(deZigZag) != len(expected) {
		t.Fatalf("deZigZag length = %d, expected length = %d", len(deZigZag), len(expected))
	}

	for i := range expected {
		if deZigZag[i] != expected[i] {
			t.Errorf("deZigZag[%d] = %d, want %d", i, deZigZag[i], expected[i])
		}
	}
}
