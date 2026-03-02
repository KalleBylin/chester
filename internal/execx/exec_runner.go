package execx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
)

type ExecRunner struct{}

func (r ExecRunner) Run(ctx context.Context, name string, args ...string) (Result, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return Result{}, fmt.Errorf("%w: %s", ErrCommandNotFound, name)
		}

		runErr := &RunError{
			Name:   name,
			Args:   append([]string(nil), args...),
			Stderr: stderr.Bytes(),
			Err:    err,
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			runErr.ExitCode = exitErr.ExitCode()
		}
		return Result{
			Stdout: stdout.Bytes(),
			Stderr: stderr.Bytes(),
		}, runErr
	}

	return Result{
		Stdout: stdout.Bytes(),
		Stderr: stderr.Bytes(),
	}, nil
}
