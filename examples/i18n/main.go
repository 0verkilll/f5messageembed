package main

import (
	"fmt"
	"log"

	"github.com/0verkilll/f5messageembed"
	"github.com/0verkilll/i18n"
)

func main() {
	fmt.Println("F5 Internationalization (i18n) Example")
	fmt.Println("=======================================")
	fmt.Println()

	// Example 1: Default behavior (no translator)
	fmt.Println("1. Default Behavior (No Translator)")
	fmt.Println("------------------------------------")
	demonstrateDefaultErrors()
	fmt.Println()

	// Example 2: Set up translator with f5messageembed locales
	fmt.Println("2. Using i18n Translator")
	fmt.Println("------------------------")
	demonstrateTranslator()
	fmt.Println()

	// Example 3: Triggering various validation errors
	fmt.Println("3. Validation Error Messages")
	fmt.Println("-----------------------------")
	demonstrateValidationErrors()
	fmt.Println()

	// Example 4: Manual translation access
	fmt.Println("4. Manual Translation Access")
	fmt.Println("-----------------------------")
	demonstrateManualTranslation()
	fmt.Println()

	// Clean up
	f5messageembed.SetTranslator(nil)
	fmt.Println("Translator reset to nil.")
}

// demonstrateDefaultErrors shows error messages without a translator.
func demonstrateDefaultErrors() {
	// Ensure no translator is set
	f5messageembed.SetTranslator(nil)

	fmt.Println("Without a translator, f5messageembed uses default English messages:")
	fmt.Println()

	// Trigger some validation errors
	_, err := f5messageembed.Embed(nil, "password", []byte("message"))
	if err != nil {
		fmt.Printf("  Empty coefficients error: %s\n", err.Error())
	}

	_, err = f5messageembed.Embed([]int16{1, 2, 3}, "", []byte("message"))
	if err != nil {
		fmt.Printf("  Empty password error:     %s\n", err.Error())
	}
}

// demonstrateTranslator shows how to set up and use i18n.
func demonstrateTranslator() {
	// Create a translator using the f5messageembed locales directory
	translator, err := i18n.New(
		i18n.WithFileSystemLoader("../../locales"),
		i18n.WithDefaultLocale("en-US"),
	)
	if err != nil {
		log.Printf("Failed to create translator: %v\n", err)
		fmt.Println("  Note: This is expected if running from a different directory.")
		fmt.Println("  In a real application, use the correct path to locales/")
		return
	}

	// Set the translator for f5messageembed
	f5messageembed.SetTranslator(translator)
	fmt.Println("Translator set for f5messageembed package.")

	// Now errors will use the translator
	_, err = f5messageembed.Embed(nil, "password", []byte("message"))
	if err != nil {
		fmt.Printf("  Translated error: %s\n", err.Error())
	}

	// Check the current translator
	currentTranslator := f5messageembed.GetTranslator()
	if currentTranslator != nil {
		fmt.Printf("  Current locale: %s\n", currentTranslator.GetLocale())
	}
}

// demonstrateValidationErrors shows all validation error messages.
func demonstrateValidationErrors() {
	// Use default messages for clarity
	f5messageembed.SetTranslator(nil)

	fmt.Println("All validation error messages:")
	fmt.Println()

	// Error 1: Empty coefficients
	_, err := f5messageembed.Embed(nil, "password", []byte("message"))
	if err != nil {
		fmt.Printf("  1. %s\n", err.Error())
	}

	// Error 2: Empty password
	_, err = f5messageembed.Embed([]int16{1, 2, 3}, "", []byte("message"))
	if err != nil {
		fmt.Printf("  2. %s\n", err.Error())
	}

	// Error 3: Invalid coefficient range
	invalidCoeffs := []int16{0, 1, 2, 3000} // 3000 is out of range
	_, err = f5messageembed.Embed(invalidCoeffs, "password", []byte("message"))
	if err != nil {
		fmt.Printf("  3. %s\n", err.Error())
	}

	// Error 4: Message too large for capacity
	smallCoeffs := make([]int16, 100)
	for i := range smallCoeffs {
		smallCoeffs[i] = int16((i % 10) + 1)
	}
	largeMessage := make([]byte, 1000)
	_, err = f5messageembed.Embed(smallCoeffs, "password", largeMessage)
	if err != nil {
		fmt.Printf("  4. %s\n", err.Error())
	}

	// Error 5: Invalid ForceK parameter
	validCoeffs := generateSyntheticCoefficients(10000)
	_, err = f5messageembed.EmbedWithOptions(
		validCoeffs,
		"password",
		[]byte("message"),
		f5messageembed.EmbedOptions{ForceK: 9}, // Invalid: must be 1-8
	)
	if err != nil {
		fmt.Printf("  5. %s\n", err.Error())
	}
}

// demonstrateManualTranslation shows how to access translations directly.
func demonstrateManualTranslation() {
	fmt.Println("Using the Translate function directly:")
	fmt.Println()

	// Without translator (fallback to default)
	f5messageembed.SetTranslator(nil)
	emptyCoeffMsg := f5messageembed.Translate(f5messageembed.ErrKeyEmptyCoefficients)
	emptyPwdMsg := f5messageembed.Translate(f5messageembed.ErrKeyEmptyPassword)

	fmt.Printf("  %s -> %s\n", f5messageembed.ErrKeyEmptyCoefficients, emptyCoeffMsg)
	fmt.Printf("  %s -> %s\n", f5messageembed.ErrKeyEmptyPassword, emptyPwdMsg)
	fmt.Println()

	fmt.Println("Available error keys:")
	fmt.Printf("  - %s\n", f5messageembed.ErrKeyEmptyCoefficients)
	fmt.Printf("  - %s\n", f5messageembed.ErrKeyEmptyPassword)
	fmt.Printf("  - %s\n", f5messageembed.ErrKeyMessageTooLarge)
	fmt.Printf("  - %s\n", f5messageembed.ErrKeyInsufficientCapacity)
	fmt.Printf("  - %s\n", f5messageembed.ErrKeyInvalidCoefficientVal)
	fmt.Printf("  - %s\n", f5messageembed.ErrKeyInvalidKParameter)
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
