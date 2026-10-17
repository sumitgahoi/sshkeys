package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseKeys(t *testing.T) {
	data := []byte(`
# comment

ssh-ed25519 KEY1 macbook-pro
ssh-ed25519 KEY2 macbook-air

`)

	got := parseKeys(data)

	want := []string{
		"ssh-ed25519 KEY1 macbook-pro",
		"ssh-ed25519 KEY2 macbook-air",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseKeys() = %q, want %q", got, want)
	}
}

func TestValidateKeysAllowsEmpty(t *testing.T) {
	if err := validateKeys(parseKeys(nil)); err != nil {
		t.Fatalf("validateKeys() error = %v", err)
	}
}

func TestValidateKeysRejectsInvalidKey(t *testing.T) {
	err := validateKeys([]string{
		"ssh-ed25519 definitely-not-a-valid-key test",
	})

	if err == nil {
		t.Fatal("validateKeys() error = nil, want error")
	}
}

func TestValidateKeysAcceptsValidKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "id_ed25519")

	cmd := exec.Command(
		"ssh-keygen",
		"-q",
		"-t", "ed25519",
		"-N", "",
		"-f", path,
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("ssh-keygen: %v: %s", err, output)
	}

	key, err := os.ReadFile(path + ".pub")
	if err != nil {
		t.Fatal(err)
	}

	if err := validateKeys(parseKeys(key)); err != nil {
		t.Fatalf("validateKeys() error = %v", err)
	}
}
