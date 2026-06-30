# Internationalization (i18n) Example

Demonstrates how to use localized error messages with the f5messageembed package.

## Quick Start

```go
import (
    "github.com/0verkilll/f5messageembed"
    "github.com/0verkilll/i18n"
)

// Create a translator using f5messageembed's locales
translator, err := i18n.New(
    i18n.WithFileSystemLoader("path/to/f5messageembed/locales"),
    i18n.WithDefaultLocale("en-US"),
)
if err != nil {
    log.Fatal(err)
}

// Set the translator for f5messageembed
f5messageembed.SetTranslator(translator)

// Now all error messages will use the translator
result, err := f5messageembed.Embed(coefficients, password, message)
if err != nil {
    fmt.Println(err) // Translated error message
}
```

## Running

```bash
go run main.go
```

## Output

```
F5 Internationalization (i18n) Example
=======================================

1. Default Behavior (No Translator)
------------------------------------
Without a translator, f5messageembed uses default English messages:

  Empty coefficients error: coefficient slice cannot be empty
  Empty password error:     password cannot be empty

2. Using i18n Translator
------------------------
Translator set for f5messageembed package.
  Translated error: coefficient slice cannot be empty
  Current locale: en-US

3. Validation Error Messages
-----------------------------
All validation error messages:

  1. coefficient slice cannot be empty
  2. password cannot be empty
  3. coefficient value out of valid range (-2048 to 2047)
  4. message too large for available coefficient capacity
  5. invalid ForceK: must be 0 (auto) or 1-8: got 9
```

## What This Demonstrates

- Default English fallback when no translator is set
- Setting up i18n translator with f5messageembed locales
- All validation error messages
- Direct access to translation keys
- Runtime translator management

## Error Keys

The f5messageembed package defines the following error keys for translation:

| Key | Default Message |
|-----|-----------------|
| `f5messageembed.error.empty_coefficients` | coefficient slice cannot be empty |
| `f5messageembed.error.empty_password` | password cannot be empty |
| `f5messageembed.error.message_too_large` | message size exceeds maximum allowed (8,388,607 bytes) |
| `f5messageembed.error.insufficient_capacity` | message too large for available coefficient capacity |
| `f5messageembed.error.invalid_coefficient_range` | coefficient value out of valid range (-2048 to 2047) |
| `f5messageembed.error.invalid_k_parameter` | k parameter must be between 1 and 8 |

## Locales Directory Structure

The f5messageembed package includes translation files in `locales/`:

```
locales/
  en-us.json    # English (US)
```

## Adding New Languages

To add support for additional languages:

1. Copy `locales/en-us.json` to a new file (e.g., `locales/es-es.json`)
2. Translate all values in the new file
3. Load the translator with the new locale:

```go
translator, _ := i18n.New(
    i18n.WithFileSystemLoader("locales"),
    i18n.WithDefaultLocale("es-ES"),
)
f5messageembed.SetTranslator(translator)
```

## TranslatorProvider Interface

The `f5messageembed.SetTranslator()` function accepts any type implementing `TranslatorProvider`:

```go
type TranslatorProvider interface {
    Translate(key string) string
    TranslateWithArgs(key string, args ...interface{}) string
    HasKey(key string) bool
    SetLocale(locale string)
    GetLocale() string
}
```

This allows integration with custom translation systems beyond the i18n package.
