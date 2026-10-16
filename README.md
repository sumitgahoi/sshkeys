# sshkeys

`sshkeys` is a small CLI for synchronizing SSH public keys from a GitHub account into the local `~/.ssh/authorized_keys`.

GitHub is the source of truth for keys managed by `sshkeys`. Existing keys and configuration outside the managed section are left untouched.

## Usage

Check whether the local file matches the keys published by a GitHub account:

```bash
sshkeys status <github-user>
```

Synchronize the keys:

```bash
sshkeys sync <github-user>
```

For example:

```bash
sshkeys status sumitgahoi
sshkeys sync sumitgahoi
```

Keys are fetched from:

```text
https://github.com/<github-user>.keys
```

## How it works

`sshkeys` owns only a clearly marked section of `authorized_keys`:

```text
# BEGIN SSHKEYS MANAGED
ssh-ed25519 AAAA...
ssh-rsa AAAA...
# END SSHKEYS MANAGED
```

Everything outside that section is preserved.

On sync:

1. Fetch the account's public SSH keys from GitHub.
2. Validate every key using the installed `ssh-keygen`.
3. Construct the complete desired `authorized_keys`.
4. If nothing changed, leave the file untouched.
5. Back up an existing file to `authorized_keys.bak`.
6. Replace the managed section with the current GitHub keys.

If `authorized_keys` does not exist, it is created.

Removing a key from GitHub removes it from the managed section on the next sync, provided at least one GitHub key remains.

An empty GitHub key list is rejected to protect SSH access. Removing the final GitHub key does not automatically revoke the last managed key; remove it from the local managed section explicitly if revocation is required.

## Safety

`sshkeys` is deliberately conservative around SSH access.

- Existing content outside the managed section is preserved.
- Malformed or duplicate managed markers cause the operation to fail.
- An existing `authorized_keys` must be a regular file.
- Symlinks and other special files are rejected rather than followed.
- The replacement file and backup are written with `0600` permissions.
- The working file is not replaced unless preparation and backup succeed.
- A no-op sync does not rewrite the file or its backup.
- Sync assumes a single writer. Do not modify `authorized_keys` or run another sync concurrently.

The symlink rule means some systems with special SSH-key management, such as a Proxmox host whose root `authorized_keys` points into `/etc/pve`, should be managed separately.

## Build

Requires Go and OpenSSH's `ssh-keygen`.

```bash
go build -o sshkeys .
```

Cross-compile for Linux AMD64:

```bash
GOOS=linux GOARCH=amd64 go build -o sshkeys-linux-amd64 .
```

For Linux ARM64:

```bash
GOOS=linux GOARCH=arm64 go build -o sshkeys-linux-arm64 .
```

## Install

On Linux, install the latest release:

```bash
curl -fsSL https://raw.githubusercontent.com/sumitgahoi/sshkeys/main/install.sh | sudo bash
```

The script detects the architecture, downloads and verifies that release, installs `/usr/local/bin/sshkeys` and the systemd units, and reloads systemd. From a checkout, `sudo ./install.sh` does the same thing.

Create one configuration file per Linux user. The GitHub username can differ from the Linux username:

```bash
sudo install -d -m 0755 /etc/sshkeys
printf 'GITHUB_USER=%s\n' sumitgahoi | sudo tee /etc/sshkeys/sumit.conf >/dev/null
```

Run the first synchronization, then enable the timer. The service runs as that Linux user:

```bash
sudo systemctl start sshkeys-sync@sumit.service
sudo systemctl enable --now sshkeys-sync@sumit.timer
```

The timer runs that service about every six hours.

On macOS, build from source or download `sshkeys-darwin-arm64` from the latest release and place it on `PATH`.

## Upgrade

Run the installer again:

```bash
curl -fsSL https://raw.githubusercontent.com/sumitgahoi/sshkeys/main/install.sh | sudo bash
```

This replaces the binary and unit files. Configuration under `/etc/sshkeys` and existing timer enablement stay as they are.

## Test

```bash
go test ./...
go test -race ./...
go vet ./...
```

## Scope

`sshkeys` intentionally stays small. It does not manage private keys, generate key pairs, modify SSH server configuration, or attempt to interpret arbitrary `authorized_keys` syntax.

Its job is simply:

**GitHub public keys → one safely managed section of `authorized_keys`.**
