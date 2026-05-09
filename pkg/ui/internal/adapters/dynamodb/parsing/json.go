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

// TODO: clean up (whole file)
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
	p.Styles.FieldNameStyle = lipgloss.NewStyle().Foreground(styles.FieldNameColour)
	p.Styles.NumberStyle = lipgloss.NewStyle().Foreground(styles.NumberColour)
	p.Styles.BoolStyle = lipgloss.NewStyle().Foreground(styles.BoolColour)
	p.Styles.BytesStyle = lipgloss.NewStyle().Foreground(styles.BytesColour)
	p.Styles.NULLStyle = lipgloss.NewStyle().Foreground(styles.NULLColour)
	p.Styles.StringStyle = lipgloss.NewStyle().Foreground(styles.StringColour)
	p.Styles.TokenStyle = lipgloss.NewStyle().Foreground(styles.TokenColour)
	p.Styles.ErrorStyle = lipgloss.NewStyle().Foreground(styles.ErrorColour)
	return p
}

func (p JSONParser) ParseToJSONWithKeys(item map[string]types.AttributeValue, hashkey string, rangekey *string) (string, []styles.JSONLineStyling, []apitypes.KeyValue) {
	json, styled, keyValues := p.pJSON(item, hashkey, rangekey, 0)
	trim := func(s, token string) string { return strings.TrimSuffix(s, spf("%s\n", token)) }
	if len(styled) > 0 {
		styled[len(styled)-1] = styled[len(styled)-1].TrimEnd(1)
	}
	return trim(json, ","), styled, keyValues
}

func (p JSONParser) ParseItemToJSON(item map[string]types.AttributeValue, hashkey string, rangekey *string) (string, []styles.JSONLineStyling) {
	json, styled, _ := p.pJSON(item, hashkey, rangekey, 0)
	trim := func(s, token string) string { return strings.TrimSuffix(s, spf("%s\n", token)) }
	if len(styled) > 0 {
		styled[len(styled)-1] = styled[len(styled)-1].TrimEnd(1)
	}
	return trim(json, ","), styled
}

// nestLevel determines indentations pJSON is an internal, recursive function
// that takes a dynamo-db item and parses it to a json-formatted string.
// TODO: consider elegant way of separating json-parsing from string->string
// key-value mapping, but for now this saves double work and the two are always
// used together.
func (p JSONParser) pJSON(elements map[string]types.AttributeValue, hashkey string, rangekey *string, nestLevel int) (string, []styles.JSONLineStyling, []apitypes.KeyValue) {
	raw := strings.Builder{}
	styled := []styles.JSONLineStyling{}

	// obtain sorted keys
	keysSorted := getSortedKeys(hashkey, rangekey, elements, nestLevel == 0)

	nestLevel += 1

	isRootLevel := nestLevel == 1
	var kv []apitypes.KeyValue
	if isRootLevel {
		kv = make([]apitypes.KeyValue, len(keysSorted))
	}

	tokenSt := p.Styles.TokenStyle
	fieldSt := p.Styles.FieldNameStyle

	if len(keysSorted) == 0 { // no content
		raw, styled := emptyBrackets("{}", tokenSt)
		return raw, styled, kv
	}

	// write prefix token
	prefix := func(token string) string { return spf("%s\n", token) }
	raw.WriteString(prefix("{"))
	styled = append(styled, styles.JSONLineStyling{}.AppendRuneStyleLG((tokenSt)))

	for i, k := range keysSorted {
		v := elements[k]
		isLast := i == len(keysSorted)-1

		// write field-name
		fieldName := func(quotedName, colon string) string {
			return spf("%s%s%s ", tabs(nestLevel), quotedName, colon)
		}
		quotedName := spf("\"%s\"", k)
		raw.WriteString(fieldName(quotedName, ":"))

		styled = append(styled, styles.JSONLineStyling{}.AppendStringStyleLG(quotedName, fieldSt, styles.
			WithStringInitialPadding(len(tabs(nestLevel)))).
			AppendRuneStyleLG(tokenSt). // :
			AppendRuneStyleLG(tokenSt), // _ (space)
		)

		// obtain block content
		content, styledContent := p.switchAttrValueJSON(v, hashkey, rangekey, nestLevel)

		// prepare table keys
		if isRootLevel {
			kv[i] = apitypes.KeyValue{Key: k, Value: flatten(content, ","), ValueStyling: flattenStyles(styledContent).TrimEnd(1)}
		}

		// write comma & newline, unless last element
		withSuffix := func(s, comma string) string {
			return spf("%s", suffixIf(trimSuffixIf(s, spf("%s\n", comma), isLast), "\n", isLast)) // if last, replace "<comma>\n" with "\n"
		}
		raw.WriteString(withSuffix(content, ","))
		styled[len(styled)-1] = styled[len(styled)-1].AppendLineStyle(styledContent[0]) // key & first line of value are on same line
		if len(styledContent) > 1 {
			styled = append(styled, styledContent[1:]...) // append remaining lines if there was more
		}
		if isLast {
			styled[len(styled)-1] = styled[len(styled)-1].TrimEnd(1) // remove styling for a trailing comma on last line
		}
	}

	//write suffix tokens
	suffix := func(token, comma string) string {
		return spf("%s%s%s\n", tabs(nestLevel-1), token, comma)
	}
	raw.WriteString(suffix("}", ","))
	// TODO: handle hasContent earlier to ensure '},' is not treated as new line
	// if there was no content
	styled = append(styled, styles.JSONLineStyling{}.AppendRuneStyleLG(tokenSt.PaddingLeft(len(tabs(nestLevel-1)))).AppendRuneStyleLG(tokenSt))

	return raw.String(), styled, kv
}

