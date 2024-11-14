package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vimiix/ssx/ssx"
)

func newInfoCmd() *cobra.Command {
	opt := new(ssx.CmdOption)
	cmd := &cobra.Command{
		Use:   "info",
		Short: "show entry detail",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				// just use first word as search key
				opt.Keyword = args[0]
			}
			e, err := ssxInst.GetEntry(opt)
			if err != nil {
				return err
			}
			if e.ID <= 0 && opt.Keyword != "" {
				return fmt.Errorf("not matched any entry for %q", opt.Keyword)
			}
			bs, err := e.JSON()
			if err != nil {
				return err
			}
			fmt.Println(string(bs))
			return nil
		}}
	cmd.Flags().Uint64VarP(&opt.EntryID, "id", "", 0, "entry id")
	cmd.Flags().StringVarP(&opt.Addr, "server", "s", "", "target server address\nsupport formats: [user@]host[:port]")
	cmd.Flags().StringVarP(&opt.Tag, "tag", "t", "", "search entry by tag")

	return cmd
}
