# Capacity Calculation Example

Demonstrates how to analyze JPEG DCT coefficients to determine embedding capacity.

## Quick Start

```go
import "github.com/0verkilll/f5messageembed"

// Analyze coefficient array capacity
coefficients := []int16{/* ... */}
result := f5messageembed.CalculateCapacity(coefficients)

fmt.Printf("Total: %d, Usable: %d\n", result.TotalCoefficients, result.UsableCoefficients)
fmt.Printf("Capacity at k=4: %d bytes\n", result.CapacityByK[4])
fmt.Printf("Expected shrinkage: %.1f%%\n", result.EstimatedShrinkageFactor*100)
```

## Running

```bash
go run main.go
```

## Output

```
F5 Capacity Calculation Example
================================

1. Small Coefficient Array (1,000 coefficients)
------------------------------------------------
Total coefficients:      1000
Usable coefficients:     841 (84.1%)
Est. shrinkage factor:   33.5%

Capacity by k parameter:
  k | n (code word) | Capacity (bytes)
  --|---------------|------------------
  1 |             1 |              101
  2 |             3 |               66
  3 |             7 |               44
  ...
```

## What This Demonstrates

- Calculating embedding capacity before embedding
- Understanding the trade-off between k values
- Analyzing coefficient distributions
- Validating message sizes against capacity

## Understanding CapacityResult

| Field | Description |
|-------|-------------|
| `TotalCoefficients` | Total number of coefficients |
| `UsableCoefficients` | Non-zero, non-DC coefficients (can carry data) |
| `CapacityByK` | Max message bytes for each k value (1-8) |
| `EstimatedShrinkageFactor` | Proportion of |1| coefficients (cause shrinkage) |

## K Parameter Trade-offs

| k | Code Word Size (n) | Efficiency | Capacity |
|---|-------------------|------------|----------|
| 1 | 1 | 1.000 | Highest |
| 2 | 3 | 0.667 | ... |
| 4 | 15 | 0.267 | ... |
| 8 | 255 | 0.031 | Lowest |

- **Higher k**: Fewer changes per message bit (harder to detect), but lower capacity
- **Lower k**: More capacity, but more coefficient modifications required

## Usable vs Total Coefficients

Not all coefficients can carry steganographic data:
- **DC coefficients** (index % 64 == 0): Skipped to preserve image quality
- **Zero coefficients**: Already zero, can't encode meaningful bits

The usable coefficient count determines actual embedding capacity.

## Shrinkage Factor

The shrinkage factor estimates how many coefficients have magnitude 1.
When these are modified, they become 0 (shrinkage), requiring re-embedding
with additional coefficients.

High shrinkage (>40%) may reduce effective capacity below theoretical limits.
