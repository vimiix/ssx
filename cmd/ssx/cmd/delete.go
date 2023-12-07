package cmd

import (
	"fmt"

	"github.com/vimiix/cobra"
)

func newDeleteCmd() *cobra.Command {
	var ids []int
	cmd := &cobra.Command{
		Use:   "del",
		Short: "Delete an entry by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(ids) == 0 {
				fmt.Println("no id specified, do nothing")
				return nil
			}
			return ssxInst.DeleteEntryByID(cmd.Context(), ids...)
		},
	}
	cmd.Flags().IntSliceVarP(&ids, "id", "i", nil, "Entry ID")
	return cmd
}
