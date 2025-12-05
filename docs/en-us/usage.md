# Usage

## Add Entry

A successful login automatically creates a new entry.

```bash
ssx [-J PROXY_USER@PROXY_HOST:PROXY_PORT] [USER@]HOST[:PORT] [-i IDENTITY_FILE] [-p PORT]
```

| Parameter | Description | Required | Default |
|:---|:---|:---|:---|
| `USER` | OS user to login as | No | `root` |
| `HOST` | Target server IP (IPv4 only) | Yes | |
| `PORT` | SSH service port | No | 22 |
| `-i IDENTITY_FILE` | Private key file | No | `~/.ssh/id_rsa` |
| `-J` | Jump server for proxy login (password auth only) | No | |

On first login without an available private key, you'll be prompted to enter a password interactively. Once logged in successfully, the password will be saved to the local data file (default: **~/.ssx/db**, customizable via `SSX_DB_PATH` environment variable).

For subsequent logins, simply run `ssx <IP>` to login automatically.

You can also use partial IP fragments for fuzzy matching. For example, if you have an entry for `192.168.1.100`, you can login with just `ssx 100`.

## List Entries

```bash
ssx list
# output example
# Entries (stored in ssx)
#  ID |       Address        |          Tags
#-----+----------------------+--------------------------
#  1  | root@172.23.1.84:22  | centos
```

By default, ssx doesn't load `~/.ssh/config` unless the `SSX_IMPORT_SSH_CONFIG` environment variable is set.

ssx doesn't store entries from your ssh config file in its database, so you won't see an "ID" field for those entries in the list output.

```bash
export SSX_IMPORT_SSH_CONFIG=true
ssx list
# output example
# Entries (stored in ssx)
#  ID |       Address        |          Tags
#-----+----------------------+--------------------------
#  1  | root@172.23.1.84:22  | centos
#
# Entries (found in ssh config)
#               Address              |           Tags
# -----------------------------------+----------------------------
#   git@ssh.github.com:22            | github.com
```

## Tag Entries

ssx assigns a unique `ID` to each stored server. Use this ID to specify the server entry when tagging.

Use the `tag` subcommand to manage tags:

```bash
ssx tag --id <ENTRY_ID> [-t TAG1 [-t TAG2 ...]] [-d TAG3 [-d TAG4 ...]]
```

- `--id`: Specify the server ID from the list command output
- `-t`: Add tags (can be specified multiple times)
- `-d`: Delete existing tags (can be specified multiple times)

After tagging a server (e.g., with `centos`), you can login using the tag:

```bash
ssx centos
```

## Login to Server

Without any parameter flags, ssx treats the second argument as a search keyword, searching hosts and tags. If no entry matches, ssx treats it as a new entry and attempts to login.

```bash
# Interactive login
ssx

# Login by entry ID
ssx --id <ID>

# Login by address (supports partial match)
ssx <ADDRESS>

# Login by tag
ssx <TAG>
```

## Execute Single Command

SSX supports executing a shell command via the `-c` parameter, then exiting after execution. This is useful for non-interactive remote command execution in embedded scenarios.

```bash
ssx centos -c 'pwd'
```

## File Copy

> v0.6.0+

SSX supports copying files between local and remote hosts using the `cp` subcommand with the SCP protocol.

### Basic Usage

```bash
ssx cp <SOURCE> <TARGET>
```

### Path Formats

- **Local path**: `/path/to/file` or `./relative/path`
- **Remote path**: `[user@]host[:port]:/path/to/file`
- **Tag reference**: `tag:/path/to/file` (use stored entry tag or keyword)

### Upload Files to Remote Server

```bash
# Using full address
ssx cp ./local.txt root@192.168.1.100:/tmp/remote.txt

# Using tag
ssx cp ./local.txt myserver:/tmp/remote.txt

# With custom port
ssx cp ./local.txt root@192.168.1.100:2222:/tmp/remote.txt

# With identity file
ssx cp -i ~/.ssh/id_rsa ./local.txt root@192.168.1.100:/tmp/remote.txt
```

### Download Files from Remote Server

```bash
# Using full address
ssx cp root@192.168.1.100:/tmp/remote.txt ./local.txt

# Using tag
ssx cp myserver:/tmp/remote.txt ./local.txt
```

### Remote-to-Remote Copy

SSX supports copying files directly between two remote servers. Files are streamed through SSX without being stored locally.

```bash
# Using full addresses
ssx cp root@192.168.1.100:/tmp/file.txt root@192.168.1.200:/tmp/file.txt

# Using tags
ssx cp server1:/data/file.txt server2:/backup/file.txt
```

### cp Command Options

| Option | Description | Default |
|:---|:---|:---|
| `-i, --identity-file` | Private key file path | |
| `-J, --jump-server` | Jump server address | |
| `-P, --port` | Remote host port | 22 |

## Upgrade SSX

> v0.3.0+

```bash
ssx upgrade [<version>]
```

Without specifying a version, it automatically updates to the latest version on GitHub.
