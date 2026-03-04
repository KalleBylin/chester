package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

const agentsContent = `## Repository Archaeology

Use ` + "`chester`" + ` before changing unfamiliar code so you understand why it exists.

Rules:
- Copy the command forms below exactly. Do not invent syntax.
- Pass exact IDs, file paths, line ranges, or git ranges.
- ` + "`chester`" + ` only fetches history and discussion. You decide what matters.

Start with the narrowest command:
- ` + "`chester unearth-lines <file>:<start>:<end>`" + ` - why exact code exists
- ` + "`chester file-history <path>`" + ` - history for one file
- ` + "`chester read-thread <id>`" + ` - PR or issue body, comments, and reviews
- ` + "`chester unearth-range <from>..<to>`" + ` - PRs represented by a git range

If a ` + "`chester`" + ` command gives you a PR number, follow it with ` + "`chester read-thread <id>`" + `.`

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
	if err := writeln("Paste this into AGENTS.md or .github/copilot-instructions.md:"); err != nil {
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
