package cmd

import (
	"github.com/spf13/cobra"
)

func newTagCmd() *cobra.Command {
	var (
		tags []string
		id   int
	)
	cmd := &cobra.Command{
		Use:     "tag",
		Short:   "Tag an entry by ID",
		Example: "ssx tag -i 1 -t tag1 [-t tag2]",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ssxInst.AppendTagByID(id, tags...)
		},
	}

	cmd.Flags().StringSliceVarP(&tags, "tag", "t", nil, "Tag name")
	cmd.Flags().IntVarP(&id, "id", "i", 0, "Entry ID")
	cmd.MarkFlagsRequiredTogether("id", "tag")
	return cmd
}
