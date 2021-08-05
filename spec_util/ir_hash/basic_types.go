package ir_hash

import (
	"bytes"
	"fmt"
	"math"
	"sort"

	"github.com/OneOfOne/xxhash"
)

// Precomputed hashes of small integer values, used for keys of elements
// within structs.
var intHashes [20][]byte = [20][]byte{
	{162, 114, 117, 172, 151, 133, 142, 24}, // 0
	{47, 84, 127, 235, 64, 101, 237, 100},   // 1
	{89, 194, 39, 250, 214, 7, 181, 11},
	{144, 61, 249, 201, 147, 47, 226, 15},
	{41, 81, 62, 160, 172, 176, 136, 243},
	{70, 0, 154, 221, 80, 195, 6, 43}, // 5
	{12, 80, 18, 243, 166, 19, 187, 70},
	{209, 133, 123, 24, 11, 194, 73, 121},
	{175, 111, 62, 127, 242, 49, 78, 64},
	{215, 156, 57, 162, 131, 74, 8, 3},
	{126, 36, 73, 85, 196, 10, 126, 61}, // 10
	{67, 253, 46, 121, 160, 140, 94, 191},
	{233, 222, 63, 217, 105, 109, 202, 168},
	{113, 14, 223, 191, 27, 211, 146, 205},
	{136, 161, 51, 107, 95, 211, 232, 37},

	{242, 136, 183, 32, 40, 8, 79, 220}, // 15
	{195, 21, 239, 13, 255, 220, 164, 186},
	{30, 67, 57, 202, 81, 216, 90, 144},
	{76, 184, 221, 131, 91, 228, 27, 127},
	{63, 69, 245, 47, 112, 175, 230, 25},
}

func Hash_Int64(i int64) []byte {
	hf := xxhash.New64()
	hf.Write([]byte(`u`))
	hf.Write([]byte(fmt.Sprintf("%d", i)))
	return hf.Sum(nil)
}

func Hash_Uint64(i uint64) []byte {
	hf := xxhash.New64()
	hf.Write([]byte(`i`))
	hf.Write([]byte(fmt.Sprintf("%d", i)))
	return hf.Sum(nil)
}

func Hash_Int32(i int32) []byte {
	hf := xxhash.New64()
	hf.Write([]byte(`i`))
	hf.Write([]byte(fmt.Sprintf("%d", i)))
	return hf.Sum(nil)
}

func Hash_Uint32(i uint32) []byte {
	hf := xxhash.New64()
	hf.Write([]byte(`i`))
	hf.Write([]byte(fmt.Sprintf("%d", i)))
	return hf.Sum(nil)
}

func Hash_Unicode(s string) []byte {
	hf := xxhash.New64()
	hf.Write([]byte(`u`))
	hf.Write([]byte(s))
	return hf.Sum(nil)
}

func Hash_Bool(b bool) []byte {
	hf := xxhash.New64()
	hf.Write([]byte(`b`))
	if b {
		hf.Write([]byte(`1`))
	} else {
		hf.Write([]byte(`0`))
	}
	return hf.Sum(nil)
}

func Hash_Bytes(b []byte) []byte {
	hf := xxhash.New64()
	hf.Write([]byte(`r`))
	hf.Write(b)
	return hf.Sum(nil)
}

type KeyValuePair struct {
	KeyHash   []byte
	ValueHash []byte
}

func Hash_KeyValues(pairs []KeyValuePair) []byte {
	hf := xxhash.New64()
	hf.Write([]byte(`d`))
	sort.Slice(pairs, func(i, j int) bool {
		return bytes.Compare(pairs[i].KeyHash, pairs[j].KeyHash) < 0
	})
	for _, p := range pairs {
		hf.Write(p.KeyHash)
		hf.Write(p.ValueHash)
	}
	return hf.Sum(nil)
}

func Hash_Float32(v float32) []byte {
	return Hash_Float64(float64(v))
}

func Hash_Float64(f float64) []byte {
	hf := xxhash.New64()
	hf.Write([]byte(`f`))
	switch {
	case math.IsInf(f, 1):
		hf.Write([]byte("Infinity"))
	case math.IsInf(f, -1):
		hf.Write([]byte("-Infinity"))
	case math.IsNaN(f):
		hf.Write([]byte("NaN"))
	default:
		normalizedFloat, _ := floatNormalize(f)
		hf.Write([]byte(normalizedFloat))
	}
	return hf.Sum(nil)
}

// Copied from normalization.go in objecthash-proto
// Copyright 2017 The ObjectHash-Proto Authors
func floatNormalize(originalFloat float64) (string, error) {
	// Special case 0
	// Note that if we allowed f to end up > .5 or == 0, we'd get the same thing.
	if originalFloat == 0 {
		return "+0:", nil
	}

	// Sign
	f := originalFloat
	s := `+`
	if f < 0 {
		s = `-`
		f = -f
	}
	// Exponent
	e := 0
	for f > 1 {
		f /= 2
		e++
	}
	for f <= .5 {
		f *= 2
		e--
	}
	s += fmt.Sprintf("%d:", e)
	// Mantissa
	if f > 1 || f <= .5 {
		return "", fmt.Errorf("Could not normalize float: %f", originalFloat)
	}
	for f != 0 {
		if f >= 1 {
			s += `1`
			f--
		} else {
			s += `0`
		}
		if f >= 1 {
			return "", fmt.Errorf("Could not normalize float: %f", originalFloat)
		}
		if len(s) >= 1000 {
			return "", fmt.Errorf("Could not normalize float: %f", originalFloat)
		}
		f *= 2
	}
	return s, nil
}
