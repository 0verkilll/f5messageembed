package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/0verkilll/f5messageembed"
	"github.com/0verkilll/logger"
)

func main() {
	fmt.Println("F5 Custom Logger Example")
	fmt.Println("========================")
	fmt.Println()

	// Create synthetic coefficients
	coefficients := generateSyntheticCoefficients(10000)
	message := []byte("Hello, World!")
	password := "secret-password"

	// Example 1: Using the simple console logger
	fmt.Println("1. Console Logger (Simple Implementation)")
	fmt.Println("------------------------------------------")
	consoleLogger := NewConsoleLogger(logger.LevelDebug)

	result, err := f5messageembed.EmbedWithOptions(
		coefficients,
		password,
		message,
		f5messageembed.EmbedOptions{
			Logger: consoleLogger,
		},
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("\nEmbedding complete: %d bytes with k=%d\n", result.BytesEmbedded, result.KParameter)
	fmt.Println()

	// Example 2: Using a capturing logger for analysis
	fmt.Println("2. Capturing Logger (For Analysis)")
	fmt.Println("-----------------------------------")

	// Reset coefficients for second embedding
	coefficients = generateSyntheticCoefficients(10000)
	capturingLogger := NewCapturingLogger()

	result, err = f5messageembed.EmbedWithOptions(
		coefficients,
		password,
		message,
		f5messageembed.EmbedOptions{
			Logger: capturingLogger,
		},
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	entries := capturingLogger.Entries()
	fmt.Printf("Captured %d log entries:\n", len(entries))
	for i, entry := range entries {
		fmt.Printf("  [%d] %s: %s\n", i+1, entry.Level, entry.Message)
	}
	fmt.Println()

	// Example 3: Filtered logging (Info level only)
	fmt.Println("3. Filtered Logger (Info Level Only)")
	fmt.Println("------------------------------------")

	coefficients = generateSyntheticCoefficients(10000)
	infoLogger := NewConsoleLogger(logger.LevelInfo)

	_, err = f5messageembed.EmbedWithOptions(
		coefficients,
		password,
		message,
		f5messageembed.EmbedOptions{
			Logger: infoLogger,
		},
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println()

	// Example 4: No logging (default behavior)
	fmt.Println("4. No Logger (Default Behavior)")
	fmt.Println("--------------------------------")
	fmt.Println("When no logger is provided, embedding is silent.")

	coefficients = generateSyntheticCoefficients(10000)
	result, err = f5messageembed.Embed(coefficients, password, message)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Embedding complete: %d bytes (no log output above)\n", result.BytesEmbedded)
}

// LogEntry represents a captured log entry.
type LogEntry struct {
	Level   string
	Message string
	Fields  map[string]any
}

// ConsoleLogger is a simple logger that writes to stdout.
type ConsoleLogger struct {
	minLevel logger.Level
	fields   map[string]any
}

// NewConsoleLogger creates a console logger with the specified minimum level.
func NewConsoleLogger(minLevel logger.Level) *ConsoleLogger {
	return &ConsoleLogger{
		minLevel: minLevel,
		fields:   make(map[string]any),
	}
}

func (c *ConsoleLogger) log(level logger.Level, levelStr, msg string, args ...any) {
	if level < c.minLevel {
		return
	}

	timestamp := time.Now().Format("15:04:05.000")
	formattedMsg := msg
	if len(args) > 0 {
		formattedMsg = fmt.Sprintf(msg, args...)
	}

	// Build fields string
	var fieldsStr string
	if len(c.fields) > 0 {
		var parts []string
		for k, v := range c.fields {
			parts = append(parts, fmt.Sprintf("%s=%v", k, v))
		}
		fieldsStr = " " + strings.Join(parts, " ")
	}

	fmt.Printf("[%s] %s %s%s\n", timestamp, levelStr, formattedMsg, fieldsStr)
}

func (c *ConsoleLogger) Debug(msg string, args ...any) {
	c.log(logger.LevelDebug, "DEBUG", msg, args...)
}

func (c *ConsoleLogger) Info(msg string, args ...any) {
	c.log(logger.LevelInfo, "INFO ", msg, args...)
}

func (c *ConsoleLogger) Warn(msg string, args ...any) {
	c.log(logger.LevelWarn, "WARN ", msg, args...)
}

func (c *ConsoleLogger) Error(msg string, args ...any) {
	c.log(logger.LevelError, "ERROR", msg, args...)
}

func (c *ConsoleLogger) Fatal(msg string, args ...any) {
	c.log(logger.LevelFatal, "FATAL", msg, args...)
}

func (c *ConsoleLogger) WithFields(fields ...any) logger.Logger {
	newFields := make(map[string]any, len(c.fields)+len(fields)/2)
	for k, v := range c.fields {
		newFields[k] = v
	}
	for i := 0; i < len(fields)-1; i += 2 {
		if key, ok := fields[i].(string); ok {
			newFields[key] = fields[i+1]
		}
	}
	return &ConsoleLogger{
		minLevel: c.minLevel,
		fields:   newFields,
	}
}

func (c *ConsoleLogger) WithContext(_ context.Context) logger.Logger {
	return c
}

func (c *ConsoleLogger) WithLevel(level logger.Level) logger.Logger {
	return &ConsoleLogger{
		minLevel: level,
		fields:   c.fields,
	}
}

func (c *ConsoleLogger) Enabled(level logger.Level) bool {
	return level >= c.minLevel
}

// entryStore holds captured log entries with thread-safe access.
// It allows derived loggers from WithFields to share the same entry list.
type entryStore struct {
	mu      sync.Mutex
	entries []LogEntry
}

func (s *entryStore) append(e LogEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, e)
}

func (s *entryStore) getEntries() []LogEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]LogEntry, len(s.entries))
	copy(result, s.entries)
	return result
}

