# F5 — A Steganographic Algorithm

## High Capacity Despite Better Steganalysis

**Author:** Andreas Westfeld
**Institution:** Technische Universität Dresden, Institute for System Architecture
**Contact:** westfeld@inf.tu-dresden.de
**Published:** Information Hiding 2001, LNCS 2137, pp. 289–302
**Publisher:** Springer-Verlag Berlin Heidelberg 2001

---

## Abstract

Many steganographic systems are weak against visual and statistical attacks. Systems without these weaknesses offer only a relatively small capacity for steganographic messages. The newly developed algorithm F5 withstands visual and statistical attacks, yet it still offers a large steganographic capacity. F5 implements matrix encoding to improve the efficiency of embedding. Thus it reduces the number of necessary changes. F5 employs permutative straddling to uniformly spread out the changes over the whole steganogram.

---

## 1. Introduction

Secure steganographic algorithms hide confidential messages within other extensive data (carrier media). An attacker should not be able to find out that something is embedded in the steganogram (i.e., a steganographically modified carrier medium).

> **Note:** The steganographic techniques considered here are not intended for robust watermarking.

Visual attacks on steganographic systems are based on essential information in the carrier medium that steganographic algorithms overwrite. Adaptive techniques (that bring the embedding rate in line with the carrier content) prevent visual attacks, however, they also reduce the proportion of steganographic information in a carrier medium. Lossy compressed carrier media (JPEG, MP3, ...) are originally adaptive and immune against visual (and auditory respectively) attacks.

The steganographic tool **Jsteg** embeds messages in lossy compressed JPEG files. It has a high capacity—e.g., 12% of the steganogram's size—and is immune against visual attacks. However, a statistical attack discovers changes made by Jsteg.

