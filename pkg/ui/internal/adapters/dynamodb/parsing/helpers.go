package parsing

import (
	"fmt"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
)

func spf(format string, a ...any) string {
	return fmt.Sprintf(format, a...)
}

func newLineIf(b bool) string {
	if b {
		return "\n"
	}
	return ""
}

func suffixIf(s string, x string, b bool) string {
	if b {
		return s + x
	}
	return s
}

func prefixIf(s string, p string, b bool) string {
	if b {
		return p + s
	}
	return s
}

func trimPrefixIf(s string, p string, b bool) string {
	if b {
		return strings.TrimPrefix(s, p)
	}
	return s
}
func trimSuffixIf(s string, x string, b bool) string {
	if b {
		return strings.TrimSuffix(s, x)
	}
	return s
}

// TODO: make configurable
var tabsize = 3

func tabs(n int) string {
	var res string
	for range n * tabsize {
		res += spf(" ")
	}
	return res
}

// getSortedKeys returns the dynamo-db item keys as a `[]string` sorted
// alphabetically. If `rootLevel` equals `true`, the hashkey and rangekey (if
// not nil) are extracted from the slice and prefixed at indices 0 and 1,
// respectively.
func getSortedKeys(hashkey string, rangekey *string, elements map[string]types.AttributeValue, rootLevel bool) []string {
	keysSorted := make([]string, 0, len(elements))
	for key := range elements {
		keysSorted = append(keysSorted, key)
	}
	slices.Sort(keysSorted)

	if !rootLevel {
		return keysSorted
	}

	var hidx *int
	var ridx *int

	for i, k := range keysSorted {
		if k == hashkey {
			hidx = &i
			if rangekey == nil {
				break
			}
		}
		if rangekey != nil && k == *rangekey {
			ridx = &i
			if hidx != nil {
				break
			}
		}
	}

	if hidx == nil {
		panic(spf("\nhashkey: %s; keys: %+v\n", hashkey, keysSorted))
	}

	keysSorted = slices.Delete(keysSorted, *hidx, *hidx+1)
	if ridx != nil {
		if *hidx < *ridx { // removed element before ridx; shifting the item ridx points to
			*ridx -= 1
		}
		keysSorted = slices.Delete(keysSorted, *ridx, *ridx+1)
	}

	ret := []string{hashkey}
	if rangekey != nil {
		ret = append(ret, *rangekey)
	}
	return slices.Clip(append(ret, keysSorted...))
}

// flatten takes a string and removes newlines and any spaces that are not
// captured within a double-quoted string. It also removes a trailing comma.
func flatten(in, token string) string {
	lines := strings.Split(in, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimSpace(l)
	}
	str := strings.Join(lines, "")
	return strings.TrimSuffix(str, token)
}

func flattenStyles(multilineStyling []styles.LineStyle) styles.LineStyle {
	if len(multilineStyling) == 0 {
		return styles.LineStyle{}
	}
	if len(multilineStyling) == 1 {
		return multilineStyling[0]
	}
	res := styles.LineStyle{}.AppendLines(multilineStyling)
	return res.UnsetPaddingAll()
}
