package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/vimiix/cobra"

	"github.com/vimiix/ssx/internal/config"
	"github.com/vimiix/ssx/internal/selector"
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
			nodes, err := config.Load(cmd.Context())
			if err != nil {
				lg.Errorf(err.Error())
				return err
			}

			if len(nodes) == 0 {
				fmt.Fprintln(os.Stderr, color.YellowString(`No nodes have been configured.
You can put config file ".ssx.yaml" in home directory or current directory or declare with SSXCONFIG environment.
Also you can add node via command: "%s add", it will append node to default config file (~/.ssx.yaml)`, os.Args[0]))
				return errors.New("no nodes")
			}
			// TODO
			_ = selector.NewSelector(nodes, nil)
			return cmd.Usage()
		},
	}

	root.PersistentFlags().BoolVarP(&versionFlag, "version", "v", false, "Print ssx version")
	root.PersistentFlags().BoolVar(&verboseFlag, "verbose", false, "Output detail logs")

	return root
}
