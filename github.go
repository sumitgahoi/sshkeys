package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func fetchGitHubKeys(user string) ([]byte, error) {
	if err := validateGitHubUser(user); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://github.com/%s.keys", user)

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch GitHub keys: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch GitHub keys: %s", resp.Status)
	}

	keys, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read GitHub keys: %w", err)
	}

	return keys, nil
}

func githubKeys(user string) ([]string, error) {
	data, err := fetchGitHubKeys(user)
	if err != nil {
		return nil, err
	}

	keys := parseKeys(data)

	if err := validateKeys(keys); err != nil {
		return nil, err
	}

	return keys, nil
}

func validateGitHubUser(user string) error {
	if user == "" || len(user) > 39 {
		return fmt.Errorf("invalid GitHub username: %q", user)
	}

	for i, r := range user {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') {
			continue
		}

		if r == '-' && i > 0 && i < len(user)-1 {
			continue
		}

		return fmt.Errorf("invalid GitHub username: %q", user)
	}

	if strings.Contains(user, "--") {
		return fmt.Errorf("invalid GitHub username: %q", user)
	}

	return nil
}
