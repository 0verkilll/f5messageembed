# Examples

Runnable examples demonstrating f5messageembed package features.

## Running Examples

```bash
cd examples/<example>
go run main.go
```

## Examples Overview

| Example | Description                                                       |
|---------|-------------------------------------------------------------------|
| [basic/](basic/) | Simple F5 embedding: embedding a message into JPEG coefficients   |
| [capacity/](capacity/) | Capacity calculation: analyzing coefficient arrays before embedding |
| [custom-logger/](custom-logger/) | Logger injection: using a custom logger for debugging             |
| [i18n/](i18n/) | Internationalization: localized error messages                    |

## Prerequisites

All examples require simulated JPEG DCT coefficients. In real usage, you would obtain
these from a JPEG decoder. These examples use synthetic coefficients for demonstration.

## Security Notice

The F5 algorithm uses SHA1PRNG for pseudorandom number generation to maintain
compatibility with the Java reference implementation. SHA1PRNG is NOT
cryptographically secure by modern standards.

This package is intended for:
- Compatibility with existing F5-encoded images (e.g., PixelKnot)
- Research and educational purposes
- Applications where statistical undetectability is prioritized over cryptographic security
