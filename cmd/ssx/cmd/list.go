package cmd

import (
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "ls"},
		Short:   "list all entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ssxInst.ListEntries()
		}}

	return cmd
}
