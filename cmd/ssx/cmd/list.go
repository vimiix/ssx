package cmd

import (
	"github.com/vimiix/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ssxInst.ListEntries()
		}}

	return cmd
}
