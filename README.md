<p align="center">
    <img src="https://raw.githubusercontent.com/vimiix/ssx/master/static/logo.svg?sanitize=true"
        height="130">
</p>

<p align="center">
    <a href="https://github.com/vimiix/ssx/blob/main/LICENSE" alt="license">
    <img src="https://img.shields.io/badge/License-MIT-jasper" /></a>
    <a href="https://github.com/vimiix" alt="author">
    <img src="https://img.shields.io/badge/Author-Vimiix-blue" /></a>
</p>

🦅 ssx is a retentive ssh client.

It will automatically remember the server which login through it,
so you do not need to enter the password again when you log in again.

<p align="center">
    <img src="https://raw.githubusercontent.com/vimiix/ssx/master/static/demo.svg?sanitize=true"
        height="500">
</p>

## Getting Started

### Installation

Download binary from [releases](https://github.com/vimiix/ssx/releases), extract it and add its path to your `$PATH` list.

If you want to install from source code, you can run command under project root directory:

```bash
make ssx
```

then you can get ssx binary in **dist** directory.

### Add a new entry

```bash
ssx [-s] [USER@]HOST[:PORT] [-k IDENTITY_FILE]
```

If given address matched an exist entry, ssx will login directly.

### List exist entries

```bash
ssx list
# output example
# Entries (stored in ssx)
#  ID |       Address        |          Tags
#-----+----------------------+--------------------------
#  1  | root@172.23.1.84:22  | centos
```

ssx does not read `~/.ssh/config` by default unless the environment variable `SSX_IMPORT_SSH_CONFIG` is set.
ssx will not store user ssh config entries to itself db, so you won't see their `ID` in the output of the list command

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

### Tag an entry

```bash
ssx tag -i <ENTRY_ID> [-t TAG1 [-t TAG2 ...]] [-d TAG3 [-d TAG4 ...]]
```

Once we tag the entry, we can log in through the tag later.

### Login

If more than one flag of `-i`, `-s` ,`-t` specified,

priority is `ENTRY_ID > ADDRESS > TAG_NAME`

If not specified any flag, ssx will treat the second argument as a keyword for searching from host and tags, if not matched any entry, ssx will treat it as a new entry, and try to login.

```bash
# login by interacting, just run ssx
ssx

# login by entry id
ssx -i <ID>

# login by address, support partial words
ssx [-s] <ADDRESS>

# login by tag
ssx [-t] <TAG>
```

### Execute command

```bash
ssx <ADDRESS> [-c] <COMMAND> [--timeout 30s]
ssx <TAG> [-c] <COMMAND> [--timeout 30s]

# for example: login 192.168.1.100 and execute command 'pwd':
ssx 1.100 pwd
```

### Delete an entry

```bash
ssx delete -i <ENTRY_ID>
```

## Supported environment variables

- `SSX_DB_PATH`: DB file to store entries, default is `~/.ssx.db`.
- `SSX_CONNECT_TIMEOUT`: SSH connect timeout, default is `10s`.
- `SSX_IMPORT_SSH_CONFIG`: Whether to import the user ssh config, default is empty.

## Copyright

© 2023 Qian Yao

Distributed under the MIT License. See [LICENSE](https://github.com/vimiix/ssx/blob/main/LICENSE) file for details.
