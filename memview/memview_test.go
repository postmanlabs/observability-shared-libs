package memview

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var errWriterErr = fmt.Errorf("errWriter: you've requested an error")

// Returns an error on the ith write.
type errWriter struct {
	targetCount int
	writeCount  int
}

func (w *errWriter) Write(data []byte) (int, error) {
	w.writeCount += 1
	if w.writeCount == w.targetCount {
		return 0, errWriterErr
	}
	return len(data), nil
}

func TestAppend(t *testing.T) {
	var mv MemView
	mv.Append(New([]byte("hello ")))
	mv.Append(New([]byte("prince!")))
	if mv.String() != "hello prince!" {
		t.Errorf(`expected "hello prince!" got "%s"`, mv.String())
	} else if mv.Len() != int64(len("hello prince!")) {
		t.Errorf(`expected new length %d, got %d`, len("hello prince!"), mv.Len())
	}
}

// DeepCopy MemViews should operate independently.
func TestDeepCopy(t *testing.T) {
	mv1 := New([]byte("hello"))
	mv2 := mv1.DeepCopy()
	mv2.Append(New([]byte(" prince!")))
	mv1.Append(New([]byte(" pineapple!")))

	if mv1.String() != "hello pineapple!" {
		t.Errorf(`expected "hello pineapple@" got "%s"`, mv1.String())
	} else if mv1.Len() != int64(len("hello pineapple!")) {
		t.Errorf(`expected length %d, got %d`, len("hello pineapple!"), mv1.Len())
	}

	if mv2.String() != "hello prince!" {
		t.Errorf(`expected "hello prince!" got "%s"`, mv2.String())
	} else if mv2.Len() != int64(len("hello prince!")) {
		t.Errorf(`expected length %d, got %d`, len("hello prince!"), mv2.Len())
	}
}

func TestReaderReflectChange(t *testing.T) {
	mv := New([]byte("hello"))
	r := mv.CreateReader()
	// Appends to mv should reflect in reader.
	mv.Append(New([]byte(" prince!")))

	actual, err := ioutil.ReadAll(r)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if string(actual) != "hello prince!" {
		t.Errorf(`expected "hello prince!" got "%s"`, string(actual))
	}
}

func TestReader(t *testing.T) {
	mv := New([]byte("hello"))
	mv.Append(New([]byte(" prince!")))

	// Test with every possible buffer size, including oversized ones.
	for bufSize := 1; bufSize < len("hello prince!")+10; bufSize++ {
		r := mv.CreateReader()
		buf := make([]byte, bufSize)
		read := []byte{}
		for {
			n, err := r.Read(buf)
			read = append(read, buf[:n]...)
			if err == io.EOF {
				break
			}
		}

		if diff := cmp.Diff(string(read), "hello prince!"); diff != "" {
			t.Errorf("found diff with bufSize=%d: %s", bufSize, diff)
		}
	}

}

func TestWriteTo(t *testing.T) {
	mv := New([]byte("hello"))
	mv.Append(New([]byte(" prince!")))

	var buf bytes.Buffer
	n, err := mv.CreateReader().WriteTo(&buf)
	if err != nil {
		t.Errorf("expected error: %v", err)
	} else if n != int64(len("hello prince!")) {
		t.Errorf("expected to write %d bytes, got %d", len("hello prince!"), n)
	} else if diff := cmp.Diff("hello prince!", string(buf.Bytes())); diff != "" {
		t.Errorf("found diff: %s", diff)
	}
}

func TestWriteToWithError(t *testing.T) {
	mv := New([]byte("hello"))
	mv.Append(New([]byte(" prince!")))

	// Return error on 2nd write, WriteTo should return bytes consumed from first
	// write and the error.
	w := &errWriter{targetCount: 2}
	n, err := mv.CreateReader().WriteTo(w)
	if err != errWriterErr {
		t.Errorf("expected errWriter error, got %v", err)
	} else if n != int64(len("hello")) {
		t.Errorf("expected to write %d bytes before error, got %d", len("hello"), n)
	}
}

func TestGetByte(t *testing.T) {
	input := "prince is a good boy"
	var mv MemView
	mv.Append(New([]byte("prince ")))
	mv.Append(New([]byte("is a ")))
	mv.Append(New([]byte("good ")))
	mv.Append(New([]byte("boy")))

	for i := 0; i < len(input); i++ {
		if b := mv.GetByte(int64(i)); b != input[i] {
			t.Errorf(`GetByte(%d) expected %s, got %s`, i, strconv.Quote(string(input[i])), strconv.Quote(string(b)))
		}
	}
}

func TestGetByteOutOfBounds(t *testing.T) {
	input := "prince is a good boy"
	var mv MemView
	mv.Append(New([]byte("prince ")))
	mv.Append(New([]byte("is a ")))
	mv.Append(New([]byte("good ")))
	mv.Append(New([]byte("boy")))

	inputs := []int64{-1, 10000, int64(len(input) + 1)}
	for _, i := range inputs {
		if b := mv.GetByte(i); b != 0 {
			t.Errorf("index=%d expected 0, got %d", i, b)
		}
	}
}

