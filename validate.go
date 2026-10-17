package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func validateKeys(keys []string) error {
	for _, key := range keys {
		cmd := exec.Command("ssh-keygen", "-l", "-f", "-")
		cmd.Stdin = strings.NewReader(key + "\n")

		output, err := cmd.CombinedOutput()
		if err != nil {
			message := strings.TrimSpace(string(output))
			if message != "" {
				return fmt.Errorf("validate SSH key: %w: %s", err, message)
			}

			return fmt.Errorf("validate SSH key: %w", err)
		}
	}

	return nil
}
