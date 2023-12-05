package utils

import (
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
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

func GenerateRandomString(length int) string {
	charset := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	//nolint: gosec
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	random := make([]byte, length)
	for i := range random {
		random[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(random)
}

func ToPulumiMap(m []string, separator string) pulumi.StringMap {
	pulumiMap := pulumi.StringMap{}

	for _, s := range m {
		k := strings.Split(s, separator)[0]
		v := strings.Split(s, separator)[1]
		pulumiMap[k] = pulumi.String(v)
	}

	return pulumiMap
}
