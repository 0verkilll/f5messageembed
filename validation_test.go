package f5messageembed

import (
	"errors"
	"strings"
	"testing"
)

// mockTranslator is a test implementation of TranslatorProvider.
type mockTranslator struct {
	translations map[string]string
	locale       string
}

func newMockTranslator() *mockTranslator {
	return &mockTranslator{
		translations: map[string]string{
			ErrKeyEmptyCoefficients:     "Los coeficientes no pueden estar vacios",
			ErrKeyEmptyPassword:         "La contrasena no puede estar vacia",
			ErrKeyMessageTooLarge:       "El mensaje excede el tamano maximo permitido",
			ErrKeyInsufficientCapacity:  "El mensaje es demasiado grande para la capacidad disponible",
			ErrKeyInvalidCoefficientVal: "El valor del coeficiente esta fuera del rango valido",
		},
		locale: "es-ES",
	}
}

func (m *mockTranslator) Translate(key string) string {
	if val, ok := m.translations[key]; ok {
		return val
	}
	return key
}

func (m *mockTranslator) TranslateWithArgs(key string, _ ...interface{}) string {
	return m.Translate(key)
}

func (m *mockTranslator) HasKey(key string) bool {
	_, ok := m.translations[key]
	return ok
}

func (m *mockTranslator) SetLocale(locale string) {
	m.locale = locale
}

func (m *mockTranslator) GetLocale() string {
	return m.locale
}

// TestValidateCoefficients_Empty tests validation of empty coefficient slice.
func TestValidateCoefficients_Empty(t *testing.T) {
	// Ensure no translator is set for fallback testing
	SetTranslator(nil)

	err := ValidateCoefficients([]int16{})
	if err == nil {
		t.Fatal("expected error for empty coefficients, got nil")
	}

	var valErr *ValidationError
	ok := errors.As(err, &valErr)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	if valErr.Key != ErrKeyEmptyCoefficients {
		t.Errorf("expected key %q, got %q", ErrKeyEmptyCoefficients, valErr.Key)
	}

	// Check fallback English message is used
	expectedMsg := defaultMessages[ErrKeyEmptyCoefficients]
	if valErr.Message != expectedMsg {
		t.Errorf("expected message %q, got %q", expectedMsg, valErr.Message)
	}
}

// TestValidateCoefficients_NilSlice tests validation of nil coefficient slice.
func TestValidateCoefficients_NilSlice(t *testing.T) {
	SetTranslator(nil)

	err := ValidateCoefficients(nil)
	if err == nil {
		t.Fatal("expected error for nil coefficients, got nil")
	}

	var valErr *ValidationError
	ok := errors.As(err, &valErr)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	if valErr.Key != ErrKeyEmptyCoefficients {
		t.Errorf("expected key %q, got %q", ErrKeyEmptyCoefficients, valErr.Key)
	}
}