**MP3Stego** and **IVS-Stego** also withstand auditory and visual attacks respectively. Apart from this, the extremely low embedding rate prevents all known statistical attacks. These two steganographic tools offer only a relatively small capacity for steganographic messages (less than 1% of the steganogram's size).

---

## 2. JPEG File Interchange Format

The file format defined by the Joint Photographic Experts Group (JPEG) stores image data in lossy compressed form as quantised frequency coefficients.

### JPEG Compression Pipeline

```
Bitmap image (BMP/PPM) → DCT → Quantisation → Huffman coding → JPEG image
```

**Process:**
1. The JPEG compressor cuts the uncompressed bitmap image into parts of 8×8 pixels
2. The **Discrete Cosine Transformation (DCT)** transfers 8×8 brightness values into 8×8 frequency coefficients (real numbers)
3. After DCT, the **quantisation** suitably rounds the frequency coefficients to integers in the range −2048...2047 (lossy step)
4. **Huffman coding** ensures redundancy-free coding of the quantised coefficients

### DCT Coefficient Distribution

The histogram of JPEG coefficients after quantisation shows two characteristic properties:

1. **Decreasing frequency:** The coefficient's frequency of occurrence decreases with increasing absolute value
2. **Decreasing rate of decrease:** The decrease of the coefficient's frequency of occurrence decreases with increasing absolute value (i.e., the difference between two bars of the histogram in the middle is larger than on the margin)

**Figure 2: Histogram for JPEG coefficients after quantisation:**
```
Frequency of occurrence
        ▲
 50,000 ┤              ████
 40,000 ┤              ████
 30,000 ┤          ████████████
 20,000 ┤      ████████████████████
 10,000 ┤  ████████████████████████████
        └┬────┬────┬────┬────┬────┬────┬────┬────┬─► JPEG coefficient
        -8   -6   -4   -2    0    2    4    6    8
```

These properties do not survive the Jsteg embedding process (see Section 3).

---

## 3. Jsteg

This algorithm made by Derek Upham serves as a starting point for the contemplation here, because it is resistant against the visual attacks and nevertheless offers an admirable capacity for steganographic messages (e.g., 12.8% of the steganogram's size).

### Embedding Mechanism

After quantisation, Jsteg replaces the **least significant bits (LSB)** of the frequency coefficients by the secret message. The embedding mechanism skips all coefficients with the values 0 or 1.

```c
short use_inject = 1;              /* set to 0 at end of message */

short inject(short inval)          /* inval is a JPEG coefficient */
{
    short inbit;
    if ((inval & 1) != inval)      /* don't embed in 0 or 1 */
        if (use_inject) {          /* still message bits to embed? */
            if ((inbit=bitgetbit()) != -1) { /* get next bit */
                inval &= ~1;       /* overwrite the lsb ... */
                inval |= inbit;    /* ... with this bit */
            } else
                use_inject = 0;    /* full message embedded */
        }
    return inval;                  /* return modified JPEG coefficient */
}
```

### Statistical Attack on Jsteg

The statistical attack on Jsteg reliably discovers the existence of embedded messages, because Jsteg replaces bits and thus introduces a dependency between the value's frequency of occurrence that only differ in the LSB.

Jsteg influences pairs of the coefficient's frequency of occurrence. For a modified image, the assumption is that adjacent frequencies c₂ᵢ and c₂ᵢ₊₁ are similar.

**Figure 5: Jsteg equalises pairs of coefficients:**
```
Frequency of occurrence
        ▲
 50,000 ┤              ████
 40,000 ┤              ████
 30,000 ┤          ████████████         ← Adjacent pairs become
 20,000 ┤      ████████████████████       equal height after
 10,000 ┤  ████████████████████████████   Jsteg embedding
        └┬────┬────┬────┬────┬────┬────┬────┬────┬─► JPEG coefficient
        -8   -6   -4   -2    0    2    4    6    8
               ╰─╯  ╰─╯  ╰─╯  ╰─╯  ╰─╯  ╰─╯
              Paired coefficients (2i, 2i+1) equalised
```

**Expected distribution:**
```
n*ᵢ = (c₂ᵢ + c₂ᵢ₊₁) / 2                                    (1)
```

**Observed distribution:**
```
nᵢ = c₂ᵢ                                                   (2)
```

**Chi-square test:**
```
χ² = Σᵢ₌₁ᵏ (nᵢ - n*ᵢ)² / n*ᵢ                                (3)
```

with k−1 degrees of freedom (number of different categories in the histogram minus one).

**Probability of embedding:**
```
p = 1 - (1 / (2^((k-1)/2) · Γ((k-1)/2))) ∫₀^χ² e^(-t/2) · t^((k-1)/2 - 1) dt   (4)
```

**Figure 6: Probability of embedding in a Jsteg steganogram (50% of capacity used):**
```
Probability of embedding
     ▲
 1.0 ┤████████████████████████████████████████████████████▄
     │                                                     ██
 0.8 ┤                                                       █
     │                                                        █
 0.6 ┤                                                         █
     │                                                          █
 0.4 ┤                                                          █
     │                                                           █
 0.2 ┤                                                            █
     │                                                             █
 0.0 ┤──────────────────────────────────────────────────────────────████
     └┬─────┬─────┬─────┬─────┬─────┬─────┬─────┬─────┬─────┬─────┬─► %
      0    10    20    30    40    50    60    70    80    90   100
                              Size of sample (%)

     p=1.00 up to 54% ──────────────╮
                                    ╰──► drops sharply at 56% (p=0.45)
                                         then p=0.00 at 59%+ as unchanged
                                         coefficients dilute the signal
```

---

## 4. F3

The algorithm F3 serves as a tutorial example. It differs in two respects from Jsteg:

### 4.1 Decrementing Instead of Overwriting

Instead of overwriting bits, F3 **decrements the coefficient's absolute values** in case their LSB does not match—except coefficients with the value zero, where we cannot decrement the absolute value.

- Zero coefficients are not used steganographically
- The LSB of nonzero coefficients matches the secret message after embedding
- The Chi-square test cannot easily detect such changes
- The symmetry of 1 and −1 visible in the histogram consequently remains

### 4.2 Shrinkage Problem

Some embedded bits fall victim to **shrinkage**. Shrinkage accrues every time F3 decrements the absolute value of 1 and −1 producing a 0.

- The receiver cannot distinguish a zero coefficient (steganographically unused) from a 0 produced by shrinkage
- The receiver skips all zero coefficients
- The sender repeatedly embeds the affected bit since he notices when he produces a 0

### F3's Statistical Weakness

The histogram shows a **relative surplus of even coefficients**. This phenomenon results from the repeated embedding after shrinkage.

- Shrinkage occurs only if we embed a zero bit
- The repetition of these zero bits shifts the ratio of steganographic values in favour of steganographic zeroes
- The F3 embedding process produces more even coefficients than odd

**Figure 7: F3 produces a superior number of even coefficients:**
```
Frequency of occurrence
        ▲
        │                    ████
 20,000 ┤              ██    ████
        │              ██  ████████
 15,000 ┤          ██  ████████████
        │          ██████████████████
 10,000 ┤      ████████████████████████
        │    ████████████████████████████
  5,000 ┤  ████████████████████████████████
        └┬────┬────┬────┬────┬────┬────┬────┬────┬─► JPEG coefficient
        -8   -6   -4   -2    0    2    4    6    8
                        │    │    │
                        └─┬──┴────┘
                          │
              Even coefficients (−2, 2, −4, 4, ...) are MORE frequent
              than odd coefficients — detectable statistical anomaly
```

---

## 5. F4

F3 has two weaknesses:

1. Because of the exclusive shrinkage of steganographic zeroes, F3 effectively embeds more zeroes than ones, producing statistically detectable peculiarities in the histogram
2. The histogram of JPEG files contains more odd than even coefficients (excluding 0). Therefore, unchanged carrier media contain (from Jsteg's or F3's perspective) more steganographic ones than zeroes

### F4's Solution: Sign-Based Encoding

The algorithm F4 eliminates these two weaknesses in one stroke by **mapping negative coefficients to the inverted steganographic value**:

| Coefficient | Steganographic Value |
|------------|---------------------|
| Even negative | 1 |
| Odd negative | 0 |
| Even positive | 0 |
| Odd positive | 1 |

**Figure 8: Histogram with F4's interpretation of steganographic values:**
```
Frequency of occurrence                              LEGEND:
        ▲                                            ▓▓ = stego 0
 50,000 ┤              ░░▓▓                          ░░ = stego 1
        │              ░░▓▓
 40,000 ┤              ░░▓▓
        │          ▓▓░░░░▓▓░░▓▓
 30,000 ┤          ▓▓░░░░▓▓░░▓▓
        │      ░░▓▓▓▓░░░░▓▓░░▓▓▓▓░░
 20,000 ┤      ░░▓▓▓▓░░░░▓▓░░▓▓▓▓░░
        │  ▓▓░░░░▓▓▓▓░░░░▓▓░░▓▓▓▓░░░░▓▓
 10,000 ┤  ▓▓░░░░▓▓▓▓░░░░▓▓░░▓▓▓▓░░░░▓▓
        └┬────┬────┬────┬────┬────┬────┬────┬────┬─► JPEG coefficient
        -8   -6   -4   -2    0    2    4    6    8
         │    │    │    │         │    │    │    │
         1    0    1    0    ─    0    1    0    1   ← Stego value
        ╰────────────────╯       ╰────────────────╯
         NEGATIVE: even=1        POSITIVE: even=0
                   odd=0                   odd=1
```

### F4 Embedding Code (Java)

```java
int nextBitToEmbed = embeddedData.readBit();
for (int i=0; i<coeff.length; i++) {
    if (i%64 == 0) continue;           // skip DC coefficients
    if (coeff[i] == 0) continue;       // skip zeroes
    if (coeff[i] > 0) {
        if ((coeff[i]&1) != nextBitToEmbed)
            coeff[i]--;                // decrease absolute value
    } else {
        if ((coeff[i]&1) == nextBitToEmbed)
            coeff[i]++;                // decrease absolute value
    }
    if (coeff[i] != 0) {               // successfully embedded
        if (embeddedData.available()==0)
            break;                     // end of embeddedData
        nextBitToEmbed = embeddedData.readBit();
    }
}
```

### Mathematical Proof of Characteristic Preservation

Suppose we have two random variables X, Y for observed coefficients before and after F4 embeds a message.

**Characteristic properties of X:**
```
P(X=1) > P(X=2) > P(X=3) > P(X=4)                          (5)
P(X=1) - P(X=2) > P(X=2) - P(X=3) > P(X=3) - P(X=4)        (6)
```

**If the message bits are uniformly distributed:**
```
P(Y=1) = ½P(X=1) + ½P(X=2)                                 (7)
P(Y=2) = ½P(X=2) + ½P(X=3)                                 (8)
P(Y=3) = ½P(X=3) + ½P(X=4)                                 (9)
```

**Subtracting equations:**
```
P(Y=1) - P(Y=2) = ½P(X=1) - ½P(X=3)                        (10)
P(Y=2) - P(Y=3) = ½P(X=2) - ½P(X=4)                        (11)
```

From (5), the right-hand sides of (10) and (11) are positive, giving the **first characteristic property for Y**:
```
P(Y=1) > P(Y=2) > P(Y=3)                                   (12)
```

From (5) and (6):
```
P(X=1) - P(X=3) > P(X=2) - P(X=4)                          (13)
```

Therefore, the **second characteristic property for Y**:
```
P(Y=1) - P(Y=2) > P(Y=2) - P(Y=3)                          (14)
```

---

## 6. F5

Unlike stream media (like in video conferences), image files only provide a limited steganographic capacity. In many cases, an embedded message does not require the full capacity. With continuous embedding, the changes concentrate on the start of the file, and the unused rest resides on the end.

To prevent attacks, the embedding function should use the carrier medium as regular as possible. **The embedding density should be the same everywhere.**

**Figure 10: Continuous embedding concentrates changes (×) at the start:**
```
   CARRIER IMAGE                    STATISTICAL ANALYSIS
┌────────────────────────────┐    ┌────────────────────────────┐
│ × × × × × · · · · · · · · │    │▓▓▓▓▓▓▓▓▓▓░░░░░░░░░░░░░░░░░│
│ × × × × · · · · · · · · · │    │▓▓▓▓▓▓▓▓░░░░░░░░░░░░░░░░░░░│
│ × × × · · · · · · · · · · │    │▓▓▓▓▓▓░░░░░░░░░░░░░░░░░░░░░│
│ × × · · · · · · · · · · · │    │▓▓▓▓░░░░░░░░░░░░░░░░░░░░░░░│
│ × · · · · · · · · · · · · │    │▓▓░░░░░░░░░░░░░░░░░░░░░░░░░│
│ · · · · · · · · · · · · · │    │░░░░░░░░░░░░░░░░░░░░░░░░░░░│
└────────────────────────────┘    └────────────────────────────┘
  × = changes concentrated          ▓▓ = detectable region
      at file start                 ░░ = clean region
```

**Figure 11: Permutative embedding scatters the changes (×) uniformly:**
```
   CARRIER IMAGE                    STATISTICAL ANALYSIS
┌────────────────────────────┐    ┌────────────────────────────┐
│ · × · · × · · · × · · × · │    │▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓│
│ · · · × · · × · · · × · · │    │▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓│
│ × · · · · × · · × · · · × │    │▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓│
│ · · × · · · · × · · × · · │    │▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓│
│ · × · · × · · · · × · · × │    │▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓│
│ × · · × · · × · × · · × · │    │▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓│
└────────────────────────────┘    └────────────────────────────┘
  × = changes uniformly             ▓▓ = uniform low-level signal
      scattered by permutation          (harder to detect)
```

### 6.1 Permutative Straddling

Some well-known steganographic algorithms scatter the message over the whole carrier medium. Many of them have bad time complexity—they get slower if we try to exhaust the steganographic capacity completely.

Straddling is easy if the capacity of the carrier medium is known exactly. However, we cannot predict the shrinkage for F4, because it depends on which bit is embedded at which position. We merely can estimate the expected capacity.

**F5's Straddling Mechanism:**
1. Shuffles all coefficients using a permutation first
2. Then F5 embeds into the permuted sequence
3. The shrinkage does not change the number of coefficients (only their values)
4. The permutation depends on a key derived from a password
5. F5 delivers the steganographically changed coefficients in its original sequence to the Huffman coder
6. With the correct key, the receiver is able to repeat the permutation

The permutation has **linear time complexity O(n)**.

### 6.2 Matrix Encoding

Ron Crandall introduced matrix encoding as a new technique to improve the embedding efficiency. **F5 possibly is the first implementation of matrix encoding.**

If most of the capacity is unused in a steganogram, matrix encoding decreases the necessary number of changes.

**Without matrix encoding:** embedding efficiency of 2 bits per change
**With F4 shrinkage:** embedding efficiency is about 1.5 bits per change

#### Example: (1, 3, 2) Code

We want to embed two bits x₁, x₂ in three modifiable bit places a₁, a₂, a₃ changing one place at most:

| Condition | Action |
|-----------|--------|
| x₁ = a₁ ⊕ a₃, x₂ = a₂ ⊕ a₃ | change nothing |
| x₁ ≠ a₁ ⊕ a₃, x₂ = a₂ ⊕ a₃ | change a₁ |
| x₁ = a₁ ⊕ a₃, x₂ ≠ a₂ ⊕ a₃ | change a₂ |
| x₁ ≠ a₁ ⊕ a₃, x₂ ≠ a₂ ⊕ a₃ | change a₃ |

In all four cases we do not change more than one bit.

#### General Matrix Encoding

We have a code word **a** with n modifiable bit places for k secret message bits **x**. Let f be a hash function that extracts k bits from a code word.

Matrix encoding enables us to find a suitable modified code word **a'** for every **a** and **x** with **x** = f(**a'**), such that the Hamming distance:

```
d(a, a') ≤ d_max                                           (15)
```

We denote this code by an ordered triple (d_max, n, k): a code word with n places will be changed in not more than d_max places to embed k bits.

F5 implements matrix encoding only for **d_max = 1**. For (1, n, k), the code words have the length:
```
n = 2^k - 1
```

#### Change Density and Embedding Rate

**Change density (neglecting shrinkage):**
```
D(k) = 1/(n+1) = 1/2^k                                     (16)
```

**Embedding rate:**
```
R(k) = k/n = (1/n) · ld(n+1) = k/(2^k - 1)                 (17)
```

**Embedding efficiency** (average number of bits embedded per change):
```
W(k) = R(k)/D(k) = (2^k / (2^k - 1)) · k                   (18)
```

#### Table 1: Connection Between Change Density and Embedding Rate

| k | n | Change Density | Embedding Rate | Embedding Efficiency |
|---|---|----------------|----------------|---------------------|
| 1 | 1 | 50.00% | 100.00% | 2.00 |
| 2 | 3 | 25.00% | 66.67% | 2.67 |
| 3 | 7 | 12.50% | 42.86% | 3.43 |
| 4 | 15 | 6.25% | 26.67% | 4.27 |
| 5 | 31 | 3.12% | 16.13% | 5.16 |
| 6 | 63 | 1.56% | 9.52% | 6.09 |
| 7 | 127 | 0.78% | 5.51% | 7.06 |
| 8 | 255 | 0.39% | 3.14% | 8.03 |
| 9 | 511 | 0.20% | 1.76% | 9.02 |

The embedding efficiency of the (1, n, k) code is always larger than k. The rate decreases with increasing efficiency—**high efficiency can only be achieved with very short messages**.

#### Hash Function

Table 2 gives the dependencies between the message bits xᵢ and the changed bit places a'ⱼ. We assign the dependencies with the "binary coding" of j to column a'ⱼ.

**Table 2: Dependency (×) between message bits xᵢ and code word bits a'ⱼ**

**For (1, 3, 2) code:**
```
         │ a'₁ │ a'₂ │ a'₃ │
─────────┼─────┼─────┼─────┤
   x₁    │  ×  │     │  ×  │   (binary: 1, 3 → bits 1 and 3)
─────────┼─────┼─────┼─────┤
   x₂    │     │  ×  │  ×  │   (binary: 2, 3 → bits 2 and 3)
─────────┴─────┴─────┴─────┘
```

**For (1, 7, 3) code:**
```
         │ a'₁ │ a'₂ │ a'₃ │ a'₄ │ a'₅ │ a'₆ │ a'₇ │
─────────┼─────┼─────┼─────┼─────┼─────┼─────┼─────┤
   x₁    │  ×  │     │  ×  │     │  ×  │     │  ×  │  (odd positions)
─────────┼─────┼─────┼─────┼─────┼─────┼─────┼─────┤
   x₂    │     │  ×  │  ×  │     │     │  ×  │  ×  │  (2,3,6,7)
─────────┼─────┼─────┼─────┼─────┼─────┼─────┼─────┤
   x₃    │     │     │     │  ×  │  ×  │  ×  │  ×  │  (4,5,6,7)
─────────┴─────┴─────┴─────┴─────┴─────┴─────┴─────┘
```

The hash function can be determined very fast:
```
f(a) = ⊕ᵢ₌₁ⁿ aᵢ · i                                        (19)
```

We find the bit place to change:
```
s = x ⊕ f(a)                                               (20)
```

The changed code word results in:
```
a' = { a,                           if s = 0 (⟺ x = f(a))
     { (a₁, a₂, ..., ¬aₛ, ..., aₙ)  otherwise              (21)
```

#### Optimal Parameter Selection

We can find an optimal parameter k for every message to embed and every carrier medium providing sufficient capacity, so that the message just fits into the carrier medium.

**Example:** If we want to embed a message with 1000 bits into a carrier medium with a capacity of 50000 bits:
- Required embedding rate: R = 1000 ÷ 50000 = 2%
- This value is between R(k=8) and R(k=9) in Table 1
- Choose k = 8, able to embed 50000 ÷ 255 = 196 code words with n = 255
- The (1, 255, 8) code could embed 196 × 8 = 1568 bits
- If we chose k = 9 instead, we could not embed the message completely

### 6.3 Preserving Characteristic Properties

To prove the security of a steganographic algorithm, it would be necessary to formalise perceptibility. Instead, let us prove the resistance against known attacks.

The statistical attacks can reveal the presence of a hidden message if the steganographic algorithm overwrites least significant bits. This is no longer the case with F4/F5. F4 preserves characteristic properties and does not equalise frequencies.

**We can show that F5 preserves the same characteristic properties:**

Let 0 ≤ α ≤ 1 be the fraction of coefficients used for steganography (F4 is the special case α = 1).

```
P(Y=1) = (1 - α/2)P(X=1) + (α/2)P(X=2)                     (22)
P(Y=2) = (1 - α/2)P(X=2) + (α/2)P(X=3)                     (23)
P(Y=3) = (1 - α/2)P(X=3) + (α/2)P(X=4)                     (24)
```

Subtracting:
```
P(Y=1) - P(Y=2) = (1 - α/2)(P(X=1) - P(X=2)) + (α/2)(P(X=2) - P(X=3))   (25)
P(Y=2) - P(Y=3) = (1 - α/2)(P(X=2) - P(X=3)) + (α/2)(P(X=3) - P(X=4))   (26)
```

From (5), the right-hand sides of (25) and (26) are positive, giving **first characteristic property for Y**:
```
P(Y=1) > P(Y=2) > P(Y=3)                                   (27)
```

The right-hand side of (25) is greater than in (26), giving **second characteristic property for Y**:
```
P(Y=1) - P(Y=2) > P(Y=2) - P(Y=3)                          (28)
```

### 6.4 Implementation

The algorithm F5 has the following coarse structure:

1. **Start JPEG compression.** Stop after the quantisation of coefficients.

2. **Initialise** a cryptographically strong random number generator with the key derived from the password.

3. **Instantiate a permutation** (two parameters: random generator and number of coefficients including zero coefficients).

4. **Determine the parameter k** from the capacity of the carrier medium and the length of the secret message.

5. **Calculate the code word length** n = 2^k − 1.

6. **Embed the secret message** with (1, n, k) matrix encoding:
   - (a) Fill a buffer with n nonzero coefficients
   - (b) Hash this buffer (generate a hash value with k bit-places) (cf. (19))
   - (c) Add the next k bits of the message to the hash value (bit by bit, xor) (cf. (20))
   - (d) If the sum is 0, the buffer is left unchanged. Otherwise the sum is the buffer's index 1...n, the absolute value of its element has to be decremented. (cf. (21))
   - (e) **Test for shrinkage:** whether we produced a zero. If so, adjust the buffer (eliminate the 0 by reading one more nonzero coefficient, repeat step 6a beginning from the same coefficient). If no shrinkage occurred, advance to new coefficients behind the actual buffer. If there is still message data continue with step 6a.

7. **Continue JPEG compression** (Huffman coding etc.).

---

## 7. Conclusion

Many steganographic algorithms offer a high capacity for hidden messages but are weak against visual and statistical attacks. Tools withstanding these attacks provide only a very small capacity. The algorithm F4 combines both preferences: resistance against visual and statistical attacks as well as high capacity.

**Matrix encoding** and **permutative straddling** enable the user to:
- Decrease the necessary number of steganographic changes
- Equalise the embedding rate in the steganogram

**F5 accomplishes a steganographic proportion that exceeds 13% of the JPEG file size** (cf. Table 3). On the other hand, F5 is able to decrease the embedding rate arbitrarily.

**Acknowledgements.** I would like to thank Fabien Petitcolas for helpful comments.

### Table 3: Comparison of Several JPEG Files Created with F5

| File name | File size (bytes) | Embedded size (bytes) | Ratio embedded to steganogram size | Embedding efficiency | Quantiser quality |
|-----------|------------------|----------------------|-----------------------------------|---------------------|-------------------|
| expo.bmp | 1,562,030 | 0 (carrier medium) | — | — | — |
| expo80.jpg | 129,879 | 0 | — | — | 80% |
| ministeg.jpg | 129,760 | 213 | 0.2% | 3.8 | 80% |
| maxisteg.jpg | 115,685 | 15,480 | 13.4% | 1.5 | 80% |
| expo75.jpg | 114,712 | 0 | — | — | 75% |

---

## References

1. Ron Crandall: Some Notes on Steganography. Posted on Steganography Mailing List, 1998. http://os.inf.tu-dresden.de/~westfeld/crandall.pdf

2. Andy C. Hung: PVRG-JPEG Codec 1.1, Stanford University, 1993. http://archiv.leo.org/pub/comp/os/unix/graphics/jpeg/PVRG

3. Fabien Petitcolas: MP3Stego, 1998. http://www.cl.cam.ac.uk/~fapp2/steganography/mp3stego

4. Derek Upham: Jsteg, 1997. http://www.tiac.net/users/korejwa/jsteg.htm

5. Andreas Westfeld, Andreas Pfitzmann: Attacks on Steganographic Systems, in Andreas Pfitzmann (Ed.): Information Hiding. Third International Workshop, LNCS 1768, Springer-Verlag Berlin Heidelberg 2000. pp. 61–76.

6. Andreas Westfeld, Gritta Wolf: Steganography in a Video Conferencing System, in David Aucsmith (Ed.): Information Hiding, LNCS 1525, Springer-Verlag Berlin Heidelberg 1998. pp. 32–47.

7. Andreas Westfeld: The Steganographic Algorithm F5, 1999. http://wwwrn.inf.tu-dresden.de/~westfeld/f5.html

8. Jan Zöllner, Hannes Federrath, Herbert Klimant, Andreas Pfitzmann, Rudi Piotraschke, Andreas Westfeld, Guntram Wicke, Gritta Wolf: Modeling the Security of Steganographic Systems, in David Aucsmith (Ed.): Information Hiding, LNCS 1525, Springer-Verlag Berlin Heidelberg 1998. pp. 344–354.

---

## Key Equations Summary

| Equation | Description |
|----------|-------------|
| `n*ᵢ = (c₂ᵢ + c₂ᵢ₊₁) / 2` | Expected distribution for Jsteg attack |
| `χ² = Σ(nᵢ - n*ᵢ)² / n*ᵢ` | Chi-square statistic |
| `n = 2^k - 1` | Code word length for (1,n,k) matrix encoding |
| `D(k) = 1/2^k` | Change density |
| `R(k) = k/(2^k - 1)` | Embedding rate |
| `W(k) = (2^k/(2^k - 1)) · k` | Embedding efficiency |
| `f(a) = ⊕ᵢ aᵢ · i` | Hash function |
| `s = x ⊕ f(a)` | Bit position to change |

---

## Algorithm Evolution Summary

| Algorithm | LSB Handling | Uses Value 1 | Statistical Detection | Key Innovation |
|-----------|-------------|--------------|----------------------|----------------|
| **Jsteg** | Overwrite | No | Easily detected (chi-square) | High capacity |
| **F3** | Decrement | Yes | Detectable (even surplus) | No LSB overwrite |
| **F4** | Decrement | Yes | Resistant | Sign-based encoding |
| **F5** | Decrement | Yes | Resistant | Matrix encoding + permutative straddling |

---

*This document is a markdown transcription of the original paper for reference and documentation purposes.*
