package parsing

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	apitypes "github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
)

type JSONParser struct {
	Styles jsonParserStyles
}

type jsonParserStyles struct {
	FieldNameStyle lipgloss.Style
	NumberStyle    lipgloss.Style
	StringStyle    lipgloss.Style
	TokenStyle     lipgloss.Style
}

func NewJSONParser() JSONParser {
	p := JSONParser{}
	p.Styles.FieldNameStyle = lipgloss.NewStyle().Foreground(styles.SubtleColour)
	p.Styles.NumberStyle = lipgloss.NewStyle().Foreground(styles.NumberColour)
	p.Styles.StringStyle = lipgloss.NewStyle().Foreground(styles.StringColour)
	p.Styles.TokenStyle = lipgloss.NewStyle().Foreground(styles.TokenColour)
	return p
}

func (p JSONParser) ParseToJSONWithKeys(item map[string]types.AttributeValue, hashkey string, rangekey *string) (string, string, []apitypes.KeyValue) {
	token := p.Styles.TokenStyle.Render
	json, styled, keyValues := p.pJSON(item, hashkey, rangekey, 0)
	return strings.TrimSuffix(json, ",\n"), strings.TrimSuffix(styled, fmt.Sprintf("%s\n", token(","))), keyValues
}

func (p JSONParser) ParseItemToJSON(item map[string]types.AttributeValue, hashkey string, rangekey *string) (string, string) {
	token := p.Styles.TokenStyle.Render
	json, styled, _ := p.pJSON(item, hashkey, rangekey, 0)
	return strings.TrimSuffix(json, ",\n"), strings.TrimSuffix(styled, fmt.Sprintf("%s\n", token(","))) // no trailing commas
}

// nestLevel determines indentations pJSON is an internal, recursive function
// that takes a dynamo-db item and parses it to a json-formatted string.
// TODO: consider elegant way of separating json-parsing from string->string
// key-value mapping, but for now this saves double work and the two are always
// used together.
func (p JSONParser) pJSON(elements map[string]types.AttributeValue, hashkey string, rangekey *string, nestLevel int) (string, string, []apitypes.KeyValue) {
	json := strings.Builder{}
	styled := strings.Builder{}

	// obtain sorted keys
	keysSorted := getSortedKeys(hashkey, rangekey, elements, nestLevel == 0)

	nestLevel += 1

	isRootLevel := nestLevel == 1
	var kv []apitypes.KeyValue
	if isRootLevel {
		kv = make([]apitypes.KeyValue, len(keysSorted))
	}

	token := p.Styles.TokenStyle.Render
	fieldName := p.Styles.FieldNameStyle.Render

	hasContent := len(keysSorted) > 0
	fmt.Fprintf(&json, "{%s", newLineIf(hasContent))                // opening '{'
	fmt.Fprintf(&styled, "%s%s", token("{"), newLineIf(hasContent)) // opening '{'
	for i, k := range keysSorted {
		v := elements[k]
		isLast := i == len(keysSorted)-1

		fmt.Fprintf(&json, "%s\"%s\": ", tabs(nestLevel), k)                                              // write key
		fmt.Fprintf(&styled, "%s%s%s ", tabs(nestLevel), fieldName(fmt.Sprintf("\"%s\"", k)), token(":")) // write key

		content, styledContent := p.switchAttrValueJSON(v, hashkey, rangekey, nestLevel)
		if isRootLevel {
			kv[i] = apitypes.KeyValue{Key: k, Value: flatten(content)}
		}

		fmt.Fprintf(&json, "%s", suffixIf(trimSuffixIf(content, ",\n", isLast), "\n", isLast))                                   // no trailing commas
		fmt.Fprintf(&styled, "%s", suffixIf(trimSuffixIf(styledContent, fmt.Sprintf("%s\n", token(",")), isLast), "\n", isLast)) // no trailing commas
	}
	fmt.Fprintf(&json, "%s},\n", prefixIf("", tabs(nestLevel-1), hasContent))                             // closing '}'
	fmt.Fprintf(&styled, "%s%s%s\n", prefixIf("", tabs(nestLevel-1), hasContent), token("}"), token(",")) // closing '}'

	return json.String(), styled.String(), kv
}

