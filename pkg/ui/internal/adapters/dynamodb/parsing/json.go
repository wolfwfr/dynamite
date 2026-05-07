package parsing

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	apitypes "github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
	"github.com/wolfwfr/dynamite/pkg/util"
)

const (
	jsonFmt = "%s%s\n"
)

type JSONParser struct {
	Styles jsonParserStyles
}

type jsonParserStyles struct {
	FieldNameStyle lipgloss.Style
	NumberStyle    lipgloss.Style
	BoolStyle      lipgloss.Style
	BytesStyle     lipgloss.Style
	NULLStyle      lipgloss.Style
	StringStyle    lipgloss.Style
	TokenStyle     lipgloss.Style
	ErrorStyle     lipgloss.Style
}

func NewJSONParser() JSONParser {
	p := JSONParser{}
	p.Styles.FieldNameStyle = lipgloss.NewStyle().Foreground(styles.SubtleColour)
	p.Styles.NumberStyle = lipgloss.NewStyle().Foreground(styles.NumberColour)
	p.Styles.BoolStyle = lipgloss.NewStyle().Foreground(styles.BoolColour)
	p.Styles.BytesStyle = lipgloss.NewStyle().Foreground(styles.BytesColour)
	p.Styles.NULLStyle = lipgloss.NewStyle().Foreground(styles.NULLColour)
	p.Styles.StringStyle = lipgloss.NewStyle().Foreground(styles.StringColour)
	p.Styles.TokenStyle = lipgloss.NewStyle().Foreground(styles.TokenColour)
	p.Styles.ErrorStyle = lipgloss.NewStyle().Foreground(styles.ErrorColour)
	return p
}

func (p JSONParser) ParseToJSONWithKeys(item map[string]types.AttributeValue, hashkey string, rangekey *string) (string, string, []apitypes.KeyValue) {
	token := p.Styles.TokenStyle.Render
	json, styled, keyValues := p.pJSON(item, hashkey, rangekey, 0)
	trim := func(s, token string) string { return strings.TrimSuffix(s, spf("%s\n", token)) }
	return trim(json, ","), trim(styled, token(",")), keyValues
}

func (p JSONParser) ParseItemToJSON(item map[string]types.AttributeValue, hashkey string, rangekey *string) (string, string) {
	token := p.Styles.TokenStyle.Render
	json, styled, _ := p.pJSON(item, hashkey, rangekey, 0)
	trim := func(s, token string) string { return strings.TrimSuffix(s, spf("%s\n", token)) }
	return trim(json, ","), trim(styled, token(","))
}

// nestLevel determines indentations pJSON is an internal, recursive function
// that takes a dynamo-db item and parses it to a json-formatted string.
// TODO: consider elegant way of separating json-parsing from string->string
// key-value mapping, but for now this saves double work and the two are always
// used together.
func (p JSONParser) pJSON(elements map[string]types.AttributeValue, hashkey string, rangekey *string, nestLevel int) (string, string, []apitypes.KeyValue) {
	json, styled := strings.Builder{}, strings.Builder{}

	// obtain sorted keys
	keysSorted := getSortedKeys(hashkey, rangekey, elements, nestLevel == 0)

	nestLevel += 1

	isRootLevel := nestLevel == 1
	var kv []apitypes.KeyValue
	if isRootLevel {
		kv = make([]apitypes.KeyValue, len(keysSorted))
	}

	token := p.Styles.TokenStyle.Render
	field := p.Styles.FieldNameStyle.Render

	hasContent := len(keysSorted) > 0

	// write prefix token
	prefix := func(token string) string { return spf("%s%s", token, newLineIf(hasContent)) }
	json.WriteString(prefix("{"))
	styled.WriteString(prefix(token("{")))

	for i, k := range keysSorted {
		v := elements[k]
		isLast := i == len(keysSorted)-1

		// write field-name
		fieldName := func(quotedName, colon string) string {
			return spf("%s%s%s ", tabs(nestLevel), quotedName, colon)
		}
		quotedName := spf("\"%s\"", k)
		json.WriteString(fieldName(quotedName, ":"))
		styled.WriteString(fieldName(field(quotedName), token(":")))

		// obtain block content
		content, styledContent := p.switchAttrValueJSON(v, hashkey, rangekey, nestLevel)
		if isRootLevel {
			kv[i] = apitypes.KeyValue{Key: k, Value: flatten(content)}
		}

		// write comma & newline, unless last element
		withSuffix := func(s, comma string) string {
			return spf("%s", suffixIf(trimSuffixIf(s, spf("%s\n", comma), isLast), "\n", isLast)) // no trailing commas
		}
		json.WriteString(withSuffix(content, ","))
		styled.WriteString(withSuffix(styledContent, token(",")))
	}

	//write suffix tokens
	suffix := func(token, comma string) string {
		return spf("%s%s%s\n", prefixIf("", tabs(nestLevel-1), hasContent), token, comma)
	}
	json.WriteString(suffix("}", ","))
	styled.WriteString(suffix(token("}"), token(",")))

	return json.String(), styled.String(), kv
}