func TestSubView(t *testing.T) {
	input := "prince is a good boy"
	var mv MemView
	mv.Append(New([]byte("prince ")))
	mv.Append(New([]byte("is a ")))
	mv.Append(New([]byte("good ")))
	mv.Append(New([]byte("boy")))

	for i := 0; i < len(input); i++ {
		for j := i; j < len(input)+1; j++ {
			actual := mv.SubView(int64(i), int64(j))
			if diff := cmp.Diff(input[i:j], actual.String()); diff != "" {
				t.Errorf("found diff start=%d end=%d diff=%s", i, j, diff)
			} else if int64(len(input[i:j])) != actual.Len() {
				t.Errorf("subview length is wrong, expected=%d, got=%d", len(input[i:j]), actual.Len())
			}
		}
	}
}

func TestIndex(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		pattern  string
		start    int64
		expected int64
	}{
		{
			name:     "pattern only",
			input:    "<pattern>",
			pattern:  "<pattern>",
			start:    0,
			expected: 0,
		},
		{
			name:     "pattern with other data in front",
			input:    "ab <pattern>",
			pattern:  "<pattern>",
			start:    0,
			expected: 3,
		},
		{
			name:     "find pattern with start offset",
			input:    "<pattern> abc <pattern>",
			pattern:  "<pattern>",
			start:    1,
			expected: 14,
		},
		{
			name:     "pattern not in input",
			input:    "<pattern> abc <pattern>",
			pattern:  "<foobar>",
			start:    0,
			expected: -1,
		},
		{
			name:     "pattern not in input - nonzero start",
			input:    "<pattern> abc <pattern>",
			pattern:  "<foobar>",
			start:    7,
			expected: -1,
		},
		{
			name:     "find empty - zero start",
			input:    "<pattern> abc <pattern>",
			pattern:  "",
			start:    0,
			expected: 0,
		},
		{
			name:     "find empty - nonzero start",
			input:    "<pattern> abc <pattern>",
			pattern:  "",
			start:    7,
			expected: 7,
		},
		{
			name:     "find empty with empty input",
			input:    "",
			pattern:  "",
			start:    0,
			expected: 0,
		},
		{
			name:     "start offset > len with empty pattern",
			input:    "<pattern> abc <pattern>",
			pattern:  "",
			start:    int64(len("<pattern> abc <pattern>") + 100),
			expected: -1,
		},
		{
			name:     "start offset == len",
			input:    "<pattern> abc <pattern>",
			pattern:  "<pattern>",
			start:    int64(len("<pattern> abc <pattern>")),
			expected: -1,
		},
		{
			name:     "start offset > len",
			input:    "<pattern> abc <pattern>",
			pattern:  "<pattern>",
			start:    int64(len("<pattern> abc <pattern>") + 100),
			expected: -1,
		},
		/*
			{
				name:     "partial match",
				input:    "xxxxxyy",
				pattern:  "xxxyy",
				start:    0,
				expected: 2,
			},
		*/
	}

	for _, c := range testCases {
		// Try all possible ways of segmenting the input into 4 pieces.
		for i := 0; i < len(c.input); i++ {
			for j := i; j < len(c.input); j++ {
				for k := j; k < len(c.input); k++ {
					mv1 := New([]byte(c.input[:i]))
					mv2 := New([]byte(c.input[i:j]))
					mv3 := New([]byte(c.input[j:k]))
					mv4 := New([]byte(c.input[k:]))

					var mv MemView
					mv.Append(mv1)
					mv.Append(mv2)
					mv.Append(mv3)
					mv.Append(mv4)

					i := mv.Index(c.start, []byte(c.pattern))
					if i != c.expected {
						t.Errorf("[%s] expected %d, got %d, MemViews: %v", c.name, c.expected, i, []string{
							strconv.Quote(mv1.String()),
							strconv.Quote(mv2.String()),
							strconv.Quote(mv3.String()),
							strconv.Quote(mv4.String()),
						})
					}
				}
			}
		}
	}
}

func BenchmarkIndexSmall(b *testing.B) {
	letterBytes := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
	bytes1 := make([]byte, 1400)
	bytes2 := make([]byte, 1400)
	for i := range bytes1 {
		bytes1[i] = letterBytes[rand.Intn(len(letterBytes))]
		bytes2[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	view := New(bytes1)
	view.Append(New(bytes2))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		view.Index(0, []byte("POST"))
		view.Index(0, []byte("GET"))
		view.Index(0, []byte("DELETE"))
		view.Index(0, []byte("PUT"))
		view.Index(0, []byte("OPTION"))
	}
}

func BenchmarkIndexLarge(b *testing.B) {
	letterBytes := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
	view := New([]byte("xxxxxx"))
	for i := 0; i < 1000; i++ {
		bytes1 := make([]byte, 1400)
		for j := range bytes1 {
			bytes1[j] = letterBytes[rand.Intn(len(letterBytes))]
		}
		view.Append(New(bytes1))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		view.Index(0, []byte("POST"))
		view.Index(0, []byte("GET"))
		view.Index(0, []byte("DELETE"))
		view.Index(0, []byte("PUT"))
		view.Index(0, []byte("OPTION"))
	}
}
