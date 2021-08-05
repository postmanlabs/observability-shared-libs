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
var intHashes [20][]byte = [20][]byte{}

func init() {
	for i := range intHashes {
		intHashes[i] = Hash_Int64(int64(i))
	}
}

func Hash_Int64(i int64) []byte {
	hf := xxhash.New64()
	hf.Write([]byte(`i`))
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

// This function copied directly from normalization.go in objecthash-proto
//
// Copyright 2017 The ObjectHash-Proto Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
