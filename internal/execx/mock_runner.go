package execx

import (
	"context"
	"fmt"
	"strings"
)

type Expectation struct {
	Name   string
	Args   []string
	Result Result
	Err    error
}

type MockRunner struct {
	expectations []Expectation
	index        int
}

func NewMockRunner(expectations ...Expectation) *MockRunner {
	copied := make([]Expectation, len(expectations))
	for i, expectation := range expectations {
		copied[i] = Expectation{
			Name: expectation.Name,
			Args: append([]string(nil), expectation.Args...),
			Result: Result{
				Stdout: append([]byte(nil), expectation.Result.Stdout...),
				Stderr: append([]byte(nil), expectation.Result.Stderr...),
			},
			Err: expectation.Err,
		}
	}
	return &MockRunner{expectations: copied}
}

func (m *MockRunner) Run(_ context.Context, name string, args ...string) (Result, error) {
	if m.index >= len(m.expectations) {
		return Result{}, fmt.Errorf("unexpected command: %s", signature(name, args))
	}

	expectation := m.expectations[m.index]
	m.index++

	if expectation.Name != name || !sameArgs(expectation.Args, args) {
		return Result{}, fmt.Errorf(
			"unexpected command: got %s, want %s",
			signature(name, args),
			signature(expectation.Name, expectation.Args),
		)
	}

	return Result{
		Stdout: append([]byte(nil), expectation.Result.Stdout...),
		Stderr: append([]byte(nil), expectation.Result.Stderr...),
	}, expectation.Err
}

func (m *MockRunner) AssertDone() error {
	if m.index == len(m.expectations) {
		return nil
	}
	expectation := m.expectations[m.index]
	return fmt.Errorf("unmet expectation: %s", signature(expectation.Name, expectation.Args))
}

func sameArgs(want []string, got []string) bool {
	if len(want) != len(got) {
		return false
	}
	for i := range want {
		if want[i] != got[i] {
			return false
		}
	}
	return true
}

func signature(name string, args []string) string {
	if len(args) == 0 {
		return name
	}
	return name + " " + strings.Join(args, " ")
}
