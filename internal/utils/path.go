package utils

import (
	"net/url"
	"strings"
)

func EscapePathPreservingSlashes(p string) string {
	parts := strings.Split(p, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

func JoinURLPath(prefix, suffix string) string {
	prefix = strings.TrimRight(prefix, "/")
	suffix = strings.TrimLeft(suffix, "/")
	return prefix + "/" + suffix
}
