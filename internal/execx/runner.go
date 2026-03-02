package execx

import (
	"context"
	"errors"
	"fmt"
)

var ErrCommandNotFound = errors.New("command not found")

type Runner interface {
	Run(ctx context.Context, name string, args ...string) (Result, error)
}

type Result struct {
	Stdout []byte
	Stderr []byte
}

type RunError struct {
	Name     string
	Args     []string
	ExitCode int
	Stderr   []byte
	Err      error
}

func (e *RunError) Error() string {
	if e == nil {
		return "<nil>"
	}
	base := fmt.Sprintf("%s failed", e.Name)
	if e.ExitCode != 0 {
		base = fmt.Sprintf("%s with exit code %d", base, e.ExitCode)
	}
	if len(e.Stderr) == 0 {
		return base
	}
	return fmt.Sprintf("%s: %s", base, string(e.Stderr))
}

func (e *RunError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}
