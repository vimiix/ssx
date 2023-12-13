package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newTagCmd() *cobra.Command {
	var (
		appendtTags []string
		deleteTags  []string
		id          int
	)
	cmd := &cobra.Command{
		Use:     "tag",
		Aliases: []string{"t"},
		Short:   "add or delete tag for entry by id",
		Example: "ssx tag -i <ENTRY_ID> [-t TAG1 [-t TAG2 ...]] [-d TAG3 [-d TAG4 ...]]",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(appendtTags) == 0 && len(deleteTags) == 0 {
				return errors.New("no tag is spicified")
			}
			if len(deleteTags) > 0 {
				if err := ssxInst.DeleteTagByID(id, deleteTags...); err != nil {
					return err
				}
			}
			if len(appendtTags) > 0 {
				if err := ssxInst.AppendTagByID(id, appendtTags...); err != nil {
					return err
				}
			}
			return nil
		},
	}

	cmd.Flags().IntVarP(&id, "id", "i", 0, "entry id")
	cmd.Flags().StringSliceVarP(&appendtTags, "tag", "t", nil, "tag name to add")
	cmd.Flags().StringSliceVarP(&deleteTags, "delete", "d", nil, "tag name to delete")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}
