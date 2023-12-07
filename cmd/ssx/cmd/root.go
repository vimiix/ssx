package cmd

import (
	"fmt"
	"os"

	"github.com/vimiix/cobra"

	"github.com/vimiix/ssx/internal/lg"
	"github.com/vimiix/ssx/ssx"
	"github.com/vimiix/ssx/ssx/version"
)

var (
	logVerbose   bool
	printVersion bool
	ssxInst      *ssx.SSX
)

func NewRoot() *cobra.Command {
	opt := &ssx.CmdOption{}
	root := &cobra.Command{
		Use:                "ssx",
		Short:              "ssx is a ssh wrapper",
		SilenceUsage:       true,
		SilenceErrors:      true,
		DisableAutoGenTag:  true,
		DisableSuggestions: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			lg.SetVerbose(logVerbose)
			if !printVersion {
				s, err := ssx.NewSSX(opt)
				if err != nil {
					return err
				}
				ssxInst = s
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if printVersion {
				fmt.Fprintln(os.Stdout, version.Detail())
				return nil
			}
			return ssxInst.Main(cmd.Context())
		},
	}

	root.PersistentFlags().BoolVarP(&printVersion, "version", "v", false, "Print ssx version")
	root.PersistentFlags().BoolVar(&logVerbose, "verbose", false, "Output detail logs")

	root.Flags().StringVarP(&opt.DBFile, "file", "f", "", "Filepath to store auth data")
	root.Flags().StringVarP(&opt.Addr, "server", "s", "", "Target server address\nSupport formats: [user@]host[:port]")
	root.Flags().StringVarP(&opt.Tag, "tag", "t", "", "Target server address\nSupport formats: [user@]host[:port]")

	root.AddCommand(newListCmd())
	root.AddCommand(newDeleteCmd())

	root.CompletionOptions.HiddenDefaultCmd = true
	root.SetHelpCommand(&cobra.Command{Hidden: true})
	return root
}
