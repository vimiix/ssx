package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

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
		Use:   "ssx",
		Short: "🦅 ssx is a retentive ssh client",
		Example: `# If more than one flag of -i, -s ,-t specified, priority is ENTRY_ID > ADDRESS > TAG_NAME
ssx [-i ENTRY_ID] [-s [USER@]HOST[:PORT]] [-k IDENTITY_FILE] [-t TAG_NAME]`,
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
	root.Flags().StringVarP(&opt.DBFile, "file", "f", "", "filepath to store auth data")
	root.Flags().Uint64VarP(&opt.EntryID, "id", "i", 0, "entry id")
	root.Flags().StringVarP(&opt.Addr, "server", "s", "", "target server address\nsupport formats: [user@]host[:port]")
	root.Flags().StringVarP(&opt.Tag, "tag", "t", "", "search entry by tag")
	root.Flags().StringVarP(&opt.IdentityFile, "keyfile", "k", "", "identity_file path")

	root.PersistentFlags().BoolVarP(&printVersion, "version", "v", false, "print ssx version")
	root.PersistentFlags().BoolVar(&logVerbose, "verbose", false, "output detail logs")

	root.AddCommand(newListCmd())
	root.AddCommand(newDeleteCmd())
	root.AddCommand(newTagCmd())

	root.CompletionOptions.HiddenDefaultCmd = true
	root.SetHelpCommand(&cobra.Command{Hidden: true})
	return root
}
