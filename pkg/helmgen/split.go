package helmgen

import "strings"

func GetResources(data []byte) []string {
	if len(data) == 0 {
		return []string{}
	}

	parts := strings.Split(string(data), "---\n")
	return parts
}
