package cmd

import (
	"fmt"
	"os"

	"github.com/vimiix/cobra"

	"github.com/vimiix/ssx/internal/version"
	"github.com/vimiix/ssx/pkg/lg"
)

const (
	use   = "ssx"
	short = "ssx is a ssh wrapper"
)

var Root = New(use, short)

func New(use, short string) *cobra.Command {
	verboseFlag := false
	versionFlag := false
	root := &cobra.Command{
		Use:               use,
		Short:             short,
		SilenceUsage:      true,
		SilenceErrors:     true,
		DisableAutoGenTag: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			lg.SetVerbose(verboseFlag)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if versionFlag {
				fmt.Fprintln(os.Stdout, version.Detail())
				return nil
			}
			return cmd.Usage()
		},
	}

	root.PersistentFlags().BoolVarP(&versionFlag, "version", "v", false, "Print ssx version")
	root.PersistentFlags().BoolVar(&verboseFlag, "verbose", false, "Output detail logs")

	return root
}
