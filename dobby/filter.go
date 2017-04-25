package main

import (
	"regexp"
	"strings"
)

// Filter - object that predict log line for watcher.
type Filter struct {
	reg  string
	line string
	err  *Error
}

// Error - general structure for store error parsed form log line.
type Error struct {
	Time    string
	Type    string
	Message string
	Color   string
}

// TODO: Improve parser.
func (f *Filter) parse() {
	f.err = &Error{}
	re := regexp.MustCompile(f.reg)
	match := re.FindStringSubmatch(f.line)
	names := re.SubexpNames()

	for i, entry := range match {
		if i == 0 {
			continue
		}
		switch names[i] {
		case "Time":
			f.err.Time = entry
		case "Message":
			f.err.Message = entry
		}
	}
}

func (f *Filter) divide() {
	if f.isDeprecated(f.err.Message) {
		f.err.Type = "Deprecated"
		f.err.Color = GetColor(ColorYellow)
	}
	if f.isError(f.err.Message) {
		f.err.Type = "Error"
		f.err.Color = GetColor(ColorRed)
	}
}

func (f *Filter) isError(line string) bool {
	return !strings.Contains(line, "DEPRECATED")
}

func (f *Filter) isDeprecated(line string) bool {
	return strings.Contains(line, "DEPRECATED")
}