func (p JSONParser) switchAttrValueJSON(v types.AttributeValue, hashkey string, rangekey *string, nestLevel int) (string, []styles.JSONLineStyling) {
	strSt := p.Styles.StringStyle
	numSt := p.Styles.NumberStyle
	bolSt := p.Styles.BoolStyle
	bytSt := p.Styles.BytesStyle
	tokSt := p.Styles.TokenStyle
	nulSt := p.Styles.NULLStyle
	errSt := p.Styles.ErrorStyle

	switch vv := v.(type) {
	case *types.AttributeValueMemberB:
		return pJSONBytes(vv.Value, tokSt, bytSt)
	case *types.AttributeValueMemberBOOL:
		return pJSONBool(vv.Value, tokSt, bolSt)
	case *types.AttributeValueMemberBS:
		return stringableAsListJSON(p.Styles, vv.Value, nestLevel, func(s []byte) (string, []styles.JSONLineStyling) { return pJSONBytes(s, tokSt, bytSt) })
	case *types.AttributeValueMemberL:
		return stringableAsListJSON(p.Styles, vv.Value, nestLevel, func(s types.AttributeValue) (string, []styles.JSONLineStyling) {
			return p.switchAttrValueJSON(s, hashkey, rangekey, nestLevel+1)
		})
	case *types.AttributeValueMemberM:
		raw := strings.Builder{}
		str, st, _ := p.pJSON(vv.Value, hashkey, rangekey, nestLevel)
		fmt.Fprintf(&raw, "%s", str)
		return raw.String(), st
	case *types.AttributeValueMemberN:
		return pJSONNum(vv.Value, tokSt, numSt)
	case *types.AttributeValueMemberNS:
		return stringableAsListJSON(p.Styles, vv.Value, nestLevel, func(s string) (string, []styles.JSONLineStyling) { return pJSONNum(s, tokSt, numSt) })
	case *types.AttributeValueMemberNULL: // TODO: ignore?
		v := util.Ternary("NULL", "NOT NULL", vv.Value)
		return pJSONNULL(v, tokSt, nulSt)
	case *types.AttributeValueMemberS:
		return pJSONString(vv.Value, tokSt, strSt)
	case *types.AttributeValueMemberSS:
		return stringableAsListJSON(p.Styles, vv.Value, nestLevel, func(s string) (string, []styles.JSONLineStyling) { return pJSONString(s, tokSt, strSt) })
	default:
		fm := "<failed to parse>"
		return pJSONERR(fm, tokSt, errSt)
	}
}

