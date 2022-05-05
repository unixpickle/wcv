package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"
)

type Flags struct {
	Bytes bool
	Lines bool
	Words bool
	Chars bool
}

type Counts struct {
	// Ensure alignment by packing all values in nested structs.
	Bytes struct {
		Value int64
	}
	Lines struct {
		Value int64
	}
	Words struct {
		Value int64
	}
	Chars struct {
		Value int64
	}
}

func (c *Counts) Add(other *Counts) {
	c.Bytes.Value += other.Bytes.Value
	c.Lines.Value += other.Lines.Value
	c.Words.Value += other.Words.Value
	c.Chars.Value += other.Chars.Value
}

func main() {
	flags := Flags{Bytes: true, Lines: true, Words: true}
	inputFiles := []string{}

	firstFlag := true
	doneFlags := false
	for _, arg := range os.Args[1:] {
		if !doneFlags && strings.HasPrefix(arg, "-") {
			if arg == "--" {
				doneFlags = true
				continue
			}
			for _, option := range arg[1:] {
				if firstFlag {
					// Override default behavior when argument is passed.
					flags = Flags{}
					firstFlag = false
				}
				if option == 'c' {
					flags.Bytes = true
					flags.Chars = false
				} else if option == 'm' {
					flags.Chars = true
					flags.Bytes = false
				} else if option == 'w' {
					flags.Words = true
				} else if option == 'l' {
					flags.Lines = true
				} else {
					fmt.Fprintln(os.Stderr, "wcv: illegal option -- "+string(option))
					fmt.Fprintln(os.Stderr, "usage: wc [-clmw] [file ...]")
					os.Exit(1)
				}
			}
		} else {
			inputFiles = append(inputFiles, arg)
			doneFlags = true
		}
	}

	if len(inputFiles) == 0 {
		PrintLive(flags, os.Stdin, "")
	} else {
		total := Counts{}
		for _, name := range inputFiles {
			if stat, err := os.Stat(name); err != nil {
				fmt.Fprintln(os.Stderr, "wcv: "+name+": "+err.Error())
				continue
			} else if stat.IsDir() {
				fmt.Fprintln(os.Stderr, "wcv: "+name+": Is a directory")
				continue
			}
			f, err := os.Open(name)
			if err != nil {
				fmt.Fprintln(os.Stderr, "wcv: "+name+": "+err.Error())
				continue
			}
			total.Add(PrintLive(flags, f, name))
			f.Close()
		}
		if len(inputFiles) > 1 {
			PrintCounts(flags, &total, "total", "\n")
		}
	}
}

func PrintLive(f Flags, r io.Reader, name string) *Counts {
	var printLock sync.Mutex
	doneCh := make(chan struct{})

	counts := Counts{}

	go func() {
		ticker := time.NewTicker(time.Second / 2)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
			case <-doneCh:
				return
			}
			printLock.Lock()
			select {
			case <-doneCh:
				return
			default:
			}
			PrintCounts(f, &counts, name, "\r")
			printLock.Unlock()
		}
	}()

	err := ProcessCounts(r, &counts)
	printLock.Lock()
	close(doneCh)
	PrintCounts(f, &counts, name, "\n")
	printLock.Unlock()
	if err != nil {
		fmt.Fprintln(os.Stderr, "wcv: "+name+": "+err.Error())
	}

	return &counts
}

func PrintCounts(f Flags, c *Counts, name, newlineChar string) {
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
	if name != "" {
		output += fmt.Sprintf(" %s", name)
	}
	output += newlineChar
	fmt.Print(output)
}

func ProcessCounts(r io.Reader, c *Counts) error {
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
