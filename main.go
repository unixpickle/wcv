package main

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

func main() {
	args, err := ParseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		fmt.Fprintln(os.Stderr, UsageString)
		os.Exit(1)
	}

	if len(args.Paths) == 0 {
		PrintLive(args.Flags, os.Stdin, "")
	} else {
		total := Counts{}
		for _, name := range args.Paths {
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
			total.Add(PrintLive(args.Flags, f, name))
			f.Close()
		}
		if len(args.Paths) > 1 {
			PrintCounts(args.Flags, &total, "total", "", true)
		}
	}
}

func PrintLive(f Flags, r io.Reader, name string) *Counts {
	var printLock sync.Mutex
	doneCh := make(chan struct{})

	counts := Counts{}

	go func() {
		ticker := time.NewTicker(time.Second / 4)
		defer ticker.Stop()
		firstTick := make(chan struct{}, 1)
		firstTick <- struct{}{}
		previous := ""
		for {
			select {
			case <-firstTick:
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
			previous = PrintCounts(f, &counts, name, previous, false)
			printLock.Unlock()
		}
	}()

	err := counts.Update(r)
	printLock.Lock()
	close(doneCh)
	PrintCounts(f, &counts, name, "", true)
	printLock.Unlock()
	if err != nil {
		fmt.Fprintln(os.Stderr, "wcv: "+name+": "+err.Error())
	}

	return &counts
}

func PrintCounts(f Flags, c *Counts, name, previous string, final bool) string {
	output := c.Format(f)
	if name != "" {
		output += fmt.Sprintf(" %s", name)
	}
	if output != previous {
		if final {
			fmt.Println(output)
		} else {
			fmt.Fprint(os.Stderr, output+"\r")
		}
	}
	return output
}
