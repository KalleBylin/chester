package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

const agentsContent = `## Repository Archaeology

This project uses **chester** for deterministic repository archaeology.
It is named for Chesterton's Fence: before deleting or rewriting code, use chester to understand why the code exists.

**Anti-magic rule:**
- Provide exact IDs, file paths, line ranges, or git ranges.
- chester retrieves data only; it does not guess targets or make qualitative judgments.

**Quick reference:**
- ` + "`chester read-thread <id>`" + ` - Fetch the issue or PR conversation
- ` + "`chester file-history <path>`" + ` - Walk the history of one exact file
- ` + "`chester unearth-lines <file> -L <start>,<end>`" + ` - Explain why exact lines exist
- ` + "`chester unearth-range <from>..<to>`" + ` - Summarize PR intent across a git range

For full command details: ` + "`chester --help`" + ``

func renderOnboardInstructions(w io.Writer) error {
	writef := func(format string, args ...any) error {
		_, err := fmt.Fprintf(w, format, args...)
		return err
	}
	writeln := func(text string) error {
		_, err := fmt.Fprintln(w, text)
		return err
	}

	if err := writef("\nchester Onboarding\n\n"); err != nil {
		return err
	}
	if err := writeln("Add this minimal snippet to AGENTS.md (or create it):"); err != nil {
		return err
	}
	if err := writeln(""); err != nil {
		return err
	}
	if err := writeln("--- BEGIN AGENTS.MD CONTENT ---"); err != nil {
		return err
	}
	if err := writeln(agentsContent); err != nil {
		return err
	}
	if err := writeln("--- END AGENTS.MD CONTENT ---"); err != nil {
		return err
	}
	if err := writeln(""); err != nil {
		return err
	}
	if err := writeln("For GitHub Copilot users:"); err != nil {
		return err
	}
	if err := writeln("Add the same content to .github/copilot-instructions.md"); err != nil {
		return err
	}
	if err := writeln(""); err != nil {
		return err
	}
	if err := writeln("How it works:"); err != nil {
		return err
	}
	if err := writeln("  - chester keeps AGENTS.md lean while still teaching agents the exact primitives"); err != nil {
		return err
	}
	if err := writeln("  - agents choose exact inputs and compose commands themselves"); err != nil {
		return err
	}
	if err := writeln("  - the anti-magic rule keeps the tool deterministic and token-efficient"); err != nil {
		return err
	}
	return nil
}

func newOnboardCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "onboard",
		Short:        "Display a minimal snippet for AGENTS.md",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return renderOnboardInstructions(cmd.OutOrStdout())
		},
	}
}
