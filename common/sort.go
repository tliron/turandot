package common

import (
	"sort"
)

func SortedMapStringStringKeys(map_ map[string]string) []string {
	keys := make([]string, 0, len(map_))
	for key := range map_ {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
