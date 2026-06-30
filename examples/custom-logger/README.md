# Custom Logger Example

Demonstrates how to inject a custom logger for debugging and monitoring F5 embedding.

## Quick Start

```go
import (
    "github.com/0verkilll/f5messageembed"
    "github.com/0verkilll/logger"
)

// Use embedding with a custom logger
result, err := f5messageembed.EmbedWithOptions(
    coefficients,
    password,
    message,
    f5messageembed.EmbedOptions{
        Logger: myLogger,  // Any logger.Logger implementation
    },
)
```

## Running

```bash
go run main.go
```

## Output

```
F5 Custom Logger Example
========================

1. Console Logger (Simple Implementation)
------------------------------------------
[14:30:00.123] DEBUG Capacity analysis complete total=10000 usable=8413 shrinkageFactor=0.33
[14:30:00.124] DEBUG Auto-selected k parameter k=6 messageBits=104
[14:30:00.125] INFO  Starting F5 embedding messageSize=13 k=6 usableCoefficients=8413
[14:30:00.126] INFO  F5 embedding complete bytesEmbedded=13 shrinkageCount=2

Embedding complete: 13 bytes with k=6

2. Capturing Logger (For Analysis)
-----------------------------------
Captured 4 log entries:
  [1] DEBUG: Capacity analysis complete
  [2] DEBUG: Auto-selected k parameter
  [3] INFO: Starting F5 embedding
  [4] INFO: F5 embedding complete
```

## What This Demonstrates

- Injecting a custom logger via EmbedOptions
- Creating a simple console logger implementation
- Creating a capturing logger for analysis
- Filtering log output by level
- Using the default silent logging (no logger)

## Log Levels

The f5messageembed package logs at two levels:

| Level | Messages |
|-------|----------|
| DEBUG | Capacity analysis, k selection, shrinkage events, coefficient changes |
| INFO  | Embedding start, embedding complete |

## Implementing logger.Logger

To create a custom logger, implement the `logger.Logger` interface:

```go
type Logger interface {
    Debug(msg string, args ...any)
    Info(msg string, args ...any)
    Warn(msg string, args ...any)
    Error(msg string, args ...any)
    Fatal(msg string, args ...any)
    WithFields(fields ...any) Logger
    WithContext(ctx context.Context) Logger
    WithLevel(level Level) Logger
    Enabled(level Level) bool
}
```

## Using the Mock Logger for Testing

For testing, use the `logger/testing` package:

```go
import loggertesting "github.com/0verkilll/logger/testing"

mockLog := loggertesting.NewMockLogger()

result, err := f5messageembed.EmbedWithOptions(
    coefficients,
    password,
    message,
    f5messageembed.EmbedOptions{Logger: mockLog},
)

// Verify log entries
entries := mockLog.Entries()
if len(entries) < 4 {
    t.Error("expected at least 4 log entries")
}
```

## Integrating with Popular Loggers

The logger interface is compatible with adapters for popular logging libraries:

- **Zap**: Wrap `*zap.SugaredLogger`
- **Logrus**: Wrap `*logrus.Logger`
- **Zerolog**: Wrap `zerolog.Logger`

See the logger package documentation for adapter examples.
