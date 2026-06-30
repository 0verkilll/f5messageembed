package main

import (
	"fmt"
	"log"

	"github.com/0verkilll/f5messageembed"
)

func main() {
	fmt.Println("F5 Steganographic Embedding - Basic Example")
	fmt.Println("=============================================")
	fmt.Println()

	// Create synthetic JPEG DCT coefficients for demonstration
	// In real usage, these would come from a JPEG decoder
	coefficients := generateSyntheticCoefficients(10000)

	fmt.Printf("Generated %d synthetic coefficients\n", len(coefficients))

	// The secret message to embed
	message := []byte("Hello, World! This is a secret message.")
	password := "secret-password"

	fmt.Printf("Message to embed: %q\n", string(message))
	fmt.Printf("Message size: %d bytes\n", len(message))
	fmt.Printf("Password: %s\n", password)
	fmt.Println()

	// Embed the message
	result, err := f5messageembed.Embed(coefficients, password, message)
	if err != nil {
		log.Fatalf("Embedding failed: %v", err)
	}

	// Display results
	fmt.Println("Embedding Results:")
	fmt.Println("------------------")
	fmt.Printf("  Bytes embedded:      %d\n", result.BytesEmbedded)
	fmt.Printf("  K parameter used:    %d\n", result.KParameter)
	fmt.Printf("  Usable coefficients: %d\n", result.UsableCoefficients)
	fmt.Printf("  Shrinkage count:     %d\n", result.ShrinkageCount)
	fmt.Println()

	// Demonstrate coefficient modification
	fmt.Println("Coefficient Analysis:")
	fmt.Println("---------------------")
	modifiedCount := countModifiedCoefficients(coefficients, result.Coefficients)
	fmt.Printf("  Total coefficients:    %d\n", len(coefficients))
	fmt.Printf("  Modified coefficients: %d\n", modifiedCount)
	fmt.Printf("  Modification rate:     %.2f%%\n", float64(modifiedCount)*100/float64(len(coefficients)))
	fmt.Println()

	fmt.Println("Success! The message has been embedded into the coefficients.")
	fmt.Println("In a real application, you would now encode these coefficients")
	fmt.Println("back into a JPEG image file.")
}

// generateSyntheticCoefficients creates a coefficient array simulating JPEG DCT values.
// The distribution roughly mimics typical JPEG DCT coefficients.
func generateSyntheticCoefficients(count int) []int16 {
	coefficients := make([]int16, count)

	// Generate a realistic distribution of DCT coefficients
	// Most coefficients are small values close to zero
	for i := range coefficients {
		// DC coefficients at every 64th position
		if i%64 == 0 {
			coefficients[i] = int16(100 + (i % 200))
			continue
		}

		// AC coefficients with exponential-like distribution
		// Mostly small values with some larger ones
		switch {
		case i%7 == 0:
			coefficients[i] = 0 // Many zeros in typical JPEG
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

// countModifiedCoefficients counts how many coefficients were changed during embedding.
func countModifiedCoefficients(original, modified []int16) int {
	count := 0
	for i := range original {
		if original[i] != modified[i] {
			count++
		}
	}
	return count
}
