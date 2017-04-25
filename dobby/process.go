package main

import "strings"

// TODO: Improve processes detection.

// Process - object for handle script strating.
type Process struct {
	name string
	line string
}

func (p *Process) isStart() bool {
	return strings.Contains(p.line, "START")
}

func (p *Process) isEnd() bool {
	return strings.Contains(p.line, "FINISH")
}
