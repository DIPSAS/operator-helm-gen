package helmgen

import (
	"fmt"
	"strings"
)

func GetResources(data []byte) []string {
	if len(data) == 0 {
		return []string{}
	}

	parts := strings.Split(string(data), "---\n")
	return parts
}

func GetResourcesToKeep(arg string) map[string]string {
	k := make(map[string]string)
	for _, x := range strings.Split(arg, ",") {
		k[x] = "keep"
	}

	return k
}

func GetAnnotationsFromArg(arg string) (map[string]string, error) {
	annotations := make(map[string]string)
	for _, a := range strings.Split(arg, ",") {
		pair := strings.Split(a, "=")
		if len(pair) != 2 {
			return nil, fmt.Errorf("the parameter value \"%s\" is not a valid annotation", a)
		}
		annotations[pair[0]] = pair[1]
	}

	return annotations, nil
}
