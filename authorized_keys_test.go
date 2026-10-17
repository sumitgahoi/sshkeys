package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestReplaceManagedKeys(t *testing.T) {
	data := []byte(`ssh-ed25519 OUTSIDE1 emergency

# BEGIN SSHKEYS MANAGED
ssh-ed25519 OLD old-macbook
# END SSHKEYS MANAGED

ssh-ed25519 OUTSIDE2 bootstrap
`)

	keys := []string{
		"ssh-ed25519 KEY1 macbook-pro",
		"ssh-ed25519 KEY2 macbook-air",
	}

	want := `ssh-ed25519 OUTSIDE1 emergency

# BEGIN SSHKEYS MANAGED
ssh-ed25519 KEY1 macbook-pro
ssh-ed25519 KEY2 macbook-air
# END SSHKEYS MANAGED

ssh-ed25519 OUTSIDE2 bootstrap
`

	got, err := replaceManagedKeys(data, keys)
	if err != nil {
		t.Fatalf("replaceManagedKeys() error = %v", err)
	}

	if string(got) != want {
		t.Fatalf("replaceManagedKeys() =\n%s\nwant:\n%s", got, want)
	}
}

func TestReplaceManagedKeysRejectsDuplicateSections(t *testing.T) {
	data := []byte(
		managedBegin + "\n" +
			managedEnd + "\n" +
			managedBegin + "\n" +
			managedEnd + "\n",
	)

	_, err := replaceManagedKeys(data, []string{"ssh-ed25519 KEY test"})
	if err == nil {
		t.Fatal("replaceManagedKeys() error = nil, want error")
	}
}

func TestReplaceManagedKeysCreatesSection(t *testing.T) {
	data := []byte("ssh-ed25519 OUTSIDE emergency\n")

	keys := []string{
		"ssh-ed25519 KEY1 macbook-pro",
	}

	want := `ssh-ed25519 OUTSIDE emergency

# BEGIN SSHKEYS MANAGED
ssh-ed25519 KEY1 macbook-pro
# END SSHKEYS MANAGED
`

	got, err := replaceManagedKeys(data, keys)
	if err != nil {
		t.Fatalf("replaceManagedKeys() error = %v", err)
	}

	if string(got) != want {
		t.Fatalf("replaceManagedKeys() =\n%s\nwant:\n%s", got, want)
	}
}

func TestSyncAuthorizedKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "authorized_keys")

	initial := `ssh-ed25519 OUTSIDE emergency

# BEGIN SSHKEYS MANAGED
ssh-ed25519 OLD old-macbook
# END SSHKEYS MANAGED
`

	if err := os.WriteFile(path, []byte(initial), 0600); err != nil {
		t.Fatal(err)
	}

	keys := []string{
		"ssh-ed25519 KEY1 macbook-pro",
		"ssh-ed25519 KEY2 macbook-air",
	}

	if _, err := syncAuthorizedKeys(path, keys); err != nil {
		t.Fatalf("syncAuthorizedKeys() error = %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	want := `ssh-ed25519 OUTSIDE emergency

# BEGIN SSHKEYS MANAGED
ssh-ed25519 KEY1 macbook-pro
ssh-ed25519 KEY2 macbook-air
# END SSHKEYS MANAGED
`

	if string(got) != want {
		t.Fatalf("authorized_keys =\n%s\nwant:\n%s", got, want)
	}

	backup, err := os.ReadFile(path + ".bak")
	if err != nil {
		t.Fatal(err)
	}

	if string(backup) != initial {
		t.Fatalf("backup =\n%s\nwant original:\n%s", backup, initial)
	}
}

func TestSyncAuthorizedKeysCreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "authorized_keys")

	keys := []string{
		"ssh-ed25519 KEY1 macbook-pro",
	}

	if _, err := syncAuthorizedKeys(path, keys); err != nil {
		t.Fatalf("syncAuthorizedKeys() error = %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	want := `# BEGIN SSHKEYS MANAGED
ssh-ed25519 KEY1 macbook-pro
# END SSHKEYS MANAGED
`

	if string(got) != want {
		t.Fatalf("authorized_keys =\n%s\nwant:\n%s", got, want)
	}

	if _, err := os.Stat(path + ".bak"); !os.IsNotExist(err) {
		t.Fatalf("backup should not exist on first creation")
	}
}

func TestSyncAuthorizedKeysRemovesFinalManagedKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "authorized_keys")

	initial := `ssh-ed25519 OUTSIDE emergency

# BEGIN SSHKEYS MANAGED
ssh-ed25519 KEY1 macbook-pro
# END SSHKEYS MANAGED

ssh-ed25519 OUTSIDE2 bootstrap
`

	if err := os.WriteFile(path, []byte(initial), 0600); err != nil {
		t.Fatal(err)
	}

	changed, err := syncAuthorizedKeys(path, nil)
	if err != nil {
		t.Fatalf("syncAuthorizedKeys() error = %v", err)
	}

	if !changed {
		t.Fatal("syncAuthorizedKeys() changed = false, want true")
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	want := `ssh-ed25519 OUTSIDE emergency

# BEGIN SSHKEYS MANAGED
# END SSHKEYS MANAGED

ssh-ed25519 OUTSIDE2 bootstrap
`

	if string(got) != want {
		t.Fatalf("authorized_keys =\n%s\nwant:\n%s", got, want)
	}
}

func TestSyncAuthorizedKeysNoOpWhenUnchanged(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "authorized_keys")

	keys := []string{
		"ssh-ed25519 KEY1 macbook-pro",
	}

	data := []byte(`# BEGIN SSHKEYS MANAGED
ssh-ed25519 KEY1 macbook-pro
# END SSHKEYS MANAGED
`)

	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}

	backup := []byte("existing backup\n")
	if err := os.WriteFile(path+".bak", backup, 0600); err != nil {
		t.Fatal(err)
	}

	if _, err := syncAuthorizedKeys(path, keys); err != nil {
		t.Fatalf("syncAuthorizedKeys() error = %v", err)
	}

	got, err := os.ReadFile(path + ".bak")
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != string(backup) {
		t.Fatalf("backup changed on no-op sync")
	}
}

func TestSyncAuthorizedKeysNoOpDoesNotCreateBackup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "authorized_keys")

	keys := []string{
		"ssh-ed25519 KEY1 macbook-pro",
	}

	data := []byte(`# BEGIN SSHKEYS MANAGED
ssh-ed25519 KEY1 macbook-pro
# END SSHKEYS MANAGED
`)

	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}

	changed, err := syncAuthorizedKeys(path, keys)
	if err != nil {
		t.Fatalf("syncAuthorizedKeys() error = %v", err)
	}

	if changed {
		t.Fatal("syncAuthorizedKeys() changed = true, want false")
	}

	if _, err := os.Stat(path + ".bak"); !os.IsNotExist(err) {
		t.Fatal("backup created on unchanged sync")
	}
}

func TestAuthorizedKeysStatus(t *testing.T) {
	keys := []string{
		"ssh-ed25519 KEY1 macbook-pro",
	}

	t.Run("not found", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "authorized_keys")

		got, err := authorizedKeysStatus(path, keys)
		if err != nil {
			t.Fatal(err)
		}

		if got != "not found" {
			t.Fatalf("status = %q, want %q", got, "not found")
		}
	})

	t.Run("up to date", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "authorized_keys")

		data := []byte(`# BEGIN SSHKEYS MANAGED
ssh-ed25519 KEY1 macbook-pro
# END SSHKEYS MANAGED
`)

		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}

		got, err := authorizedKeysStatus(path, keys)
		if err != nil {
			t.Fatal(err)
		}

		if got != "up to date" {
			t.Fatalf("status = %q, want %q", got, "up to date")
		}
	})

	t.Run("out of date", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "authorized_keys")

		data := []byte(`# BEGIN SSHKEYS MANAGED
ssh-ed25519 OLD old-machine
# END SSHKEYS MANAGED
`)

		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}

		got, err := authorizedKeysStatus(path, keys)
		if err != nil {
			t.Fatal(err)
		}

		if got != "out of date" {
			t.Fatalf("status = %q, want %q", got, "out of date")
		}
	})
}

func TestSyncAuthorizedKeysBackupFailureLeavesOriginalUnchanged(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "authorized_keys")

	initial := []byte(`# BEGIN SSHKEYS MANAGED
ssh-ed25519 OLD old-macbook
# END SSHKEYS MANAGED
`)

	if err := os.WriteFile(path, initial, 0600); err != nil {
		t.Fatal(err)
	}

	// Make the backup destination unusable as a file.
	if err := os.Mkdir(path+".bak", 0700); err != nil {
		t.Fatal(err)
	}

	keys := []string{
		"ssh-ed25519 NEW new-macbook",
	}

	changed, err := syncAuthorizedKeys(path, keys)
	if err == nil {
		t.Fatal("syncAuthorizedKeys() error = nil, want error")
	}

	if changed {
		t.Fatal("syncAuthorizedKeys() changed = true, want false")
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, initial) {
		t.Fatalf("authorized_keys changed after backup failure:\n%s", got)
	}
}

func TestReplaceManagedKeysIgnoresMarkerTextInsideLines(t *testing.T) {
	data := []byte(`# Documentation: # BEGIN SSHKEYS MANAGED
ssh-ed25519 OUTSIDE emergency-access
# Documentation: # END SSHKEYS MANAGED
`)

	keys := []string{
		"ssh-ed25519 KEY1 macbook-pro",
	}

	want := `# Documentation: # BEGIN SSHKEYS MANAGED
ssh-ed25519 OUTSIDE emergency-access
# Documentation: # END SSHKEYS MANAGED

# BEGIN SSHKEYS MANAGED
ssh-ed25519 KEY1 macbook-pro
# END SSHKEYS MANAGED
`

	got, err := replaceManagedKeys(data, keys)
	if err != nil {
		t.Fatalf("replaceManagedKeys() error = %v", err)
	}

	if string(got) != want {
		t.Fatalf("replaceManagedKeys() =\n%s\nwant:\n%s", got, want)
	}
}

func TestSyncAuthorizedKeysCreatesParentDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ssh", "authorized_keys")

	keys := []string{
		"ssh-ed25519 KEY1 macbook-pro",
	}

	changed, err := syncAuthorizedKeys(path, keys)
	if err != nil {
		t.Fatalf("syncAuthorizedKeys() error = %v", err)
	}

	if !changed {
		t.Fatal("syncAuthorizedKeys() changed = false, want true")
	}

	info, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}

	if got := info.Mode().Perm(); got != 0700 {
		t.Fatalf(".ssh mode = %o, want 700", got)
	}

	info, err = os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("authorized_keys mode = %o, want 600", got)
	}

	if _, err := os.Stat(path + ".bak"); !os.IsNotExist(err) {
		t.Fatal("backup should not exist on first creation")
	}
}

func TestSyncAuthorizedKeysRejectsSymlink(t *testing.T) {
	dir := t.TempDir()

	target := filepath.Join(dir, "target")
	path := filepath.Join(dir, "authorized_keys")

	original := []byte(`# BEGIN SSHKEYS MANAGED
ssh-ed25519 OLD old-macbook
# END SSHKEYS MANAGED
`)

	if err := os.WriteFile(target, original, 0600); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}

	keys := []string{
		"ssh-ed25519 NEW new-macbook",
	}

	changed, err := syncAuthorizedKeys(path, keys)
	if err == nil {
		t.Fatal("syncAuthorizedKeys() error = nil, want error")
	}

	if changed {
		t.Fatal("syncAuthorizedKeys() changed = true, want false")
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, original) {
		t.Fatal("symlink target was modified")
	}

	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("authorized_keys symlink was replaced")
	}
}
