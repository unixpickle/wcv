package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sync/atomic"
	"unicode"
)

type AtomicValue struct {
	Value int64
}

// Counts manages a collection of mutable, atomic counters.
//
// While each counter is itself atomic, the overall state is not.
// For example, the output of Format() might indicate zero characters but one
// line (an impossible scenario) if the Format() call is performed while a
// newline byte is being processed by Update().
type Counts struct {
	// Ensure alignment by packing all values in nested structs.
	Bytes AtomicValue
	Lines AtomicValue
	Words AtomicValue
	Chars AtomicValue
}

// Add adds other counts to c, using atomic loads and stores per value.
func (c *Counts) Add(other *Counts) {
	atomic.AddInt64(&c.Bytes.Value, atomic.LoadInt64(&other.Bytes.Value))
	atomic.AddInt64(&c.Lines.Value, atomic.LoadInt64(&other.Lines.Value))
	atomic.AddInt64(&c.Words.Value, atomic.LoadInt64(&other.Words.Value))
	atomic.AddInt64(&c.Chars.Value, atomic.LoadInt64(&other.Chars.Value))
}

// Format formats the counts as a line of output, accessing each count
// atomically.
func (c *Counts) Format(f Flags) string {
	mask := []bool{f.Lines, f.Words, f.Bytes, f.Chars}
	ptrs := []*int64{&c.Lines.Value, &c.Words.Value, &c.Bytes.Value, &c.Chars.Value}
	output := ""
	for i, ptr := range ptrs {
		if !mask[i] {
			continue
		}
		value := atomic.LoadInt64(ptr)
		output += fmt.Sprintf(" %7d", value)
	}
	return output
}

// Update adds to the counts with the contents of r.
func (c *Counts) Update(r io.Reader) error {
	br := bufio.NewReader(r)
	onNewWord := true
	for {
		ch, size, err := br.ReadRune()
		if size > 0 {
			atomic.AddInt64(&c.Bytes.Value, int64(size))
			atomic.AddInt64(&c.Chars.Value, 1)
			if ch == '\n' {
				atomic.AddInt64(&c.Lines.Value, 1)
			}

			if unicode.IsSpace(ch) {
				onNewWord = true
			} else if onNewWord {
				onNewWord = false
				atomic.AddInt64(&c.Words.Value, 1)
			}
		}

		if err != nil && errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}
	}
	return nil
}
