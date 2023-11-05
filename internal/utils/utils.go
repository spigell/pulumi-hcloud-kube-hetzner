package utils

import (
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"sort"
)

func SortedMapKeys[V any](unsorted map[string]V) []string {
	keys := make([]string, 0, len(unsorted))

	for k := range unsorted {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func ExtractRemoteCommandResources(resources []pulumi.Resource) pulumi.Array {
	var res pulumi.Array
	for _, r := range resources {
		if r == nil {
			continue
		}
		c, ok := r.(*remote.Command)
		if !ok {
			continue
		}

		res = append(res, c.Connection)
		res = append(res, c.Create)
	}
	return res
}