// CapturingLogger captures log entries for later analysis.
// It uses a shared entry store so WithFields-derived loggers
// write to the same entry list.
type CapturingLogger struct {
	store  *entryStore
	fields map[string]any
}

// NewCapturingLogger creates a logger that captures all entries.
func NewCapturingLogger() *CapturingLogger {
	return &CapturingLogger{
		store:  &entryStore{entries: make([]LogEntry, 0)},
		fields: make(map[string]any),
	}
}

func (c *CapturingLogger) capture(level, msg string, args ...any) {
	formattedMsg := msg
	if len(args) > 0 {
		formattedMsg = fmt.Sprintf(msg, args...)
	}

	// Copy fields for this entry
	fieldsCopy := make(map[string]any, len(c.fields))
	for k, v := range c.fields {
		fieldsCopy[k] = v
	}

	c.store.append(LogEntry{
		Level:   level,
		Message: formattedMsg,
		Fields:  fieldsCopy,
	})
}

func (c *CapturingLogger) Debug(msg string, args ...any) {
	c.capture("DEBUG", msg, args...)
}

func (c *CapturingLogger) Info(msg string, args ...any) {
	c.capture("INFO", msg, args...)
}

func (c *CapturingLogger) Warn(msg string, args ...any) {
	c.capture("WARN", msg, args...)
}

func (c *CapturingLogger) Error(msg string, args ...any) {
	c.capture("ERROR", msg, args...)
}

func (c *CapturingLogger) Fatal(msg string, args ...any) {
	c.capture("FATAL", msg, args...)
}

func (c *CapturingLogger) WithFields(fields ...any) logger.Logger {
	newFields := make(map[string]any, len(c.fields)+len(fields)/2)
	for k, v := range c.fields {
		newFields[k] = v
	}
	for i := 0; i < len(fields)-1; i += 2 {
		if key, ok := fields[i].(string); ok {
			newFields[key] = fields[i+1]
		}
	}
	return &CapturingLogger{
		store:  c.store, // Share the same store
		fields: newFields,
	}
}

func (c *CapturingLogger) WithContext(_ context.Context) logger.Logger {
	return c
}

func (c *CapturingLogger) WithLevel(_ logger.Level) logger.Logger {
	return c
}

func (c *CapturingLogger) Enabled(_ logger.Level) bool {
	return true
}

// Entries returns a copy of all captured log entries.
func (c *CapturingLogger) Entries() []LogEntry {
	return c.store.getEntries()
}

// generateSyntheticCoefficients creates coefficients with realistic JPEG distribution.
func generateSyntheticCoefficients(count int) []int16 {
	coefficients := make([]int16, count)

	for i := range coefficients {
		if i%64 == 0 {
			coefficients[i] = int16(100 + (i % 200))
			continue
		}

		switch {
		case i%7 == 0:
			coefficients[i] = 0
		case i%11 == 0:
			coefficients[i] = int16(10 + (i % 50))
		case i%13 == 0:
			coefficients[i] = int16(-(10 + (i % 50)))
		case i%5 == 0:
			coefficients[i] = int16(1 + (i % 3))
		case i%3 == 0:
			coefficients[i] = int16(-(1 + (i % 3)))
		default:
			if i%2 == 0 {
				coefficients[i] = int16(2 + (i % 10))
			} else {
				coefficients[i] = int16(-(2 + (i % 10)))
			}
		}
	}

	return coefficients
}
