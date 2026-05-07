package parsing

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	apitypes "github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"
)

func ParseToJSONWithKeys(item map[string]types.AttributeValue, hashkey string, rangekey *string) (string, []apitypes.KeyValue) {
	json, keyValues := pJSON(item, hashkey, rangekey, 0)
	return strings.TrimSuffix(json, ",\n"), keyValues
}

func ParseItemToJSON(item map[string]types.AttributeValue, hashkey string, rangekey *string) string {
	json, _ := pJSON(item, hashkey, rangekey, 0)
	return strings.TrimSuffix(json, ",\n") // no trailing commas
}

// nestLevel determines indentations pJSON is an internal, recursive function
// that takes a dynamo-db item and parses it to a json-formatted string.
// TODO: consider elegant way of separating json-parsing from string->string
// key-value mapping, but for now this saves double work and the two are always
// used together.
func pJSON(elements map[string]types.AttributeValue, hashkey string, rangekey *string, nestLevel int) (string, []apitypes.KeyValue) {
	b := strings.Builder{}

	// obtain sorted keys
	keysSorted := getSortedKeys(hashkey, rangekey, elements, nestLevel == 0)

	nestLevel += 1

	isRootLevel := nestLevel == 1
	var kv []apitypes.KeyValue
	if isRootLevel {
		kv = make([]apitypes.KeyValue, len(keysSorted))
	}

	hasContent := len(keysSorted) > 0
	fmt.Fprintf(&b, "{%s", newLineIf(hasContent)) // opening '{'
	for i, k := range keysSorted {
		v := elements[k]
		isLast := i == len(keysSorted)-1

		fmt.Fprintf(&b, "%s\"%s\": ", tabs(nestLevel), k) // write key

		content := switchAttrValueJSON(v, hashkey, rangekey, nestLevel)
		if isRootLevel {
			kv[i] = apitypes.KeyValue{Key: k, Value: flatten(content)}
		}

		fmt.Fprintf(&b, "%s", suffixIf(trimSuffixIf(content, ",\n", isLast), "\n", isLast)) // no trailing commas
	}
	fmt.Fprintf(&b, "%s},\n", prefixIf("", tabs(nestLevel-1), hasContent)) // closing '}'

	return b.String(), kv
}

func switchAttrValueJSON(v types.AttributeValue, hashkey string, rangekey *string, nestLevel int) string {
	switch vv := v.(type) {
	case *types.AttributeValueMemberB:
		return fmt.Sprintf("<bytes>(len=%d),\n", len(vv.Value))
	case *types.AttributeValueMemberBOOL:
		return fmt.Sprintf("%t,\n", vv.Value)
	case *types.AttributeValueMemberBS:
		return stringableAsListJSON(vv.Value, nestLevel, func(s []byte) string { return fmt.Sprintf("<bytes>(len=%d),\n", len(s)) })
	case *types.AttributeValueMemberL:
		return stringableAsListJSON(vv.Value, nestLevel, func(s types.AttributeValue) string { return switchAttrValueJSON(s, hashkey, rangekey, nestLevel+1) })
	case *types.AttributeValueMemberM:
		b := strings.Builder{}
		str, _ := pJSON(vv.Value, hashkey, rangekey, nestLevel)
		fmt.Fprintf(&b, "%s", str)
		return b.String()
	case *types.AttributeValueMemberN:
		return fmt.Sprintf("%s,\n", vv.Value)
	case *types.AttributeValueMemberNS:
		return stringableAsListJSON(vv.Value, nestLevel, func(s string) string { return fmt.Sprintf("%s,\n", s) })
	case *types.AttributeValueMemberNULL: // TODO: ignore?
		if vv.Value {
			return "NULL,\n"
		}
		return "NOT NULL,\n"
	case *types.AttributeValueMemberS:
		return fmt.Sprintf("%q,\n", vv.Value)
	case *types.AttributeValueMemberSS:
		return stringableAsListJSON(vv.Value, nestLevel, func(s string) string { return fmt.Sprintf("%q,\n", s) })
	default:
		return "<failed to parse>,\n" // TODO: error?
	}
}

func stringableAsListJSON[S []E, E any](s S, nestLevel int, tr func(E) string) string {
	b := strings.Builder{}
	hasContent := len(s) > 0
	fmt.Fprintf(&b, "[%s", newLineIf(hasContent))
	for i, v := range s {
		fmt.Fprintf(&b, "%s%s", tabs(nestLevel+1), suffixIf(trimSuffixIf(tr(v), ",\n", i == len(s)-1), "\n", i == len(s)-1)) // no trailing commas
	}
	fmt.Fprintf(&b, "%s],\n", prefixIf("", tabs(nestLevel), hasContent))
	return b.String()
}

// flatten takes a string and removes newlines and any spaces that are not
// captured within a double-quoted string. It also removes a trailing comma.
func flatten(in string) string {
	str := strings.ReplaceAll(in, "\n", "")
	looking := true
	b := strings.Builder{}
	for _, r := range str {
		if r == '"' {
			looking = !looking
		}
		if !looking || r != ' ' {
			b.WriteRune(r)
		}
	}
	return strings.TrimSuffix(b.String(), ",")
}
