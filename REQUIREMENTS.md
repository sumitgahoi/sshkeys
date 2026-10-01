# sshkeys requirements

`sshkeys` synchronizes SSH public keys registered with a GitHub account into the local user's `~/.ssh/authorized_keys`.

The design should remain small, auditable, and conservative. Prefer Go's standard library and existing OpenSSH tools over custom implementations or additional dependencies.

## Commands

```text
sshkeys status <github-user>
sshkeys sync <github-user>
```

Any other invocation should fail with a useful usage message and non-zero exit status.

## Source of truth

Keys are fetched from:

```text
https://github.com/<github-user>.keys
```

GitHub is authoritative for keys inside the section managed by `sshkeys`.

Key comments or GitHub key titles are not required. SSH key identity is determined by the key itself/fingerprint.

Fetching GitHub keys must have a finite network timeout.

A failed fetch must never modify `authorized_keys`.

## Key validation

Ignore blank lines and comment lines in fetched key data.

At least one valid SSH public key must be returned.

Validate every fetched key before modifying any local SSH configuration.

Use the installed OpenSSH `ssh-keygen` implementation for validation rather than implementing SSH key parsing or cryptography in this project.

If any fetched key is invalid, abort without modifying `authorized_keys`.

## Managed section

`sshkeys` owns only this section:

```text
# BEGIN SSHKEYS MANAGED
<keys>
# END SSHKEYS MANAGED
```

Content outside the managed section belongs to the user and must be preserved.

On synchronization, replace the complete managed section with the current validated GitHub key set.

Therefore:

- keys added to GitHub are added on the next sync;
- keys removed from GitHub are removed from the managed section on the next sync;
- unmanaged keys outside the markers are untouched.

If no managed section exists, append one without destroying existing file contents.

Malformed or ambiguous marker state must cause the operation to fail without modifying the file. This includes missing, reversed, or duplicate markers.

## First synchronization

If `~/.ssh/authorized_keys` does not exist, `sync` should create it containing the managed section.

There is no previous file to back up in this case, so `authorized_keys.bak` should not be created solely for first-time creation.

The resulting `authorized_keys` file must use safe SSH permissions (`0600`).

## Existing-file synchronization

Before changing an existing `authorized_keys`:

1. Fetch and validate all GitHub keys.
2. Construct the complete desired `authorized_keys` contents in memory.
3. Validate the managed-section structure.
4. If the desired file is byte-for-byte identical to the existing file, perform no filesystem changes.
5. Prepare safe replacement data.
6. Preserve the previous complete `authorized_keys` as `authorized_keys.bak`.
7. Replace `authorized_keys` atomically.

The backup must represent the complete file immediately preceding a successful change.

Do not rename the working `authorized_keys` away before installing its replacement. There must never intentionally be a window where an existing working `authorized_keys` is absent.

Writes to both `authorized_keys` and its backup should use same-directory temporary files followed by atomic rename, so partially written destination files are not exposed.

The backup replacement and `authorized_keys` replacement are two separate atomic operations; they are not expected to form a filesystem transaction. Backup installation must happen first so a failure between the operations leaves the existing working `authorized_keys` intact.

## No-op synchronization

If GitHub's current keys produce exactly the existing `authorized_keys` contents:

- do not rewrite `authorized_keys`;
- do not replace `authorized_keys.bak`;
- report that the file is already up to date.

## Status

`status` must not modify local files.

It should fetch and validate the same GitHub key set used by `sync` and report:

```text
<N> GitHub keys
authorized_keys: up to date
```

or:

```text
<N> GitHub keys
authorized_keys: out of date
```

or:

```text
<N> GitHub keys
authorized_keys: not found
```

`status` and `sync` should use the same definition of desired managed-section contents.

Malformed local managed markers should produce an error rather than being reported merely as out of date.

## Failure behavior

Safety takes priority over attempting recovery automatically.

Failures involving networking, validation, marker parsing, temporary-file creation, backup creation, or replacement must return a useful error and must not intentionally destroy a previously working `authorized_keys`.

Do not silently repair ambiguous managed-marker state.

## Dependencies

Keep the implementation minimal.

Preferred dependencies:

- Go standard library
- installed OpenSSH `ssh-keygen`

Avoid adding third-party Go packages unless they solve a concrete requirement that cannot reasonably be handled by the standard library or OpenSSH.

The CLI must not require Git, a GitHub API token, a configuration file, or a local clone of a repository at runtime.

## Testing

Tests should cover behavior and safety invariants, including at minimum:

- parsing fetched key data;
- accepting a valid SSH public key;
- rejecting an invalid SSH public key;
- creating a managed section;
- replacing an existing managed section;
- preserving unmanaged content;
- rejecting malformed/duplicate marker state;
- first-time creation of `authorized_keys`;
- backup of the previous file during a change;
- no backup during first-time creation;
- no filesystem changes when already up to date;
- `status` for missing, current, and stale files.

Tests should use temporary directories/files and must not modify the developer's real `~/.ssh` directory.

## Scope

Keep v1 intentionally narrow.

Not required for v1:

- SSH key generation;
- adding keys to GitHub;
- GitHub API authentication;
- GitHub key titles/device names;
- scheduling or daemon mode;
- automatic remote-host discovery;
- configuration files;
- third-party SSH libraries;
- synchronization of private keys.
- sshkeys assumes a single writer. Do not run concurrent sync operations or modify authorized_keys while a sync is in progress. The backup guarantee assumes serialized writes.

A user enrolls a device by creating/managing their SSH key using normal OpenSSH tooling and registering the public key with GitHub. `sshkeys` handles synchronization of the resulting GitHub public-key set into `authorized_keys`.