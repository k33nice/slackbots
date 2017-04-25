package service

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
)

// Service - object that aid to manage system services.
type Service struct {
	Fixer  Fixer
	Result *Result
	Log    *log.Logger
}

// Result - handle os command exec output.
type Result struct {
	Stdout string
	IsOk   bool
}

// Fixer - interface for fix typo.
type Fixer interface {
	Fix(string) string
}

// Run - execute system command passed as string.
func (s *Service) Run(cmd string, wg *sync.WaitGroup) {
	defer wg.Done()

	s.Result = &Result{IsOk: false, Stdout: ""}
	head, parts, err := s.tokenizer(cmd)

	if err != nil {
		s.Log.Println(err)
		s.Result.IsOk = false
		s.Result.Stdout = err.Error()
		return
	}

	s.Log.Printf("Try to exec command: %s, with args: %v", head, parts)
	out, err := exec.Command(head, parts...).Output()

	if err != nil {
		s.Log.Println(err)
		s.Result.IsOk = false
		return
	}

	s.Result.Stdout = string(out)
	s.Result.IsOk = true
}

// GetResult - return command execution result.
func (s *Service) GetResult() (res *Result) {
	return s.Result
}

func (s *Service) tokenizer(cmd string) (head string, parts []string, err error) {
	tokens := strings.Fields(cmd)

	for i, w := range tokens {
		tokens[i] = s.Fixer.Fix(w)
	}

	if len(tokens) < 2 {
		err = errors.New("Command string must have at least one argument")
	}

	head = tokens[0]
	parts = tokens[1:len(tokens)]

	switch head {
	case "service", "php":
	default:
		err = fmt.Errorf("Invalid command: %s", head)
	}

	return
}