func (p JSONParser) switchAttrValueJSON(v types.AttributeValue, hashkey string, rangekey *string, nestLevel int) (string, string) {
	str := p.Styles.StringStyle.Render
	num := p.Styles.NumberStyle.Render
	token := p.Styles.TokenStyle.Render

	switch vv := v.(type) {
	case *types.AttributeValueMemberB:
		return twice(fmt.Sprintf("<bytes>(len=%d),\n", len(vv.Value)))
	case *types.AttributeValueMemberBOOL:
		return twice(fmt.Sprintf("%t,\n", vv.Value))
	case *types.AttributeValueMemberBS:
		return stringableAsListJSON(p.Styles, vv.Value, nestLevel, func(s []byte) (string, string) { return twice(fmt.Sprintf("<bytes>(len=%d),\n", len(s))) })
	case *types.AttributeValueMemberL:
		return stringableAsListJSON(p.Styles, vv.Value, nestLevel, func(s types.AttributeValue) (string, string) {
			return p.switchAttrValueJSON(s, hashkey, rangekey, nestLevel+1)
		})
	case *types.AttributeValueMemberM:
		j := strings.Builder{}
		s := strings.Builder{}
		str, st, _ := p.pJSON(vv.Value, hashkey, rangekey, nestLevel)
		fmt.Fprintf(&j, "%s", str)
		fmt.Fprintf(&s, "%s", st)
		return j.String(), s.String()
	case *types.AttributeValueMemberN:
		return fmt.Sprintf("%s,\n", vv.Value), fmt.Sprintf("%s%s\n", num(vv.Value), token(","))
	case *types.AttributeValueMemberNS:
		return stringableAsListJSON(p.Styles, vv.Value, nestLevel, func(s string) (string, string) {
			return fmt.Sprintf("%s,\n", s), fmt.Sprintf("%s%s\n", num(s), token(","))
		})
	case *types.AttributeValueMemberNULL: // TODO: ignore?
		if vv.Value {
			return twice("NULL,\n")
		}
		return twice("NOT NULL,\n")
	case *types.AttributeValueMemberS:
		return fmt.Sprintf("%q,\n", vv.Value), fmt.Sprintf("%s%s\n", str(fmt.Sprintf("%q", vv.Value)), token(","))
	case *types.AttributeValueMemberSS:
		return stringableAsListJSON(p.Styles, vv.Value, nestLevel, func(s string) (string, string) {
			return fmt.Sprintf("%q,\n", s), fmt.Sprintf("%s%s\n", str(fmt.Sprintf("%q", s)), token(","))
		})
	default:
		return twice("<failed to parse>,\n") // TODO: error?
	}
}

func twice[T any](t T) (T, T) {
	return t, t
}

func stringableAsListJSON[S []E, E any](styles jsonParserStyles, s S, nestLevel int, tr func(E) (string, string)) (string, string) {
	token := styles.TokenStyle.Render
	json := strings.Builder{}
	styled := strings.Builder{}
	hasContent := len(s) > 0
	fmt.Fprintf(&json, "[%s", newLineIf(hasContent))
	fmt.Fprintf(&styled, "%s%s", token("["), newLineIf(hasContent))
	for i, v := range s {
		j, st := tr(v)
		fmt.Fprintf(&json, "%s%s", tabs(nestLevel+1), suffixIf(trimSuffixIf(j, ",\n", i == len(s)-1), "\n", i == len(s)-1))                              // no trailing commas
		fmt.Fprintf(&styled, "%s%s", tabs(nestLevel+1), suffixIf(trimSuffixIf(st, fmt.Sprintf("%s\n", token(",")), i == len(s)-1), "\n", i == len(s)-1)) // no trailing commas
	}
	fmt.Fprintf(&json, "%s],\n", prefixIf("", tabs(nestLevel), hasContent))
	fmt.Fprintf(&styled, "%s%s\n", prefixIf("", tabs(nestLevel), hasContent), token("],"))
	return json.String(), styled.String()
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
