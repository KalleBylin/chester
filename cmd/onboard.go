package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

const agentsContent = `## Repository Archaeology

Use ` + "`chester`" + ` before changing unfamiliar code so you understand why it exists.

Rules:
- Copy the command forms below exactly. Use exact IDs, paths, lines, literals, or git ranges.
- Start with the narrowest command.
- ` + "`chester`" + ` only returns history and discussion. You decide what matters.

Commands:
- ` + "`chester why-lines <file>:<start>:<end>`" + ` - why exact code exists
- ` + "`chester why-file <path>`" + ` - history for one file
- ` + "`chester text-history <literal>`" + ` - history for one exact symbol or string
- ` + "`chester why-range <from>..<to>`" + ` - history for a git range
- ` + "`chester read-thread <id>`" + ` - PR or issue discussion

If a ` + "`chester`" + ` command gives you a PR number, follow it with ` + "`chester read-thread <id>`" + `.

Use ` + "`--json`" + ` only when another tool needs structured output instead of Markdown.`

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
