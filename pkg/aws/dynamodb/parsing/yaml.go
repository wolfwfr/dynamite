package parsing

import (
	"fmt"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func ParseItemToYAML(item map[string]types.AttributeValue, hashkey string, rangekey *string) string {
	return pYAML(item, hashkey, rangekey, 0)
}

// nestLevel determines indentations pYAML is an internal, recursive function
// that takes a dynamo-db item and parses it to a yaml-formatted string.
func pYAML(elements map[string]types.AttributeValue, hashkey string, rangekey *string, nestLevel int) string {
	b := strings.Builder{}

	// obtain sorted keys
	keysSorted := getSortedKeys(hashkey, rangekey, elements, nestLevel == 0)

	for _, k := range keysSorted {
		fmt.Fprintf(&b, "%s%s: ", tabs(nestLevel), k)
		v := elements[k]
		b.WriteString(switchAttrValueYAML(v, hashkey, rangekey, nestLevel, false))
	}

	return b.String()
}

func switchAttrValueYAML(v types.AttributeValue, hashkey string, rangekey *string, nestLevel int, isListItem bool) string {
	switch vv := v.(type) {
	case *types.AttributeValueMemberB:
		return fmt.Sprintf("%s\n", string(vv.Value))
	case *types.AttributeValueMemberBOOL:
		return fmt.Sprintf("%t\n", vv.Value)
	case *types.AttributeValueMemberBS:
		return stringableAsListYAML(vv.Value, nestLevel, func(s []byte) string { return fmt.Sprintf("%s\n", s) })
	case *types.AttributeValueMemberL:
		return stringableAsListYAML(vv.Value, nestLevel, func(s types.AttributeValue) string { return switchAttrValueYAML(s, hashkey, rangekey, nestLevel, true) })
	case *types.AttributeValueMemberM:
		b := strings.Builder{}
		str := pYAML(vv.Value, hashkey, rangekey, nestLevel+1)
		if isListItem {
			str = strings.TrimSuffix(strings.ReplaceAll(str, "\n", "\n  "), "  ")
		}
		fmt.Fprintf(&b, "%s%s%s", newLineIf(!isListItem && str != ""), trimPrefixIf(str, tabs(nestLevel+1), isListItem), newLineIf(str == ""))
		return b.String()
	case *types.AttributeValueMemberN:
		return fmt.Sprintf("%s\n", vv.Value)
	case *types.AttributeValueMemberNS:
		return stringableAsListYAML(vv.Value, nestLevel, func(s string) string { return fmt.Sprintf("%s\n", s) })
	case *types.AttributeValueMemberNULL: // TODO: ignore?
		if vv.Value {
			return "NULL\n"
		}
		return "NOT NULL\n"
	case *types.AttributeValueMemberS:
		return fmt.Sprintf("\"%s\"\n", vv.Value)
	case *types.AttributeValueMemberSS:
		return stringableAsListYAML(vv.Value, nestLevel, func(s string) string { return fmt.Sprintf("\"%s\"\n", s) })
	default:
		return "<failed to parse>\n" // TODO: error?
	}
}

func stringableAsListYAML[S []E, E any](s S, nestLevel int, tr func(E) string) string {
	b := strings.Builder{}
	b.WriteString("\n")
	for _, v := range s {
		fmt.Fprintf(&b, "%s- %s", tabs(nestLevel+1), tr(v))
	}
	return b.String()
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
		res += fmt.Sprintf(" ")
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
		panic(fmt.Sprintf("\nhashkey: %s; keys: %+v\n", hashkey, keysSorted))
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
