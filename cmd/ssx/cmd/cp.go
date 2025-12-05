package cmd

import (
	"github.com/spf13/cobra"

	"github.com/vimiix/ssx/ssx"
)

func newCpCmd() *cobra.Command {
	opt := &ssx.CpOption{}
	cmd := &cobra.Command{
		Use:   "cp <SOURCE> <TARGET>",
		Short: "copy files between local and remote hosts",
		Long: `Copy files between local and remote hosts using SCP protocol.
Supports local-to-remote, remote-to-local, and remote-to-remote transfers.

For remote-to-remote transfers, files are streamed through ssx without
being stored locally, acting as a relay between the two remote hosts.

Path format:
  Local:  /path/to/file or ./relative/path
  Remote: [user@]host[:port]:/path/to/file
          tag:/path/to/file (use stored entry by tag/keyword)

Examples:
  # Upload local file to remote
  ssx cp ./local.txt root@192.168.1.100:/tmp/remote.txt
  ssx cp ./local.txt myserver:/tmp/remote.txt

  # Download remote file to local
  ssx cp root@192.168.1.100:/tmp/remote.txt ./local.txt
  ssx cp myserver:/tmp/remote.txt ./local.txt

  # Remote to remote (streaming through ssx)
  ssx cp root@192.168.1.100:/tmp/file.txt root@192.168.1.200:/tmp/file.txt
  ssx cp server1:/data/file.txt server2:/backup/file.txt

  # With custom port
  ssx cp ./local.txt root@192.168.1.100:2222:/tmp/remote.txt

  # With identity file
  ssx cp -i ~/.ssh/id_rsa ./local.txt root@192.168.1.100:/tmp/remote.txt`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opt.Source = args[0]
			opt.Target = args[1]
			return ssxInst.Copy(cmd.Context(), opt)
		},
	}

	cmd.Flags().StringVarP(&opt.IdentityFile, "identity-file", "i", "", "identity file path for authentication")
	cmd.Flags().StringVarP(&opt.JumpServers, "jump-server", "J", "", "jump servers (proxy)")
	cmd.Flags().IntVarP(&opt.Port, "port", "P", 22, "port to connect to on the remote host")

	return cmd
}
