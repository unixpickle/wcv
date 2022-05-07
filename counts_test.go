package main

import (
	"bytes"
	"testing"
)

func TestCountsUpdate(t *testing.T) {
	testData := []struct {
		Text    string
		Results Counts
	}{
		{"hello world", createCount(0, 2, 11, 11)},
		{"hello world\ntesting", createCount(1, 3, 19, 19)},
		{"hello world\ntesting\n\nhi   hi\n", createCount(4, 5, 29, 29)},
		{"😊 hey hi", createCount(0, 3, 8, 11)},
	}
	for i, data := range testData {
		r := bytes.NewReader([]byte(data.Text))
		actual := Counts{}
		if err := actual.Update(r); err != nil {
			t.Fatal(err)
		}
		if actual != data.Results {
			t.Errorf("case %d: expected %v but got %v", i, data.Results, actual)
		}
	}
}

func createCount(lines, words, chars, bytes int64) Counts {
	return Counts{
		Bytes: AtomicValue{Value: bytes},
		Lines: AtomicValue{Value: lines},
		Words: AtomicValue{Value: words},
		Chars: AtomicValue{Value: chars},
	}
}
