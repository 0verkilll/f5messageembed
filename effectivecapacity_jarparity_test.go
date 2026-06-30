package f5messageembed

import "testing"

// jarExpected reproduces f5.jar's effective-capacity `_expected` directly from
// james/JpegEncoder.java:281-282 so the table below has an independent oracle
// (not a copy of the production code):
//
//	_large    = P - h(1)
//	_expected = _large + (int) (0.49 * _one)   // integer truncation
//
// Pure integer arithmetic on the truncated 0.49·h(1) term; no float rounding.
func jarExpected(p, h1 int) int {
	return (p - h1) + int(0.49*float64(h1))
}

// TestEffectiveCapacityMatchesJarExpected pins EffectiveCapacityCoefficients to
// f5.jar's `_expected` byte-for-byte across a range of h(1) parities. The
// pre-fix implementation used P − round(0.51·h(1)), which is +1 too large for
// every even h(1) — the divergence that flips k at a code-word bucket boundary
// and produces un-extractable high-Q stego.
func TestEffectiveCapacityMatchesJarExpected(t *testing.T) {
	t.Parallel()

	// Sweep P and both h(1) parities; high-Q covers push h(1) into the
	// thousands (Q100 on a 400x400 photo gives h(1)≈4723), so cover that range.
	for _, p := range []int{500, 1176, 9974, 14999, 15644, 19042, 21595, 60000} {
		for h1 := 0; h1 <= 8000 && h1 <= p; h1++ {
			got := EffectiveCapacityCoefficients(p, h1)
			want := jarExpected(p, h1)
			if want < 0 {
				want = 0
			}
			if got != want {
				t.Fatalf("EffectiveCapacityCoefficients(P=%d, h1=%d) = %d, want jar _expected = %d (delta %d)",
					p, h1, got, want, got-want)
			}
		}
	}
}

// TestEffectiveCapacityNotRoundForm guards specifically against a regression
// back to the old P − round(0.51·h(1)) form. The two forms diverge by exactly
// 1 for ~49% of h(1) values (whenever the fractional part of 0.51·h(1) is below
// 0.5, e.g. h(1)=2,4,...,4722,6624); that 1-coefficient gap is what flips k.
// Every h(1) listed here is a value where the old form was wrong.
func TestEffectiveCapacityNotRoundForm(t *testing.T) {
	t.Parallel()

	const p = 15644 // observed Q95 400x400 cover P
	for _, h1 := range []int{2, 4, 6, 8, 4722, 6624} {
		oldRoundForm := p - int(0.51*float64(h1)+0.5)
		jar := jarExpected(p, h1)
		if oldRoundForm == jar {
			t.Fatalf("test oracle bug: h1=%d should differ between forms", h1)
		}
		if got := EffectiveCapacityCoefficients(p, h1); got != jar {
			t.Fatalf("EffectiveCapacityCoefficients(P=%d, h1=%d) = %d, want %d (must NOT be old round form %d)",
				p, h1, got, jar, oldRoundForm)
		}
	}
}

// TestSelectOptimalKWithShrinkageJarKFlip demonstrates the user-visible
// consequence of the capacity bug: at these (P, h1, message) triples the
// pre-fix round(0.51·h1) `_expected` selected a k one higher than f5.jar,
// yielding byte-divergent stego. After the fix, the k matches f5.jar's choice
// (computed here via the independent jar-exact arithmetic).
//
// Triples were found by exhaustively sweeping message sizes for even h(1)
// covers (delta=1) until k differed; see the divergence search in the
// high-Q parity investigation.
func TestSelectOptimalKWithShrinkageJarKFlip(t *testing.T) {
	t.Parallel()

	cases := []struct {
		p, h1, msgBytes, wantK int
	}{
		{p: 737, h1: 2, msgBytes: 88, wantK: 0},
		{p: 1422, h1: 2, msgBytes: 72, wantK: 2},
		{p: 1970, h1: 2, msgBytes: 60, wantK: 3},
		{p: 3203, h1: 2, msgBytes: 167, wantK: 2},
		{p: 3340, h1: 2, msgBytes: 274, wantK: 1},
	}

	for _, c := range cases {
		// jarK: independent oracle running f5.jar's k-selection loop
		// (JpegEncoder.java:337-349) over the jar-exact `_expected`.
		jarK := jarSelectK(jarExpected(c.p, c.h1), c.msgBytes)
		if jarK != c.wantK {
			t.Fatalf("oracle disagreement: P=%d h1=%d msg=%d jarK=%d wantK=%d",
				c.p, c.h1, c.msgBytes, jarK, c.wantK)
		}

		// k=0 means the message does not fit; the production selector
		// returns an error in that case. Exercise both paths.
		gotK, err := SelectOptimalKWithShrinkage(c.p, c.h1, c.msgBytes*8)
		if c.wantK == 0 {
			if err == nil {
				t.Fatalf("SelectOptimalKWithShrinkage(P=%d h1=%d msg=%d) = k%d, want capacity error (jar picks k=0)",
					c.p, c.h1, c.msgBytes, gotK)
			}
			continue
		}
		if err != nil {
			t.Fatalf("SelectOptimalKWithShrinkage(P=%d h1=%d msg=%d) unexpected error: %v",
				c.p, c.h1, c.msgBytes, err)
		}
		if gotK != c.wantK {
			t.Fatalf("SelectOptimalKWithShrinkage(P=%d h1=%d msg=%d) = k%d, want jar k%d",
				c.p, c.h1, c.msgBytes, gotK, c.wantK)
		}
	}
}

// jarSelectK is an independent re-implementation of f5.jar's k-selection loop
// (james/JpegEncoder.java:337-349) for use as a test oracle. Returns k in
// [0,7]; 0 means no k fits the message.
func jarSelectK(expected, msgBytes int) int {
	i := 1
	for ; i < 8; i++ {
		n := (1 << i) - 1
		usable := expected*i/n - expected*i/n%n
		usable /= 8
		if usable == 0 {
			break
		}
		if usable < msgBytes+4 {
			break
		}
	}
	return i - 1
}
