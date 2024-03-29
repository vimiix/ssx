package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	var ids []int
	cmd := &cobra.Command{
		Use:     "delete",
		Aliases: []string{"d", "del"},
		Short:   "delete entry by id",
		Example: "ssx delete -i1 [-i2 ...]",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(ids) == 0 {
				fmt.Println("no id specified, do nothing")
				return nil
			}
			return ssxInst.DeleteEntryByID(ids...)
		},
	}
	cmd.Flags().IntSliceVarP(&ids, "id", "i", nil, "entry id")
	return cmd
}