func (p JSONParser) switchAttrValueJSON(v types.AttributeValue, hashkey string, rangekey *string, nestLevel int) (string, string) {
	str := p.Styles.StringStyle.Render
	num := p.Styles.NumberStyle.Render
	bol := p.Styles.BoolStyle.Render
	byt := p.Styles.BytesStyle.Render
	tok := p.Styles.TokenStyle.Render
	nul := p.Styles.NULLStyle.Render
	err := p.Styles.ErrorStyle.Render

	switch vv := v.(type) {
	case *types.AttributeValueMemberB:
		return pJSONBytes(vv.Value, tok(","), byt)
	case *types.AttributeValueMemberBOOL:
		return pJSONBool(vv.Value, tok(","), bol)
	case *types.AttributeValueMemberBS:
		return stringableAsListJSON(p.Styles, vv.Value, nestLevel, func(s []byte) (string, string) { return pJSONBytes(s, tok(","), byt) })
	case *types.AttributeValueMemberL:
		return stringableAsListJSON(p.Styles, vv.Value, nestLevel, func(s types.AttributeValue) (string, string) {
			return p.switchAttrValueJSON(s, hashkey, rangekey, nestLevel+1)
		})
	case *types.AttributeValueMemberM:
		raw, styled := strings.Builder{}, strings.Builder{}
		str, st, _ := p.pJSON(vv.Value, hashkey, rangekey, nestLevel)
		fmt.Fprintf(&raw, "%s", str)
		fmt.Fprintf(&styled, "%s", st)
		return raw.String(), styled.String()
	case *types.AttributeValueMemberN:
		return pJSONNum(vv.Value, tok(","), num)
	case *types.AttributeValueMemberNS:
		return stringableAsListJSON(p.Styles, vv.Value, nestLevel, func(s string) (string, string) { return pJSONNum(s, tok(","), num) })
	case *types.AttributeValueMemberNULL: // TODO: ignore?
		v := util.Ternary("NULL", "NOT NULL", vv.Value)
		return spf(jsonFmt, v, ","), spf(jsonFmt, nul(v), tok(","))
	case *types.AttributeValueMemberS:
		return pJSONString(vv.Value, tok(","), str)
	case *types.AttributeValueMemberSS:
		return stringableAsListJSON(p.Styles, vv.Value, nestLevel, func(s string) (string, string) { return pJSONString(s, tok(","), str) })
	default:
		fm := "<failed to parse>"
		return spf(jsonFmt, fm, ","), spf(jsonFmt, err(fm), tok(",")) // TODO: error?
	}
}

func stringableAsListJSON[S []E, E any](styles jsonParserStyles, s S, nestLevel int, tr func(E) (string, string)) (string, string) {
	token := styles.TokenStyle.Render

	json := strings.Builder{}
	styled := strings.Builder{}

	hasContent := len(s) > 0

	prefix := func(token string) string { return spf("%s%s", token, newLineIf(hasContent)) }
	json.WriteString(prefix("["))
	styled.WriteString(prefix(token("[")))

	line := func(in, token string, atEnd bool) string {
		return spf("%s%s", tabs(nestLevel+1), suffixIf(trimSuffixIf(in, spf("%s\n", token), atEnd), "\n", atEnd)) // no trailing commas
	}
	for i, v := range s {
		j, st := tr(v)
		json.WriteString(line(j, ",", i == len(s)-1))
		styled.WriteString(line(st, token(","), i == len(s)-1))
	}

	suffix := func(token, comma string) string {
		return spf("%s%s%s\n", prefixIf("", tabs(nestLevel), hasContent), token, comma)
	}
	json.WriteString(suffix("]", ","))
	styled.WriteString(suffix(token("]"), token(",")))

	return json.String(), styled.String()
}

func pJSONBool(bl bool, renderedToken string, render func(...string) string) (string, string) {
	b := spf("%t", bl)
	return spf(jsonFmt, b, ","), spf(jsonFmt, render(b), renderedToken)
}

func pJSONBytes(bt []byte, renderedToken string, render func(...string) string) (string, string) {
	bytesFmt := "<bytes>(len=%d)"
	b := spf(bytesFmt, len(bt))
	return spf(jsonFmt, b, ","), spf(jsonFmt, render(b), renderedToken)
}

func pJSONString(str, renderedToken string, render func(...string) string) (string, string) {
	s := spf("%q", str)
	return spf(jsonFmt, s, ","), spf(jsonFmt, render(s), renderedToken)
}

func pJSONNum(num, renderedToken string, render func(...string) string) (string, string) {
	n := spf("%s", num)
	return spf(jsonFmt, n, ","), spf(jsonFmt, render(n), renderedToken)
}
