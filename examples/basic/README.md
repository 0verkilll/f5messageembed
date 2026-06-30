# Basic Example

Simple F5 steganographic embedding demonstration.

## Quick Start

```go
import "github.com/0verkilll/f5messageembed"

// JPEG DCT coefficients (from a JPEG decoder)
coefficients := []int16{/* ... */}

// Embed a secret message
result, err := f5messageembed.Embed(coefficients, "password", []byte("secret"))
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Embedded %d bytes with k=%d\n", result.BytesEmbedded, result.KParameter)
```

## Running

```bash
go run main.go
```

## Output

```
F5 Steganographic Embedding - Basic Example
=============================================

Generated 10000 synthetic coefficients
Message to embed: "Hello, World! This is a secret message."
Message size: 39 bytes
Password: secret-password

Embedding Results:
------------------
  Bytes embedded:      39
  K parameter used:    6
  Usable coefficients: 8413
  Shrinkage count:     5

Coefficient Analysis:
---------------------
  Total coefficients:    10000
  Modified coefficients: 42
  Modification rate:     0.42%

Success! The message has been embedded into the coefficients.
```

## What This Demonstrates

- Creating F5 embedder with default settings
- Embedding a message into JPEG DCT coefficients
- Understanding the EmbedResult structure
- Coefficient modification efficiency

## Key Concepts

### EmbedResult Fields

| Field | Description |
|-------|-------------|
| `Coefficients` | The modified coefficient slice (same reference as input) |
| `KParameter` | The matrix encoding parameter used (1-8) |
| `BytesEmbedded` | Number of message bytes successfully embedded |
| `ShrinkageCount` | Number of shrinkage events during embedding |
| `UsableCoefficients` | Count of non-zero, non-DC coefficients |

### K Parameter

The k parameter controls the trade-off between embedding efficiency and capacity:
- Higher k = fewer coefficient changes per message bit (better stealth)
- Lower k = more capacity but more changes required

The `Embed` function automatically selects the optimal k based on message size
and available coefficient capacity.
