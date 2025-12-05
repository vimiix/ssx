# Environment Variables

SSX supports the following environment variables:

| Variable | Description | Default |
|:---|:---|:---|
| `SSX_DB_PATH` | Database file for storing entries | ~/.ssx.db |
| `SSX_CONNECT_TIMEOUT` | SSH connection timeout (supports h/m/s units) | `10s` |
| `SSX_IMPORT_SSH_CONFIG` | Whether to import user ssh config | |
| `SSX_SECRET_KEY` | [Deprecated in v0.4+] For backward compatibility, equivalent to `SSX_DEVICE_ID` | |
| `SSX_DEVICE_ID` | Device ID to bind the database file. Set the same value across devices to share a database | [Device ID](#device-id) |

## Explanation

### SSX_IMPORT_SSH_CONFIG

When this environment variable is not set, ssx doesn't read the user's `~/.ssh/config` file by default. ssx only uses its own storage file for searching. If you set this environment variable to any non-empty string, ssx will load server entries from the user's ssh config file during initialization. However, ssx only reads these for searching and login purposes - it doesn't persist them to ssx's storage file. So when you run `ssx IP` and that IP is already configured in `~/.ssh/config` with authentication, ssx will match and login directly. In `ssx list`, these servers appear in the `found in ssh config` table, which doesn't have an ID property.

### Device ID

- Linux uses `/var/lib/dbus/machine-id` ([man](http://man7.org/linux/man-pages/man5/machine-id.5.html))
- OS X uses `IOPlatformUUID`
- Windows uses `MachineGuid` from `HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Cryptography`
