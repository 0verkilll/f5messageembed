package main

import (
	"crypto/sha1"
	"fmt"

	"github.com/0verkilll/f5prng"
)

func main() {
	password := "abc123"

	// Step 1: show initial state = SHA1(password)
	h := sha1.New()
	h.Write([]byte(password))
	initialState := h.Sum(nil)
	fmt.Printf("=== STEP 1: Seed ===\n")
	fmt.Printf("SHA1(%q) = %x\n\n", password, initialState)

	// Step 2: show first few PRNG outputs (what updateState produces)
	factory := f5prng.NewDefaultFactory()
	prng := factory.NewPRNG()
	defer prng.Clear()
	_ = prng.Seed([]byte(password))

	fmt.Printf("=== STEP 2: First PRNG outputs (each is SHA1 of evolving state) ===\n")
	for i := 0; i < 5; i++ {
		b := prng.NextBytes(20)
		fmt.Printf("Block %d: %x\n", i, b)
	}

	// Reset and show Fisher-Yates draws for a small N=10 permutation
	_ = prng.Seed([]byte(password))
	fmt.Printf("\n=== STEP 3: Fisher-Yates shuffle for N=10 ===\n")
	fmt.Printf("Start: perm = [0 1 2 3 4 5 6 7 8 9]\n")
	perm := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	for maxRandom := len(perm); maxRandom > 0; maxRandom-- {
		raw := prng.NextInt()
		idx := int(raw) % maxRandom
		if idx < 0 {
			idx += maxRandom
		}
		swapPos := maxRandom - 1
		fmt.Printf("  maxRandom=%2d  NextInt()=%12d  idx=%d  swap perm[%d]<->perm[%d]  → %v\n",
			maxRandom, raw, idx, idx, swapPos, perm)
		perm[idx], perm[swapPos] = perm[swapPos], perm[idx]
	}
	fmt.Printf("Final perm: %v\n", perm)
	fmt.Printf("Embedding walks this left to right, skipping DC (pos%%64==0) and zero coefficients\n")

	// Reset and show header XOR bytes (drawn after permutation)
	_ = prng.Seed([]byte(password))
	fmt.Printf("\n=== STEP 4: Header XOR bytes (drawn after permutation, N=10 for demo) ===\n")
	// consume N=10 permutation draws
	for i := 0; i < 10; i++ {
		prng.NextInt()
	}
	headerXorBytes := prng.NextBytes(4)
	fmt.Printf("Header XOR bytes: %x\n", headerXorBytes)

	// Build the header for k=1 (auto for 39 bytes), messageSize=39
	// k=1 is typical for small messages; show what the raw header looks like
	k := 1
	msgLen := 39
	rawHeader := int32(k<<24) | int32(msgLen&0x7FFFFF)
	fmt.Printf("\n=== STEP 5: Header for k=%d, messageLen=%d ===\n", k, msgLen)
	fmt.Printf("Raw header:  0x%08X  (bits 24-31 = k=%d, bits 0-22 = len=%d)\n", uint32(rawHeader), k, msgLen)

	xored := rawHeader
	xored ^= int32(int8(headerXorBytes[0]))
	xored ^= int32(int8(headerXorBytes[1])) << 8
	xored ^= int32(int8(headerXorBytes[2])) << 16
	xored ^= int32(int8(headerXorBytes[3])) << 24
	fmt.Printf("XORed header: 0x%08X  (this is what gets embedded into coefficients)\n", uint32(xored))
	fmt.Printf("  byte0: 0x%02X ^ 0x%02X = 0x%02X\n", rawHeader&0xFF, headerXorBytes[0], xored&0xFF)
	fmt.Printf("  byte1: 0x%02X ^ 0x%02X = 0x%02X\n", (rawHeader>>8)&0xFF, headerXorBytes[1], (xored>>8)&0xFF)
	fmt.Printf("  byte2: 0x%02X ^ 0x%02X = 0x%02X\n", (rawHeader>>16)&0xFF, headerXorBytes[2], (xored>>16)&0xFF)
	fmt.Printf("  byte3: 0x%02X ^ 0x%02X = 0x%02X\n", (rawHeader>>24)&0xFF, headerXorBytes[3], (xored>>24)&0xFF)

	fmt.Printf("\n=== STEP 6: Message XOR stream (first 10 bytes, drawn after header XOR) ===\n")
	msgXor := prng.NextBytes(10)
	fmt.Printf("First 10 message XOR bytes: %x\n", msgXor)
	fmt.Printf("Each message byte gets XORed with one of these before being embedded as bits\n")
}
