package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	managedBegin = "# BEGIN SSHKEYS MANAGED"
	managedEnd   = "# END SSHKEYS MANAGED"
)

func authorizedKeysPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".ssh", "authorized_keys"), nil
}

func replaceManagedKeys(data []byte, keys []string) ([]byte, error) {
	content := string(data)

	begin, end, err := markerOffsets(data)
	if err != nil {
		return nil, err
	}

	if begin == -1 {
		if content != "" && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		if content != "" && !strings.HasSuffix(content, "\n\n") {
			content += "\n"
		}

		return []byte(content + buildManagedSection(keys)), nil
	}

	return []byte(
		content[:begin] +
			strings.TrimSuffix(buildManagedSection(keys), "\n") +
			content[end:],
	), nil
}

func buildManagedSection(keys []string) string {
	var b strings.Builder

	b.WriteString(managedBegin)
	b.WriteByte('\n')

	for _, key := range keys {
		b.WriteString(key)
		b.WriteByte('\n')
	}

	b.WriteString(managedEnd)
	b.WriteByte('\n')

	return b.String()
}

func syncAuthorizedKeys(path string, keys []string) (bool, error) {
	exists, err := checkAuthorizedKeysFile(path)
	if err != nil {
		return false, err
	}

	var data []byte

	if exists {
		data, err = os.ReadFile(path)
		if err != nil {
			return false, fmt.Errorf("read authorized_keys: %w", err)
		}
	}

	updated, err := replaceManagedKeys(data, keys)
	if err != nil {
		return false, fmt.Errorf("replace managed keys: %w", err)
	}

	if exists && bytes.Equal(data, updated) {
		return false, nil
	}

	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0700); err != nil {
		return false, fmt.Errorf("create SSH directory: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".authorized_keys-*")
	if err != nil {
		return false, fmt.Errorf("create temporary file: %w", err)
	}
	tmpPath := tmp.Name()

	defer os.Remove(tmpPath)

	if _, err := tmp.Write(updated); err != nil {
		tmp.Close()
		return false, fmt.Errorf("write temporary file: %w", err)
	}

	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return false, fmt.Errorf("chmod temporary file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return false, fmt.Errorf("close temporary file: %w", err)
	}

	// Replacement is fully prepared. Preserve the current working file.
	if exists {
		if err := atomicWrite(path+".bak", data, 0600); err != nil {
			return false, fmt.Errorf("backup authorized_keys: %w", err)
		}
	}

	// Same-directory rename gives us atomic replacement.
	if err := os.Rename(tmpPath, path); err != nil {
		return false, fmt.Errorf("replace authorized_keys: %w", err)
	}

	return true, nil
}

func atomicWrite(path string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)

	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+"-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()

	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}

	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}
func authorizedKeysStatus(path string, keys []string) (string, error) {
	exists, err := checkAuthorizedKeysFile(path)
	if err != nil {
		return "", err
	}

	if !exists {
		return "not found", nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read authorized_keys: %w", err)
	}

	updated, err := replaceManagedKeys(data, keys)
	if err != nil {
		return "", fmt.Errorf("check managed keys: %w", err)
	}

	if bytes.Equal(data, updated) {
		return "up to date", nil
	}

	return "out of date", nil
}

func markerOffsets(data []byte) (begin, end int, err error) {
	begin, end = -1, -1

	offset := 0
	for line := range strings.SplitSeq(string(data), "\n") {
		lineEnd := offset + len(line)

		switch strings.TrimSuffix(line, "\r") {
		case managedBegin:
			if begin != -1 {
				return -1, -1, fmt.Errorf("managed section malformed: duplicate begin marker")
			}
			begin = offset

		case managedEnd:
			if end != -1 {
				return -1, -1, fmt.Errorf("managed section malformed: duplicate end marker")
			}
			end = lineEnd
		}

		offset = lineEnd + 1
	}

	switch {
	case begin == -1 && end == -1:
		return -1, -1, nil
	case begin == -1 || end == -1:
		return -1, -1, fmt.Errorf("managed section malformed")
	case end < begin:
		return -1, -1, fmt.Errorf("managed section malformed")
	}

	return begin, end, nil
}

func checkAuthorizedKeysFile(path string) (bool, error) {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspect authorized_keys: %w", err)
	}

	if !info.Mode().IsRegular() {
		return false, fmt.Errorf("authorized_keys is not a regular file")
	}

	return true, nil
}