func stringableAsListJSON[S []E, E any](stls jsonParserStyles, items S, nestLevel int, tr func(E) (string, []styles.JSONLineStyling)) (string, []styles.JSONLineStyling) {
	tokenSt := stls.TokenStyle

	if len(items) == 0 {
		return emptyBrackets("[]", tokenSt)
	}

	json := strings.Builder{}
	styled := []styles.JSONLineStyling{}

	prefix := func(token string) string { return spf("%s\n", token) }
	json.WriteString(prefix("["))
	styled = append(styled, styles.JSONLineStyling{}.AppendRuneStyleLG(tokenSt))

	listItem := func(in, token string, atEnd bool) string {
		return spf("%s%s",
			tabs(nestLevel+1), // only tab the first line of the parsed item, rest should already be tabbed; helps with tabbing of '{' in list vs outside list
			suffixIf(trimSuffixIf(in, spf("%s\n", token), atEnd), "\n", atEnd), // no trailing commas
		)
	}
	for i, v := range items {
		j, styledContent := tr(v)
		json.WriteString(listItem(j, ",", i == len(items)-1))
		for i, st := range styledContent {
			if i == 0 {
				st = st.SetLeftPaddingFirst(len(tabs(nestLevel + 1))) // only tab first line
			}
			styled = append(styled, st)
		}
		if i == len(items)-1 { // if last line
			styled[len(styled)-1] = styled[len(styled)-1].TrimEnd(1) // trim comma token style
		}
	}

	suffix := func(token, comma string) string {
		return spf("%s%s%s\n", tabs(nestLevel), token, comma)
	}
	json.WriteString(suffix("]", ","))
	styled = append(styled, styles.JSONLineStyling{}.AppendRuneStyleLG(tokenSt.PaddingLeft(len(tabs(nestLevel)))).AppendRuneStyleLG(tokenSt))

	return json.String(), styled
}

func emptyBrackets(brackets string, tokenStyle lipgloss.Style) (string, []styles.JSONLineStyling) {
	return spf("%s,\n", brackets), []styles.JSONLineStyling{styles.JSONLineStyling{}.AppendStringStyleLG(brackets, tokenStyle).AppendRuneStyleLG(tokenStyle)}
}

func pJSONBool(bl bool, tokenStyle, contentStyle lipgloss.Style) (string, []styles.JSONLineStyling) {
	b := spf("%t", bl)
	styled := []styles.JSONLineStyling{styles.JSONLineStyling{}.AppendStringStyleLG(b, contentStyle).AppendRuneStyleLG(tokenStyle)}
	return spf(jsonFmt, b, ","), styled
}

func pJSONBytes(bt []byte, tokenStyle, contentStyle lipgloss.Style) (string, []styles.JSONLineStyling) {
	bytesFmt := "<bytes>(len=%d)"
	b := spf(bytesFmt, len(bt))
	styled := []styles.JSONLineStyling{styles.JSONLineStyling{}.AppendStringStyleLG(b, contentStyle).AppendRuneStyleLG(tokenStyle)}
	return spf(jsonFmt, b, ","), styled
}

func pJSONNULL(n string, tokenStyle, contentStyle lipgloss.Style) (string, []styles.JSONLineStyling) {
	s := spf("%s", n)
	styled := []styles.JSONLineStyling{styles.JSONLineStyling{}.AppendStringStyleLG(s, contentStyle).AppendRuneStyleLG(tokenStyle)}
	return spf(jsonFmt, s, ","), styled
}

func pJSONERR(err string, tokenStyle, contentStyle lipgloss.Style) (string, []styles.JSONLineStyling) {
	s := spf("%q", err)
	styled := []styles.JSONLineStyling{styles.JSONLineStyling{}.AppendStringStyleLG(s, contentStyle).AppendRuneStyleLG(tokenStyle)}
	return spf(jsonFmt, s, ","), styled
}

func pJSONString(str string, tokenStyle, contentStyle lipgloss.Style) (string, []styles.JSONLineStyling) {
	s := spf("%q", str)
	styled := []styles.JSONLineStyling{styles.JSONLineStyling{}.AppendStringStyleLG(s, contentStyle).AppendRuneStyleLG(tokenStyle)}
	return spf(jsonFmt, s, ","), styled
}

func pJSONNum(num string, tokenStyle, contentStyle lipgloss.Style) (string, []styles.JSONLineStyling) {
	n := spf("%s", num)
	styled := []styles.JSONLineStyling{styles.JSONLineStyling{}.AppendStringStyleLG(n, contentStyle).AppendRuneStyleLG(tokenStyle)}
	return spf(jsonFmt, n, ","), styled
}
