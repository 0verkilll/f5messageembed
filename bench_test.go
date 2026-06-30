package f5messageembed

import (
	"testing"

	"github.com/0verkilll/f5prng"
)

// BenchmarkEmbed_MessageXOR_Medium exercises the Embed hot path with a
// medium-sized message so the per-byte XOR-stream loop in newEmbeddingState
// dominates allocation counts. Before the bulk-NextBytes refactor each byte
// consumed its own one-byte slice from the PRNG; after the refactor the
// entire XOR stream is produced in one call (or fills xoredMessage directly
// via NextBytesInto when the source supports it).
func BenchmarkEmbed_MessageXOR_Medium(b *testing.B) {
	message := make([]byte, 1024)
	for i := range message {
		message[i] = byte(i * 17)
	}
	password := "bench-password-medium"
	// 256 KiB of coefficients — plenty of headroom at default k, avoids the
	// capacity selection from influencing the measured path.
	base := createJavaCompatCoefficients(256 * 1024)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Fresh coefficient buffer per iteration — Embed modifies in place.
		coefficients := make([]int16, len(base))
		copy(coefficients, base)
		b.StartTimer()

		_, err := Embed(coefficients, password, message)
		if err != nil {
			b.Fatalf("Embed failed: %v", err)
		}
	}
}

// BenchmarkEmbed_MessageXOR_Short focuses on the short-message regime where
// the per-byte allocation pattern in the pre-refactor code was most
// measurable (each byte paid the full PRNG-dispatch overhead). A 16-byte
// message sits right in the sweet spot.
func BenchmarkEmbed_MessageXOR_Short(b *testing.B) {
	message := []byte("sixteen bytes!!!")
	password := "bench-password-short"
	base := createJavaCompatCoefficients(32 * 1024)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		coefficients := make([]int16, len(base))
		copy(coefficients, base)
		b.StartTimer()

		_, err := Embed(coefficients, password, message)
		if err != nil {
			b.Fatalf("Embed failed: %v", err)
		}
	}
}

// BenchmarkXORStream_PerByte_OLD reproduces the pre-refactor per-byte
// NextBytes(1) loop so the delta against the bulk implementation is
// measurable without having to stash/unstash the package. This is the
// strategy used before the bulk refactor; it is kept only as a benchmark
// baseline and is not referenced by production code.
func BenchmarkXORStream_PerByte_OLD(b *testing.B) {
	sizes := []int{16, 256, 1024, 4096}
	for _, size := range sizes {
		b.Run(sizeLabel(size), func(b *testing.B) {
			message := make([]byte, size)
			factory := f5prng.NewDefaultFactory()

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				prng := factory.NewPRNG()
				if err := prng.Seed([]byte("bench-password")); err != nil {
					b.Fatal(err)
				}

				xoredMessage := make([]byte, len(message))
				for j := range message {
					xorBytes := prng.NextBytes(1)
					xorByte := int8(xorBytes[0])
					xoredMessage[j] = byte(int(message[j]) ^ int(xorByte))
				}
				_ = xoredMessage
				prng.Clear()
			}
		})
	}
}

// BenchmarkXORStream_Bulk_NEW mirrors the post-refactor strategy: one
// allocation for the XOR buffer, one NextBytesInto fill, then a straight
// loop XORing message bytes. Directly compares against BenchmarkXORStream_
// PerByte_OLD at the same message sizes.
func BenchmarkXORStream_Bulk_NEW(b *testing.B) {
	sizes := []int{16, 256, 1024, 4096}
	for _, size := range sizes {
		b.Run(sizeLabel(size), func(b *testing.B) {
			message := make([]byte, size)
			factory := f5prng.NewDefaultFactory()

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				prng := factory.NewPRNG()
				if err := prng.Seed([]byte("bench-password")); err != nil {
					b.Fatal(err)
				}

				xoredMessage := make([]byte, len(message))
				if into, ok := prng.(f5prng.RandomSourceWithBytesInto); ok {
					if err := into.NextBytesInto(xoredMessage); err != nil {
						b.Fatal(err)
					}
				} else {
					copy(xoredMessage, prng.NextBytes(len(message)))
				}
				for j := range message {
					xorByte := int8(xoredMessage[j])
					xoredMessage[j] = byte(int(message[j]) ^ int(xorByte))
				}
				_ = xoredMessage
				prng.Clear()
			}
		})
	}
}

func sizeLabel(n int) string {
	// Tiny int-to-decimal helper so we don't pull strconv into the benchmark.
	if n == 0 {
		return "size=0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return "size=" + string(buf[i:])
}
