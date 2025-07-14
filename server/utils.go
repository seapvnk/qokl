package server

import "strings"

func buildRoutePath(parts []string) string {
	for i, p := range parts {
		if strings.HasPrefix(p, ":") {
			parts[i] = "{" + p[1:len(p)] + "}"
		}
	}

	path := strings.Join(parts, "/")
	return "/" + path
}

