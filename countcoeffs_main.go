//go:build ignore

// countcoeffs loads a JPEG and prints a full breakdown of its DCT coefficients:
// total, DC, zero AC, usable (non-zero AC), magnitude-1, and the F5 embedding
// capacity table for k=1..8.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/0verkilll/f5messageembed"
	"github.com/0verkilll/jpeg"
)

func main() {
	path := "../sample.jpg"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("read %s: %v", path, err)
	}

	dec := jpeg.NewStandardDecoder()
	coeffsInt, err := dec.ExtractCoefficients(data)
	if err != nil {
		log.Fatalf("extract coefficients: %v", err)
	}

	// Convert []int → []int16 (F5 works in int16 space)
	coeffs := make([]int16, len(coeffsInt))
	for i, v := range coeffsInt {
		coeffs[i] = int16(v) //nolint:gosec
	}

	// Raw pass: count every category manually so you can see the arithmetic.
	var (
		totalCoeffs int = len(coeffs)
		dcCount     int
		zeroAC      int
		usable      int
		magOne      int
	)
	for i, c := range coeffs {
		if i%64 == 0 {
			dcCount++
			continue
		}
		if c == 0 {
			zeroAC++
			continue
		}
		usable++
		if c == 1 || c == -1 {
			magOne++
		}
	}

	// Cross-check via the production CalculateCapacity call.
	cap := f5messageembed.CalculateCapacity(coeffs)

	fmt.Printf("File: %s (%d bytes on disk)\n\n", path, len(data))

	fmt.Println("=== Coefficient breakdown ===")
	fmt.Printf("  Total coefficients  : %d\n", totalCoeffs)
	fmt.Printf("  DC  (idx%%64==0)     : %d\n", dcCount)
	fmt.Printf("  Zero AC             : %d\n", zeroAC)
	fmt.Printf("  Usable (non-zero AC): %d\n", usable)
	fmt.Printf("  Magnitude-1 |c|=1   : %d  (%.2f%% of usable)\n",
		magOne, pct(magOne, usable))

	fmt.Println()
	fmt.Println("=== Production CalculateCapacity cross-check ===")
	fmt.Printf("  TotalCoefficients       : %d  (match=%v)\n",
		cap.TotalCoefficients, cap.TotalCoefficients == totalCoeffs)
	fmt.Printf("  UsableCoefficients      : %d  (match=%v)\n",
		cap.UsableCoefficients, cap.UsableCoefficients == usable)
	fmt.Printf("  MagnitudeOneCount h(1)  : %d  (match=%v)\n",
		cap.MagnitudeOneCount, cap.MagnitudeOneCount == magOne)
	fmt.Printf("  EstimatedShrinkageFactor: %.4f\n", cap.EstimatedShrinkageFactor)

	fmt.Println()
	fmt.Println("=== Capacity by k (message bytes, header overhead subtracted) ===")
	fmt.Printf("  %-4s  %-10s  %-10s  %s\n", "k", "n=2^k-1", "codewords", "capacity (bytes)")
	for k := 1; k <= 8; k++ {
		n := (1 << k) - 1
		codewords := usable / n
		fmt.Printf("  k=%-2d  n=%-8d  cw=%-8d  %d B\n", k, n, codewords, cap.CapacityByK[k])
	}
}

func pct(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return 100 * float64(part) / float64(total)
}
