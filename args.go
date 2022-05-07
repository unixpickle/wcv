package main

import (
	"errors"
	"strings"
)

const UsageString = "usage: wcv [-clmw] [file ...]"

type Flags struct {
	Bytes bool
	Lines bool
	Words bool
	Chars bool
}

type Args struct {
	Flags Flags
	Paths []string
}

// ParseArgs parses command-line arguments or returns an error to print before
// exiting the program.
func ParseArgs(args []string) (*Args, error) {
	flags := Flags{Bytes: true, Lines: true, Words: true}
	inputFiles := []string{}

	firstFlag := true
	doneFlags := false
	for _, arg := range args {
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
					return nil, errors.New("wcv: illegal option -- " + string(option))
				}
			}
		} else {
			inputFiles = append(inputFiles, arg)
			doneFlags = true
		}
	}
	return &Args{Flags: flags, Paths: inputFiles}, nil
}
