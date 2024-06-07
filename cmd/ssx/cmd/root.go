package cmd

import (
	"fmt"
	"os"
	"strings"

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
ssx [-i ENTRY_ID] [-s [USER@]HOST[:PORT]] [-k IDENTITY_FILE] [-t TAG_NAME]

# You can also skip the parameters and log in directly with host or tag
ssx [USER@]HOST[:PORT]
ssx TAG_NAME

# Fuzzy search is also supported
# For example, you want to login to 192.168.1.100 and
# suppose you can uniquely locate one entry by '100',
# you just need to enter:
ssx 100

# If a command is specified, it will be executed on the remote host instead of a login shell.
ssx 100 -c pwd
# if the '-c' is omitted, the secend and subsequent arguments will be treated as COMMAND
ssx 100 pwd`,
		SilenceUsage:       true,
		SilenceErrors:      true,
		DisableAutoGenTag:  true,
		DisableSuggestions: true,
		Args:               cobra.ArbitraryArgs, // accept arbitrary args for supporting quick login
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
			if len(args) > 0 {
				// just use first word as search key
				opt.Keyword = args[0]
			}
			if len(args) > 1 && len(opt.Command) == 0 {
				opt.Command = strings.Join(args[1:], " ")
			}
			return ssxInst.Main(cmd.Context())
		},
	}
	root.Flags().StringVarP(&opt.DBFile, "file", "f", "", "filepath to store auth data")
	root.Flags().Uint64VarP(&opt.EntryID, "id", "i", 0, "entry id")
	root.Flags().StringVarP(&opt.Addr, "server", "s", "", "target server address\nsupport format: [user@]host[:port]")
	root.Flags().StringVarP(&opt.Tag, "tag", "t", "", "search entry by tag")
	root.Flags().StringVarP(&opt.IdentityFile, "keyfile", "k", "", "identity_file path")
	root.Flags().StringVarP(&opt.JumpServers, "jump-server", "J", "", "jump servers, multiple jump hops may be specified separated by comma characters\nformat: [user1@]host1[:port1][,[user2@]host2[:port2]...]")
	root.Flags().StringVarP(&opt.Command, "cmd", "c", "", "the command to execute\nssh connection will exit after the execution complete")
	root.Flags().DurationVar(&opt.Timeout, "timeout", 0, "timeout for connecting and executing command")

	root.PersistentFlags().BoolVarP(&printVersion, "version", "v", false, "print ssx version")
	root.PersistentFlags().BoolVar(&logVerbose, "verbose", false, "output detail logs")

	root.AddCommand(newListCmd())
	root.AddCommand(newDeleteCmd())
	root.AddCommand(newTagCmd())
	root.AddCommand(newInfoCmd())

	root.CompletionOptions.HiddenDefaultCmd = true
	root.SetHelpCommand(&cobra.Command{Hidden: true})
	return root
}
