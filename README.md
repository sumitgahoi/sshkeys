# sshkeys

`sshkeys` copies the public SSH keys from a GitHub account into `~/.ssh/authorized_keys`. GitHub is the source of truth for the keys it manages.

Keys are read from `https://github.com/<github-user>.keys`. OpenSSH `ssh-keygen` must be on `PATH`.

## Install

On Linux (x86_64 or ARM64), install the latest release:

```bash
curl -fsSL https://raw.githubusercontent.com/sumitgahoi/sshkeys/main/install.sh | sudo bash
```

The script downloads that release, verifies its checksum, installs `/usr/local/bin/sshkeys` and the systemd units, and reloads systemd. `sudo ./install.sh` from a checkout does the same thing.

On macOS, download `sshkeys-darwin-arm64` from the [latest release](https://github.com/sumitgahoi/sshkeys/releases/latest) and place it on `PATH`.

## Setup

Create one file per Linux user. The GitHub username can differ from the Linux username:

```bash
sudo install -d -m 0755 /etc/sshkeys
printf 'GITHUB_USER=%s\n' <github-user> | sudo tee /etc/sshkeys/<linux-user>.conf >/dev/null
```

Run a first sync as that user, then enable the timer. It runs about every six hours:

```bash
sudo systemctl start sshkeys-sync@<linux-user>.service
sudo systemctl enable --now sshkeys-sync@<linux-user>.timer
```

## Usage

```bash
sshkeys status <github-user>
sshkeys sync <github-user>
```

`status` prints the number of GitHub keys and whether the local file matches. It does not change any files. `sync` updates the managed section, or prints `already up to date`.

## What gets written

Only the marked section is replaced. Every other line stays as it is. If `authorized_keys` is missing, `sync` creates it.

```text
# BEGIN SSHKEYS MANAGED
ssh-ed25519 AAAA...
ssh-rsa AAAA...
# END SSHKEYS MANAGED
```

A key removed on GitHub is removed from this section on the next successful sync.

If GitHub returns no keys, the command fails with `no SSH keys found` and leaves `authorized_keys` unchanged. The last managed key stays until you delete that line locally.

## Safety

- Missing, reversed, or duplicate markers make the command fail. Repair the file before syncing again.
- `authorized_keys` must be a regular file. Symlinks and other special files are rejected, including a Proxmox host whose root `authorized_keys` points into `/etc/pve`.
- Before replacing an existing file, `sync` copies it to `authorized_keys.bak`. A sync with nothing to change leaves both files alone. The file and its backup are mode `0600`.
- A failed download, an invalid key, or an error while preparing the backup leaves the current `authorized_keys` in place.

`sshkeys` updates that managed section only. It does not create key pairs, handle private keys, or change SSH server configuration.

## Upgrade

Rerun the installation command above. It replaces the binary and systemd units without changing your configuration or enabled timers.

If the timer unit changes, restart the timer:

```bash
sudo systemctl restart sshkeys-sync@<linux-user>.timer
```
