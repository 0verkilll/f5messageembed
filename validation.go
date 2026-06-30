package f5messageembed

import (
	"fmt"
	"sync/atomic"
)

// Error keys for i18n translation.
// These keys correspond to entries in locales/en-US.json.
// #nosec G101 -- These are translation keys, not hardcoded credentials
const (
	ErrKeyEmptyCoefficients     = "f5messageembed.error.empty_coefficients" //nolint:gosec // G101: translation key, not credential
	ErrKeyEmptyPassword         = "f5messageembed.error.empty_password"
	ErrKeyMessageTooLarge       = "f5messageembed.error.message_too_large"
	ErrKeyInsufficientCapacity  = "f5messageembed.error.insufficient_capacity"
	ErrKeyInvalidCoefficientVal = "f5messageembed.error.invalid_coefficient_range"
	ErrKeyInvalidKParameter     = "f5messageembed.error.invalid_k_parameter"
	ErrKeyInvalidForceK         = "f5messageembed.error.invalid_force_k"
	ErrKeyPermutationFailed     = "f5messageembed.error.permutation_failed"
	ErrKeyHeaderEmbedFailed     = "f5messageembed.error.header_embed_failed"
	ErrKeyMessageEmbedFailed    = "f5messageembed.error.message_embed_failed"
	ErrKeyMatrixEncodeFailed    = "f5messageembed.error.matrix_encode_failed"
	ErrKeyInsufficientCoeffs    = "f5messageembed.error.insufficient_coefficients"
	ErrKeyInsufficientHeader    = "f5messageembed.error.insufficient_header_coefficients"
	ErrKeyCapacityDetails       = "f5messageembed.error.capacity_details"
	ErrKeyNilRandomSource       = "f5messageembed.error.nil_random_source"
)

// Default English error messages for fallback when no translator is set.
var defaultMessages = map[string]string{
	ErrKeyEmptyCoefficients:     "coefficient slice cannot be empty",
	ErrKeyEmptyPassword:         "password cannot be empty",
	ErrKeyMessageTooLarge:       "message size exceeds maximum allowed (8,388,607 bytes)",
	ErrKeyInsufficientCapacity:  "message too large for available coefficient capacity",
	ErrKeyInvalidCoefficientVal: "coefficient value out of valid range (-2048 to 2047)",
	ErrKeyInvalidKParameter:     "k parameter must be between 1 and 8",
	ErrKeyInvalidForceK:         "invalid ForceK: must be 0 (auto) or 1-8",
	ErrKeyPermutationFailed:     "failed to generate coefficient permutation: %v",
	ErrKeyHeaderEmbedFailed:     "failed to embed message header: %v",
	ErrKeyMessageEmbedFailed:    "failed to embed message: %v",
	ErrKeyMatrixEncodeFailed:    "matrix encoding failed for code word: %v",
	ErrKeyInsufficientCoeffs:    "insufficient coefficients during embedding: need %d, got %d",
	ErrKeyInsufficientHeader:    "insufficient usable coefficients for header embedding: need 32 bits, found only %d usable coefficients",
	ErrKeyCapacityDetails:       "need %d bits, maximum capacity at k=1 is %d bits",
	ErrKeyNilRandomSource:       "random source cannot be nil: use InitializePRNG(password) or provide a pre-seeded RandomSource",
}

// TranslatorProvider defines the interface for translation services.
// This interface allows the f5messageembed package to accept translation capabilities
// without creating a hard dependency on the full i18n package.
//
// To integrate i18n support, create a Translator using the i18n package and
// pass it to SetTranslator(). The f5messageembed package will use it for all error
// messages.
//
// # Example Integration
//
//	import (
//	    "github.com/0verkilll/f5messageembed"
//	    "github.com/0verkilll/i18n"
//	)
//
//	translator, _ := i18n.New(
//	    i18n.WithFileSystemLoader("locales"),
//	    i18n.WithDefaultLocale("en-US"),
//	)
//	f5messageembed.SetTranslator(translator)
//
// When no translator is set, the package falls back to English messages.
type TranslatorProvider interface {
	// Translate looks up a translation key in the current locale.
	// If the key is not found, it tries the fallback chain.
	// Returns the key itself if not found in any locale.
	Translate(key string) string

	// TranslateWithArgs looks up a translation key and formats it with arguments.
	// Uses fmt.Sprintf formatting. If the key is not found, returns the key itself.
	TranslateWithArgs(key string, args ...interface{}) string

	// HasKey checks if a translation key exists in the current locale or fallback chain.
	HasKey(key string) bool

	// SetLocale changes the current locale for translation lookups.
	// The locale will be normalized before being set.
	SetLocale(locale string)

	// GetLocale returns the current locale being used for translations.
	GetLocale() string
}

