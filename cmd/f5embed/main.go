// f5embed embeds a secret message into a JPEG using the F5 algorithm.
//
// Usage:
//
//	f5embed -in cover.jpg -out stego.jpg -pass "password" -msg "your message"
//
// Flags:
//
//	-in   input JPEG cover file (required)
//	-out  output JPEG stego file (default: <basename>-stego.jpg)
//	-pass password for the F5 permutation (default: abc123)
//	-msg  message to embed
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	f5 "github.com/0verkilll/f5messageembed"
	jpeglib "github.com/0verkilll/jpeg"
)

func main() {
	inPath := flag.String("in", "", "input JPEG cover image (required)")
	outPath := flag.String("out", "", "output JPEG stego image (default: <name>-stego.jpg)")
	password := flag.String("pass", "abc123", "embedding password")
	message := flag.String("msg", "hello, world", "message to embed")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: f5embed -in cover.jpg -pass <password> -msg <text> [-out stego.jpg]\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *inPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *outPath == "" {
		ext := filepath.Ext(*inPath)
		base := strings.TrimSuffix(*inPath, ext)
		*outPath = base + "-stego" + ext
	}

	data, err := os.ReadFile(*inPath)
	if err != nil {
		log.Fatalf("read %s: %v", *inPath, err)
	}

	dec := jpeglib.NewStandardDecoder()
	coeffsInt, err := dec.ExtractCoefficients(data)
	if err != nil {
		log.Fatalf("extract coefficients: %v", err)
	}

	coeffs := make([]int16, len(coeffsInt))
	for i, v := range coeffsInt {
		coeffs[i] = int16(v) //nolint:gosec
	}

	fmt.Printf("cover:    %s  (%d bytes, %d coefficients)\n", *inPath, len(data), len(coeffs))
	fmt.Printf("password: %q\n", *password)
	fmt.Printf("message:  %q  (%d bytes)\n", *message, len(*message))
	fmt.Println()

	result, err := f5.Embed(coeffs, *password, []byte(*message))
	if err != nil {
		log.Fatalf("F5 embed: %v", err)
	}

	outInts := make([]int, len(result.Coefficients))
	for i, v := range result.Coefficients {
		outInts[i] = int(v)
	}

	stegoBytes, err := dec.EncodeCoefficients(outInts)
	if err != nil {
		log.Fatalf("re-encode: %v", err)
	}

	if err := os.WriteFile(*outPath, stegoBytes, 0o644); err != nil {
		log.Fatalf("write %s: %v", *outPath, err)
	}

	fmt.Printf("k parameter:    %d\n", result.KParameter)
	fmt.Printf("bytes embedded: %d\n", result.BytesEmbedded)
	fmt.Printf("usable coeffs:  %d\n", result.UsableCoefficients)
	fmt.Printf("shrinkage:      %d\n", result.ShrinkageCount)
	fmt.Printf("\nwrote: %s  (%d bytes)\n", *outPath, len(stegoBytes))
}
