package execx

import "context"

type MuxRunner struct {
	Fallback Runner
	Routes   map[string]Runner
}

func NewMuxRunner(fallback Runner) *MuxRunner {
	return &MuxRunner{
		Fallback: fallback,
		Routes:   make(map[string]Runner),
	}
}

func (m *MuxRunner) Set(name string, runner Runner) {
	if m.Routes == nil {
		m.Routes = make(map[string]Runner)
	}
	m.Routes[name] = runner
}

func (m *MuxRunner) Run(ctx context.Context, name string, args ...string) (Result, error) {
	if runner, ok := m.Routes[name]; ok {
		return runner.Run(ctx, name, args...)
	}
	return m.Fallback.Run(ctx, name, args...)
}
