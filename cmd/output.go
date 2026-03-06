package cmd

import (
	"fmt"

	"github.com/KalleBylin/chester/internal/app"
	"github.com/spf13/cobra"
)

func writeCommandOutput(cmd *cobra.Command, asJSON bool, markdown string, value any) error {
	if asJSON {
		output, err := app.RenderJSON(value)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), output)
		return err
	}

	_, err := fmt.Fprintln(cmd.OutOrStdout(), markdown)
	return err
}