// globalTranslator holds the package-level translator instance.
// Stored via atomic.Pointer so read-heavy lookups (one per error translation)
// take the lock-free fast path. Writers go through Store; readers go through
// Load. The stored value is a non-nil *TranslatorProvider wrapper so Load can
// distinguish "never set" (nil pointer) from "set to nil" (pointer to nil
// interface) — SetTranslator(nil) clears the slot back to the nil-pointer
// state so readers see no translator.
var globalTranslator atomic.Pointer[TranslatorProvider]

// SetTranslator sets the global translator for i18n error messages.
// Pass nil to disable translation and use English fallback messages.
//
// This function is safe for concurrent use.
func SetTranslator(t TranslatorProvider) {
	if t == nil {
		globalTranslator.Store(nil)
		return
	}
	globalTranslator.Store(&t)
}

// GetTranslator returns the current global translator, or nil if not set.
// This function is safe for concurrent use.
func GetTranslator() TranslatorProvider {
	if p := globalTranslator.Load(); p != nil {
		return *p
	}
	return nil
}

// Translate returns a translated message for the given key.
// If no translator is set or the key is not found, returns the default English message.
// This is the public translation helper function for i18n support.
func Translate(key string) string {
	return translateError(key)
}

// translateError returns a translated error message for the given key.
// If no translator is set or the key is not found, returns the default English message.
func translateError(key string) string {
	if p := globalTranslator.Load(); p != nil {
		translated := (*p).Translate(key)
		// If translation returns the key itself, the key wasn't found
		if translated != key {
			return translated
		}
	}

	// Fallback to default English message
	if msg, ok := defaultMessages[key]; ok {
		return msg
	}

	// Last resort: return the key itself
	return key
}

// ValidationError represents an input validation error with i18n support.
type ValidationError struct {
	// Key is the i18n key for the error message
	Key string
	// Message is the translated or fallback error message
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return e.Message
}

// newValidationError creates a new ValidationError with translated message.
func newValidationError(key string) *ValidationError {
	return &ValidationError{
		Key:     key,
		Message: translateError(key),
	}
}

// ValidateCoefficients validates that the coefficient slice is valid for embedding.
// It checks that:
//   - The slice is not empty
//   - All coefficient values are within the valid JPEG DCT range [-2048, 2047]
//
// Returns nil if validation passes, or a ValidationError if it fails.
func ValidateCoefficients(coefficients []int16) error {
	if len(coefficients) == 0 {
		return newValidationError(ErrKeyEmptyCoefficients)
	}

	for _, coeff := range coefficients {
		if coeff < CoefficientMin || coeff > CoefficientMax {
			return newValidationError(ErrKeyInvalidCoefficientVal)
		}
	}

	return nil
}

// ValidatePassword validates that the password is valid for embedding.
// It checks that the password is not an empty string.
//
// Returns nil if validation passes, or a ValidationError if it fails.
func ValidatePassword(password string) error {
	if password == "" {
		return newValidationError(ErrKeyEmptyPassword)
	}
	return nil
}

// ValidateMessage validates that the message can be embedded within the given capacity.
// It checks that:
//   - The message size does not exceed MaxMessageSize (2^23 - 1 bytes)
//   - The message fits within the specified capacity in bytes
//
// The capacity parameter should be the maximum bytes that can be embedded,
// typically obtained from CapacityResult.CapacityByK for the selected k value.
//
// Returns nil if validation passes, or a ValidationError if it fails.
func ValidateMessage(message []byte, capacity int) error {
	if len(message) > MaxMessageSize {
		return newValidationError(ErrKeyMessageTooLarge)
	}

	if len(message) > capacity {
		return newValidationError(ErrKeyInsufficientCapacity)
	}

	return nil
}

// ValidateInputs performs all input validations for an embedding operation.
// This is a convenience function that combines ValidateCoefficients,
// ValidatePassword, and ValidateMessage.
//
// Returns the first validation error encountered, or nil if all validations pass.
func ValidateInputs(coefficients []int16, password string, message []byte, capacity int) error {
	if err := ValidateCoefficients(coefficients); err != nil {
		return err
	}

	if err := ValidatePassword(password); err != nil {
		return err
	}

	if err := ValidateMessage(message, capacity); err != nil {
		return err
	}

	return nil
}

// ValidationErrorf creates a validation error with a formatted message.
// This is useful for validation errors that need dynamic content.
// The key should be an i18n key, and args are passed to fmt.Sprintf
// after translation if the translator supports TranslateWithArgs.
func ValidationErrorf(key string, args ...interface{}) *ValidationError {
	var message string
	if p := globalTranslator.Load(); p != nil {
		translated := (*p).TranslateWithArgs(key, args...)
		if translated != key {
			message = translated
		}
	}

	if message == "" {
		// Fallback: use default message with formatting
		if msg, ok := defaultMessages[key]; ok {
			message = fmt.Sprintf(msg, args...)
		} else {
			message = key
		}
	}

	return &ValidationError{
		Key:     key,
		Message: message,
	}
}
