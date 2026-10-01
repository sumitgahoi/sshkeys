package main

import "strings"

func parseKeys(data []byte) []string {
	var keys []string

	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		keys = append(keys, line)
	}

	return keys
}