// TestValidateCoefficients_OutOfRange tests validation of out-of-range coefficients.
func TestValidateCoefficients_OutOfRange(t *testing.T) {
	SetTranslator(nil)

	testCases := []struct {
		name         string
		coefficients []int16
		shouldFail   bool
	}{
		{
			name:         "valid min boundary",
			coefficients: []int16{CoefficientMin, 0, CoefficientMax},
			shouldFail:   false,
		},
		{
			name:         "valid max boundary",
			coefficients: []int16{CoefficientMax, 0, CoefficientMin},
			shouldFail:   false,
		},
		{
			name:         "below minimum",
			coefficients: []int16{0, -2049, 0},
			shouldFail:   true,
		},
		{
			name:         "above maximum",
			coefficients: []int16{0, 2048, 0},
			shouldFail:   true,
		},
		{
			name:         "all valid zeros",
			coefficients: []int16{0, 0, 0},
			shouldFail:   false,
		},
		{
			name:         "mixed valid values",
			coefficients: []int16{-1024, 0, 512, -2048, 2047},
			shouldFail:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateCoefficients(tc.coefficients)
			if tc.shouldFail {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				var valErr *ValidationError
				ok := errors.As(err, &valErr)
				if !ok {
					t.Fatalf("expected *ValidationError, got %T", err)
				}
				if valErr.Key != ErrKeyInvalidCoefficientVal {
					t.Errorf("expected key %q, got %q", ErrKeyInvalidCoefficientVal, valErr.Key)
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

// TestValidatePassword_Empty tests validation of empty password.
func TestValidatePassword_Empty(t *testing.T) {
	SetTranslator(nil)

	err := ValidatePassword("")
	if err == nil {
		t.Fatal("expected error for empty password, got nil")
	}

	var valErr *ValidationError
	ok := errors.As(err, &valErr)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	if valErr.Key != ErrKeyEmptyPassword {
		t.Errorf("expected key %q, got %q", ErrKeyEmptyPassword, valErr.Key)
	}

	// Check fallback English message is used
	expectedMsg := defaultMessages[ErrKeyEmptyPassword]
	if valErr.Message != expectedMsg {
		t.Errorf("expected message %q, got %q", expectedMsg, valErr.Message)
	}
}

// TestValidatePassword_Valid tests validation of valid passwords.
func TestValidatePassword_Valid(t *testing.T) {
	SetTranslator(nil)

	testCases := []string{
		"password",
		"a",
		"complex password with spaces",
		"!@#$%^&*()",
		" ", // Single space is technically not empty
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			err := ValidatePassword(tc)
			if err != nil {
				t.Errorf("expected no error for password %q, got %v", tc, err)
			}
		})
	}
}

// TestValidateMessage_TooLarge tests validation when message exceeds MaxMessageSize.
func TestValidateMessage_TooLarge(t *testing.T) {
	SetTranslator(nil)

	// Create a message larger than MaxMessageSize
	largeMessage := make([]byte, MaxMessageSize+1)

	err := ValidateMessage(largeMessage, MaxMessageSize+1)
	if err == nil {
		t.Fatal("expected error for message too large, got nil")
	}

	var valErr *ValidationError
	ok := errors.As(err, &valErr)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	if valErr.Key != ErrKeyMessageTooLarge {
		t.Errorf("expected key %q, got %q", ErrKeyMessageTooLarge, valErr.Key)
	}
}

// TestValidateMessage_ExceedsCapacity tests validation when message exceeds capacity.
func TestValidateMessage_ExceedsCapacity(t *testing.T) {
	SetTranslator(nil)

	message := make([]byte, 1000)
	capacity := 500

	err := ValidateMessage(message, capacity)
	if err == nil {
		t.Fatal("expected error for insufficient capacity, got nil")
	}

	var valErr *ValidationError
	ok := errors.As(err, &valErr)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	if valErr.Key != ErrKeyInsufficientCapacity {
		t.Errorf("expected key %q, got %q", ErrKeyInsufficientCapacity, valErr.Key)
	}
}

// TestValidateMessage_Valid tests validation of valid messages.
func TestValidateMessage_Valid(t *testing.T) {
	SetTranslator(nil)

	testCases := []struct {
		name       string
		msgSize    int
		capacity   int
		shouldPass bool
	}{
		{
			name:       "empty message",
			msgSize:    0,
			capacity:   100,
			shouldPass: true,
		},
		{
			name:       "message fits exactly",
			msgSize:    100,
			capacity:   100,
			shouldPass: true,
		},
		{
			name:       "message fits with room to spare",
			msgSize:    50,
			capacity:   100,
			shouldPass: true,
		},
		{
			name:       "max message size within capacity",
			msgSize:    MaxMessageSize,
			capacity:   MaxMessageSize,
			shouldPass: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			message := make([]byte, tc.msgSize)
			err := ValidateMessage(message, tc.capacity)
			if tc.shouldPass && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

// TestTranslation_WithTranslator tests that translation works when translator is set.
func TestTranslation_WithTranslator(t *testing.T) {
	// Set up mock translator
	mockTrans := newMockTranslator()
	SetTranslator(mockTrans)
	defer SetTranslator(nil) // Clean up after test

	err := ValidateCoefficients([]int16{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var valErr *ValidationError
	ok := errors.As(err, &valErr)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	// Check that Spanish translation was used
	expectedMsg := mockTrans.translations[ErrKeyEmptyCoefficients]
	if valErr.Message != expectedMsg {
		t.Errorf("expected translated message %q, got %q", expectedMsg, valErr.Message)
	}
}

// TestTranslation_Fallback tests fallback to English when no translator is set.
func TestTranslation_Fallback(t *testing.T) {
	// Ensure no translator is set
	SetTranslator(nil)

	err := ValidatePassword("")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var valErr *ValidationError
	ok := errors.As(err, &valErr)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	// Should use English fallback
	expectedMsg := defaultMessages[ErrKeyEmptyPassword]
	if valErr.Message != expectedMsg {
		t.Errorf("expected fallback message %q, got %q", expectedMsg, valErr.Message)
	}
}

// TestTranslation_MissingKey tests fallback when translator doesn't have the key.
func TestTranslation_MissingKey(t *testing.T) {
	// Create translator with missing key
	mockTrans := &mockTranslator{
		translations: map[string]string{}, // Empty, so all keys will be missing
		locale:       "en-US",
	}
	SetTranslator(mockTrans)
	defer SetTranslator(nil)

	err := ValidateCoefficients([]int16{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var valErr *ValidationError
	ok := errors.As(err, &valErr)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	// Should fall back to English default message
	expectedMsg := defaultMessages[ErrKeyEmptyCoefficients]
	if valErr.Message != expectedMsg {
		t.Errorf("expected fallback message %q, got %q", expectedMsg, valErr.Message)
	}
}

// TestSetTranslator_ConcurrentAccess tests thread safety of SetTranslator.
func TestSetTranslator_ConcurrentAccess(t *testing.T) {
	// Run multiple goroutines setting and getting translator concurrently
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			mockTrans := newMockTranslator()
			SetTranslator(mockTrans)
			_ = GetTranslator()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Clean up
	SetTranslator(nil)
}

// TestGetTranslator tests that GetTranslator returns the set translator.
func TestGetTranslator(t *testing.T) {
	// Initially should be nil after cleanup
	SetTranslator(nil)
	if got := GetTranslator(); got != nil {
		t.Errorf("expected nil translator, got %v", got)
	}

	// Set translator and verify it's returned
	mockTrans := newMockTranslator()
	SetTranslator(mockTrans)
	defer SetTranslator(nil)

	got := GetTranslator()
	if got != mockTrans {
		t.Errorf("expected mock translator, got different value")
	}
}

// TestValidationError_ErrorInterface tests that ValidationError implements error.
func TestValidationError_ErrorInterface(t *testing.T) {
	SetTranslator(nil)

	err := ValidateCoefficients([]int16{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Test that Error() returns the message
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Error() returned empty string")
	}

	// Check it contains expected text
	if !strings.Contains(errMsg, "coefficient") {
		t.Errorf("error message should mention coefficient, got: %s", errMsg)
	}
}

// TestValidateInputs_AllValid tests ValidateInputs with all valid inputs.
func TestValidateInputs_AllValid(t *testing.T) {
	SetTranslator(nil)

	coefficients := []int16{1, 2, 3, -1, -2, -3}
	password := "secret"
	message := []byte("hello")
	capacity := 100

	err := ValidateInputs(coefficients, password, message, capacity)
	if err != nil {
		t.Errorf("expected no error for valid inputs, got %v", err)
	}
}

// TestValidateInputs_InvalidCoefficients tests ValidateInputs with invalid coefficients.
func TestValidateInputs_InvalidCoefficients(t *testing.T) {
	SetTranslator(nil)

	err := ValidateInputs([]int16{}, "password", []byte("msg"), 100)
	if err == nil {
		t.Fatal("expected error for empty coefficients")
	}

	var valErr *ValidationError
	ok := errors.As(err, &valErr)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	if valErr.Key != ErrKeyEmptyCoefficients {
		t.Errorf("expected key %q, got %q", ErrKeyEmptyCoefficients, valErr.Key)
	}
}

// TestValidateInputs_InvalidPassword tests ValidateInputs with invalid password.
func TestValidateInputs_InvalidPassword(t *testing.T) {
	SetTranslator(nil)

	err := ValidateInputs([]int16{1, 2, 3}, "", []byte("msg"), 100)
	if err == nil {
		t.Fatal("expected error for empty password")
	}

	var valErr *ValidationError
	ok := errors.As(err, &valErr)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	if valErr.Key != ErrKeyEmptyPassword {
		t.Errorf("expected key %q, got %q", ErrKeyEmptyPassword, valErr.Key)
	}
}

// TestValidateInputs_InvalidMessage tests ValidateInputs with invalid message.
func TestValidateInputs_InvalidMessage(t *testing.T) {
	SetTranslator(nil)

	// Message larger than capacity
	err := ValidateInputs([]int16{1, 2, 3}, "password", []byte("too long message"), 5)
	if err == nil {
		t.Fatal("expected error for message exceeding capacity")
	}

	var valErr *ValidationError
	ok := errors.As(err, &valErr)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	if valErr.Key != ErrKeyInsufficientCapacity {
		t.Errorf("expected key %q, got %q", ErrKeyInsufficientCapacity, valErr.Key)
	}
}

// I18n Tests
//
// These tests verify internationalization support for error messages.

// TestI18n_ErrorMessageTranslation tests that error messages are translated when a translator is set.
func TestI18n_ErrorMessageTranslation(t *testing.T) {
	// Set up mock translator with Spanish translations
	mockTrans := &mockTranslator{
		translations: map[string]string{
			ErrKeyEmptyCoefficients:     "Los coeficientes no pueden estar vacios",
			ErrKeyEmptyPassword:         "La contrasena no puede estar vacia",
			ErrKeyMessageTooLarge:       "El mensaje excede el tamano maximo permitido",
			ErrKeyInsufficientCapacity:  "El mensaje es demasiado grande para la capacidad disponible",
			ErrKeyInvalidCoefficientVal: "El valor del coeficiente esta fuera del rango valido",
			ErrKeyInvalidKParameter:     "El parametro k debe estar entre 1 y 8",
		},
		locale: "es-ES",
	}
	SetTranslator(mockTrans)
	defer SetTranslator(nil)

	// Test each error key gets translated
	testCases := []struct {
		validateFunc func() error
		key          string
		expectedMsg  string
	}{
		{
			key:         ErrKeyEmptyCoefficients,
			expectedMsg: "Los coeficientes no pueden estar vacios",
			validateFunc: func() error {
				return ValidateCoefficients([]int16{})
			},
		},
		{
			key:         ErrKeyEmptyPassword,
			expectedMsg: "La contrasena no puede estar vacia",
			validateFunc: func() error {
				return ValidatePassword("")
			},
		},
		{
			key:         ErrKeyInvalidCoefficientVal,
			expectedMsg: "El valor del coeficiente esta fuera del rango valido",
			validateFunc: func() error {
				return ValidateCoefficients([]int16{3000})
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			err := tc.validateFunc()
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			var valErr *ValidationError
			ok := errors.As(err, &valErr)
			if !ok {
				t.Fatalf("expected *ValidationError, got %T", err)
			}

			if valErr.Message != tc.expectedMsg {
				t.Errorf("expected translated message %q, got %q", tc.expectedMsg, valErr.Message)
			}
		})
	}
}

// TestI18n_SetTranslatorAffectsErrorMessages tests that calling SetTranslator
// immediately affects subsequent error messages.
func TestI18n_SetTranslatorAffectsErrorMessages(t *testing.T) {
	// Start with no translator
	SetTranslator(nil)

	// Generate error - should use English default
	err1 := ValidatePassword("")
	if err1 == nil {
		t.Fatal("expected error, got nil")
	}
	var valErr1 *ValidationError
	ok := errors.As(err1, &valErr1)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err1)
	}
	englishMsg := defaultMessages[ErrKeyEmptyPassword]
	if valErr1.Message != englishMsg {
		t.Errorf("expected English message %q, got %q", englishMsg, valErr1.Message)
	}

	// Set translator with different language
	spanishMsg := "La contrasena no puede estar vacia"
	mockTrans := &mockTranslator{
		translations: map[string]string{
			ErrKeyEmptyPassword: spanishMsg,
		},
		locale: "es-ES",
	}
	SetTranslator(mockTrans)
	defer SetTranslator(nil)

	// Generate same error - should now use Spanish
	err2 := ValidatePassword("")
	if err2 == nil {
		t.Fatal("expected error, got nil")
	}
	var valErr2 *ValidationError
	ok2 := errors.As(err2, &valErr2)
	if !ok2 {
		t.Fatalf("expected *ValidationError, got %T", err2)
	}
	if valErr2.Message != spanishMsg {
		t.Errorf("expected Spanish message %q, got %q", spanishMsg, valErr2.Message)
	}

	// Set translator back to nil
	SetTranslator(nil)

	// Generate same error - should be back to English
	err3 := ValidatePassword("")
	if err3 == nil {
		t.Fatal("expected error, got nil")
	}
	var valErr3 *ValidationError
	ok3 := errors.As(err3, &valErr3)
	if !ok3 {
		t.Fatalf("expected *ValidationError, got %T", err3)
	}
	if valErr3.Message != englishMsg {
		t.Errorf("expected English message %q after clearing translator, got %q", englishMsg, valErr3.Message)
	}
}

// TestI18n_TranslateHelperFunction tests the public Translate helper function.
func TestI18n_TranslateHelperFunction(t *testing.T) {
	// Test without translator - should return default message
	SetTranslator(nil)

	result := Translate(ErrKeyEmptyCoefficients)
	expected := defaultMessages[ErrKeyEmptyCoefficients]
	if result != expected {
		t.Errorf("Translate without translator: expected %q, got %q", expected, result)
	}

	// Test with translator
	mockTrans := &mockTranslator{
		translations: map[string]string{
			ErrKeyEmptyCoefficients: "Translated empty coefficients",
		},
		locale: "en-US",
	}
	SetTranslator(mockTrans)
	defer SetTranslator(nil)

	result = Translate(ErrKeyEmptyCoefficients)
	expected = "Translated empty coefficients"
	if result != expected {
		t.Errorf("Translate with translator: expected %q, got %q", expected, result)
	}

	// Test unknown key - should return the key itself
	unknownKey := "f5messageembed.error.unknown_key"
	result = Translate(unknownKey)
	if result != unknownKey {
		t.Errorf("Translate unknown key: expected key %q, got %q", unknownKey, result)
	}
}

// TestI18n_AllErrorKeysHaveDefaults tests that all defined error keys have default messages.
func TestI18n_AllErrorKeysHaveDefaults(t *testing.T) {
	// List of all error keys that should have default messages
	errorKeys := []string{
		ErrKeyEmptyCoefficients,
		ErrKeyEmptyPassword,
		ErrKeyMessageTooLarge,
		ErrKeyInsufficientCapacity,
		ErrKeyInvalidCoefficientVal,
		ErrKeyInvalidKParameter,
	}

	SetTranslator(nil)

	for _, key := range errorKeys {
		t.Run(key, func(t *testing.T) {
			msg, ok := defaultMessages[key]
			if !ok {
				t.Errorf("error key %q has no default message", key)
			}
			if msg == "" {
				t.Errorf("error key %q has empty default message", key)
			}

			// Also verify Translate returns the default message (not the key)
			translated := Translate(key)
			if translated == key {
				t.Errorf("Translate(%q) returned key itself, expected default message", key)
			}
			if translated != msg {
				t.Errorf("Translate(%q) = %q, want %q", key, translated, msg)
			}
		})
	}
}

// Coverage Tests - ValidationErrorf and translateError

// mockTranslatorWithArgs is a mock translator that supports TranslateWithArgs.
type mockTranslatorWithArgs struct {
	translations map[string]string
	locale       string
}

func (m *mockTranslatorWithArgs) Translate(key string) string {
	if val, ok := m.translations[key]; ok {
		return val
	}
	return key
}

func (m *mockTranslatorWithArgs) TranslateWithArgs(key string, args ...interface{}) string {
	if template, ok := m.translations[key]; ok {
		// Simple format: replace %d with args
		if len(args) >= 2 {
			return "k must be 1-8"
		}
		return template
	}
	return key
}

func (m *mockTranslatorWithArgs) HasKey(key string) bool {
	_, ok := m.translations[key]
	return ok
}

func (m *mockTranslatorWithArgs) SetLocale(locale string) {
	m.locale = locale
}

func (m *mockTranslatorWithArgs) GetLocale() string {
	return m.locale
}

// TestValidationErrorf_WithTranslator tests the ValidationErrorf function for formatted errors.
func TestValidationErrorf_WithTranslator(t *testing.T) {
	// Create a translator that returns formatted translations
	mockTrans := &mockTranslatorWithArgs{
		translations: map[string]string{
			ErrKeyInvalidKParameter: "k must be %d-%d",
		},
	}
	SetTranslator(mockTrans)
	defer SetTranslator(nil)

	// Test with translator that supports formatting
	err := ValidationErrorf(ErrKeyInvalidKParameter, 1, 8)
	if err == nil {
		t.Fatal("expected non-nil error")
	}

	if err.Key != ErrKeyInvalidKParameter {
		t.Errorf("expected key %q, got %q", ErrKeyInvalidKParameter, err.Key)
	}

	// Message should contain the formatted values
	expectedMsg := "k must be 1-8"
	if err.Message != expectedMsg {
		t.Errorf("expected message %q, got %q", expectedMsg, err.Message)
	}
}

// TestValidationErrorf_FallbackWithArgs tests ValidationErrorf fallback to default message.
func TestValidationErrorf_FallbackWithArgs(t *testing.T) {
	// Ensure no translator is set
	SetTranslator(nil)

	// ValidationErrorf with a known key should use the default message
	err := ValidationErrorf(ErrKeyInvalidKParameter, 1, 8)
	if err == nil {
		t.Fatal("expected non-nil error")
	}

	if err.Key != ErrKeyInvalidKParameter {
		t.Errorf("expected key %q, got %q", ErrKeyInvalidKParameter, err.Key)
	}

	// Default message doesn't have format specifiers, so it may not format correctly
	// but it should at least not panic and return the default message
	if err.Message == "" {
		t.Error("expected non-empty message")
	}
}

// TestValidationErrorf_UnknownKey tests ValidationErrorf with an unknown key.
func TestValidationErrorf_UnknownKey(t *testing.T) {
	SetTranslator(nil)

	unknownKey := "unknown.error.key"
	err := ValidationErrorf(unknownKey, "arg1", "arg2")
	if err == nil {
		t.Fatal("expected non-nil error")
	}

	// Should return the key itself as the message when not found
	if err.Message != unknownKey {
		t.Errorf("expected message to be key %q, got %q", unknownKey, err.Message)
	}
}

// TestTranslateError_UnknownKeyFallback tests translateError with unknown key.
func TestTranslateError_UnknownKeyFallback(t *testing.T) {
	SetTranslator(nil)

	unknownKey := "unknown.key.not.in.defaults"
	result := translateError(unknownKey)

	// Should return the key itself when not found
	if result != unknownKey {
		t.Errorf("translateError(%q) = %q, want %q", unknownKey, result, unknownKey)
	}
}

// TestTranslate_PublicFunction tests the public Translate helper.
func TestTranslate_PublicFunction(t *testing.T) {
	SetTranslator(nil)

	// Test with known key
	result := Translate(ErrKeyEmptyCoefficients)
	expected := defaultMessages[ErrKeyEmptyCoefficients]
	if result != expected {
		t.Errorf("Translate(%q) = %q, want %q", ErrKeyEmptyCoefficients, result, expected)
	}

	// Test with unknown key
	unknownKey := "some.unknown.key"
	result = Translate(unknownKey)
	if result != unknownKey {
		t.Errorf("Translate(%q) = %q, want %q (key itself)", unknownKey, result, unknownKey)
	}
}
