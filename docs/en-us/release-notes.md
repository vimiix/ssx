# Release Notes

## v0.6.0

Release Date: TBD

### Features

- Added `cp` subcommand for file copy operations using SCP protocol
- Support local-to-remote file upload
- Support remote-to-local file download
- Support remote-to-remote file transfer (streaming through ssx without local storage)
- Support tag/keyword reference in remote paths

## v0.5.0

Release Date: November 14, 2024

### BREAKING CHANGE

- Changed the SSH identity file flag from `-k` to `-i` to be more compatible with the standard `ssh` command
- Changed the entry ID flag from `-i` to `--id` across all commands for consistency
- Updated command examples and help text to reflect the new flag names

### Why

- The `-i` flag is the standard flag for specifying identity files in SSH, making SSX more intuitive for users familiar with SSH
- Using `--id` for entry IDs makes the parameter name more descriptive and avoids conflict with the SSH identity file flag

### Example Usage

```text
Old: ssx delete -i 123
New: ssx delete --id 123

Old: ssx -k ~/.ssh/id_rsa
New: ssx -i ~/.ssh/id_rsa
```

## v0.4.3

Release Date: September 20, 2024

**Bug Fix:**

- Fixed "unexpected fault address 0xxxxx" issue on Mac M1 ([#62](https://github.com/vimiix/ssx/issues/62))

## v0.4.2

Release Date: September 18, 2024

**Changelog:**

- Updated dependency library versions

## v0.4.1

Release Date: August 28, 2024

**Changelog:**

- Updated dependency library versions

## v0.4.0

Release Date: July 10, 2024

**Features:**

- Mandatory admin password for database file; first login will prompt if not set
- Added `SSX_DEVICE_ID` environment variable; database file binds to device by default; migration requires admin password verification

**BREAKING CHANGE:**

- Deprecated `--unsafe` parameter
- Deprecated `SSX_UNSAFE_MODE` and `SSX_SECRET_KEY` environment variables
- If safe mode entries exist from older versions, password re-entry is required on login

## v0.3.1

Release Date: June 18, 2024

**Features:**

- Version check before upgrade to avoid redundant upgrades
- No need to load entry database during upgrade, reducing unnecessary logic

## v0.3.0

Release Date: June 12, 2024

**Features:**

- Support refreshing stored key records via `-k` parameter
- Support online upgrade

**BREAKING CHANGE:**

- Marked `--server` and `--tag` as deprecated parameters

## v0.2.0

Release Date: June 11, 2024

**Features:**

- Added `-p` parameter for explicit port specification
- Added `-J` parameter for jump server login
- Default encryption of entry passwords using device ID
- Changed default login user from current user to root

## v0.1.0

Release Date: February 29, 2024

**Features:**

- Completed initial design requirements, implemented minimum viable version
