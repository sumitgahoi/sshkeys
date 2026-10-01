package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: sshkeys <status|sync> <github-user>\n")
		os.Exit(2)
	}

	command := os.Args[1]
	user := os.Args[2]

	switch command {
	case "status":
		keys, err := githubKeys(user)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		path, err := authorizedKeysPath()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		status, err := authorizedKeysStatus(path, keys)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Printf("%d GitHub keys\n", len(keys))
		fmt.Printf("authorized_keys: %s\n", status)
	case "sync":
		keys, err := githubKeys(user)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		path, err := authorizedKeysPath()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		changed, err := syncAuthorizedKeys(path, keys)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if !changed {
			fmt.Println("already up to date")
			return
		}

		fmt.Printf("synced %d keys from GitHub user %s\n", len(keys), user)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		fmt.Fprintln(os.Stderr, "usage: sshkeys <status|sync> <github-user>")
		os.Exit(2)
	}
}
