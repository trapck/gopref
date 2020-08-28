package fs

import "strings"

// PathToParts splits path to parts
func PathToParts(path string) []string {
	return strings.Split(strings.ReplaceAll(path, "\"", ""), "/")
}

// LastPart takes the last part of the path
func LastPart(path string) string {
	parts := PathToParts(path)
	return parts[len(parts)-1]
}

// LastPartExclude excludes the last part of the path
func LastPartExclude(path string) string {
	parts := PathToParts(path)
	return strings.Join(parts[:len(parts)-1], "/")
}
