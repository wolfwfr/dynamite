package parsing

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/wolfwfr/dynamite/lib/styles"
	apitypes "github.com/wolfwfr/dynamite/pkg/adapters/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/theme"
	"github.com/wolfwfr/dynamite/pkg/util"
)

const (
	jsonFmt = "%s%s\n"
)

type JSONParser struct {
	styles  jsonParserStyles
	tabSize int
}

type jsonParserStyles struct {
	fieldNameStyle lipgloss.Style
	numberStyle    lipgloss.Style
	boolStyle      lipgloss.Style
	bytesStyle     lipgloss.Style
	nullStyle      lipgloss.Style
	stringStyle    lipgloss.Style
	tokenStyle     lipgloss.Style
	errorStyle     lipgloss.Style
}

func newjsonParserStyles() jsonParserStyles {
	p := jsonParserStyles{}
	p.fieldNameStyle = lipgloss.NewStyle().Foreground(theme.FieldNameFg)
	p.numberStyle = lipgloss.NewStyle().Foreground(theme.NumberFg)
	p.boolStyle = lipgloss.NewStyle().Foreground(theme.BoolFg)
	p.bytesStyle = lipgloss.NewStyle().Foreground(theme.BytesFg)
	p.nullStyle = lipgloss.NewStyle().Foreground(theme.NULLFg)
	p.stringStyle = lipgloss.NewStyle().Foreground(theme.StringFg)
	p.tokenStyle = lipgloss.NewStyle().Foreground(theme.TokenFg)
	p.errorStyle = lipgloss.NewStyle().Foreground(theme.ErrorFg)
	return p
}

func NewJSONParser(tabSize int) JSONParser {
	p := JSONParser{tabSize: tabSize}
	p.styles = newjsonParserStyles()
	return p
}

func (p JSONParser) ParseToJSONWithKeys(item map[string]types.AttributeValue, hashkey string, rangekey *string) (string, []styles.LineStyle, []apitypes.KeyValue) {
	json, styled, keyValues := p.pJSON(item, hashkey, rangekey, 0)
	// trim trailing comma
	if len(styled) > 0 {
		styled[len(styled)-1] = styled[len(styled)-1].TrimEnd(1)
	}
	return strings.TrimSuffix(json, ",\n"), styled, keyValues
}

func (p JSONParser) ParseItemToJSON(item map[string]types.AttributeValue, hashkey string, rangekey *string) (string, []styles.LineStyle) {
	json, styled, _ := p.pJSON(item, hashkey, rangekey, 0)
	// trim trailing comma
	if len(styled) > 0 {
		styled[len(styled)-1] = styled[len(styled)-1].TrimEnd(1)
	}
	return strings.TrimSuffix(json, ",\n"), styled
}

// nestLevel determines indentations pJSON is an internal, recursive function
// that takes a dynamo-db item and parses it to a json-formatted string.
// While parsing to JSON, it also tracks per-rune styling for applying syntax
// highlighting to the raw string.
// TODO: consider elegant way of separating json-parsing from string->string
// key-value mapping, but for now this saves double work and the two are always
// used together.
func (p JSONParser) pJSON(elements map[string]types.AttributeValue, hashkey string, rangekey *string, nestLevel int) (string, styles.ObjectStyle, []apitypes.KeyValue) {
	raw := strings.Builder{}
	styled := []styles.LineStyle{}

	// obtain sorted keys
	keysSorted := getSortedKeys(hashkey, rangekey, elements, nestLevel == 0)

	nestLevel += 1

	isRootLevel := nestLevel == 1
	var kv []apitypes.KeyValue
	if isRootLevel {
		kv = make([]apitypes.KeyValue, len(keysSorted))
	}

	tokenSt := p.styles.tokenStyle
	fieldSt := p.styles.fieldNameStyle

	if len(keysSorted) == 0 { // no content
		raw, styled := emptyBrackets("{}", tokenSt)
		return raw, styled, kv
	}

	// write prefix token
	raw.WriteString("{\n")
	styled = append(styled, styles.LineStyle{}.AppendRuneLG((tokenSt)))

	for i, k := range keysSorted {
		v := elements[k]
		isLast := i == len(keysSorted)-1

		// write field-name
		quotedName := spf("\"%s\"", k)
		tbs := tabs(p.tabSize, nestLevel)
		raw.WriteString(spf("%s%s: ", tbs, quotedName))

		styled = append(styled, styles.LineStyle{}.AppendStringLG(quotedName, fieldSt, styles.
			WithStringInitialPadding(len(tbs))).
			AppendRuneLG(tokenSt). // :
			AppendRuneLG(tokenSt), // _ (space)
		)

		// obtain block content
		content, styledContent := p.switchAttrValueJSON(v, hashkey, rangekey, nestLevel)

		// prepare table keys
		if isRootLevel {
			kv[i] = apitypes.KeyValue{Key: k, Value: flatten(content, ","), ValueStyling: flattenStyles(styledContent).TrimEnd(1)}
		}

		// write comma & newline, unless last element
		raw.WriteString(suffixIf(trimSuffixIf(content, ",\n", isLast), "\n", isLast)) // if last, replace "<comma>\n" with "\n"
		styled[len(styled)-1] = styled[len(styled)-1].AppendLine(styledContent[0])    // key & first line of value are on same line
		if len(styledContent) > 1 {
			styled = append(styled, styledContent[1:]...) // append remaining lines if there was more
		}
		if isLast {
			styled[len(styled)-1] = styled[len(styled)-1].TrimEnd(1) // remove styling for a trailing comma on last line
		}
	}

	//write suffix tokens
	tbs := tabs(p.tabSize, nestLevel-1)
	raw.WriteString(spf("%s},\n", tbs))
	styled = append(styled, styles.LineStyle{}.AppendRuneLG(tokenSt.PaddingLeft(len(tbs))).AppendRuneLG(tokenSt))

	return raw.String(), styled, kv
}

func (p JSONParser) switchAttrValueJSON(v types.AttributeValue, hashkey string, rangekey *string, nestLevel int) (string, styles.ObjectStyle) {
	strSt := p.styles.stringStyle
	numSt := p.styles.numberStyle
	bolSt := p.styles.boolStyle
	bytSt := p.styles.bytesStyle
	tokSt := p.styles.tokenStyle
	nulSt := p.styles.nullStyle
	errSt := p.styles.errorStyle

	obj := parseListHelper

	switch vv := v.(type) {
	case *types.AttributeValueMemberB:
		return obj(pJSONBytes(vv.Value, tokSt, bytSt))
	case *types.AttributeValueMemberBOOL:
		return obj(pJSONBool(vv.Value, tokSt, bolSt))
	case *types.AttributeValueMemberBS:
		return stringableAsListJSON(p.styles, vv.Value, p.tabSize, nestLevel, func(s []byte) (string, styles.ObjectStyle) { return obj(pJSONBytes(s, tokSt, bytSt)) })
	case *types.AttributeValueMemberL:
		return stringableAsListJSON(p.styles, vv.Value, p.tabSize, nestLevel, func(s types.AttributeValue) (string, styles.ObjectStyle) {
			return p.switchAttrValueJSON(s, hashkey, rangekey, nestLevel+1)
		})
	case *types.AttributeValueMemberM:
		raw := strings.Builder{}
		str, st, _ := p.pJSON(vv.Value, hashkey, rangekey, nestLevel)
		fmt.Fprintf(&raw, "%s", str)
		return raw.String(), st
	case *types.AttributeValueMemberN:
		return obj(pJSONNum(vv.Value, tokSt, numSt))
	case *types.AttributeValueMemberNS:
		return stringableAsListJSON(p.styles, vv.Value, p.tabSize, nestLevel, func(s string) (string, styles.ObjectStyle) { return obj(pJSONNum(s, tokSt, numSt)) })
	case *types.AttributeValueMemberNULL:
		v := util.Ternary("NULL", "NOT NULL", vv.Value)
		return obj(pJSONNULL(v, tokSt, nulSt))
	case *types.AttributeValueMemberS:
		return obj(pJSONString(vv.Value, tokSt, strSt))
	case *types.AttributeValueMemberSS:
		return stringableAsListJSON(p.styles, vv.Value, p.tabSize, nestLevel, func(s string) (string, styles.ObjectStyle) { return obj(pJSONString(s, tokSt, strSt)) })
	default:
		fm := "<failed to parse>"
		return obj(pJSONERR(fm, tokSt, errSt))
	}
}

func stringableAsListJSON[S []E, E any](stls jsonParserStyles, items S, tabSize, nestLevel int, tr func(E) (string, styles.ObjectStyle)) (string, styles.ObjectStyle) {
	tokenSt := stls.tokenStyle

	if len(items) == 0 {
		return emptyBrackets("[]", tokenSt)
	}

	json := strings.Builder{}
	styled := styles.ObjectStyle{}

	json.WriteString("[\n")
	styled = append(styled, styles.LineStyle{}.AppendRuneLG(tokenSt))

	tbs := tabs(tabSize, nestLevel+1)
	listItem := func(in string, atEnd bool) string {
		return spf("%s%s",
			tbs, // only tab the first line of the parsed item, rest should already be tabbed; helps with tabbing of '{' in list vs outside list
			suffixIf(trimSuffixIf(in, ",\n", atEnd), "\n", atEnd), // no trailing commas
		)
	}
	for i, v := range items {
		j, styledContent := tr(v)
		json.WriteString(listItem(j, i == len(items)-1))
		for i, st := range styledContent {
			if i == 0 {
				st = st.SetLeftPaddingFirst(len(tbs)) // only tab first line
			}
			styled = append(styled, st)
		}
		if i == len(items)-1 { // if last line
			styled[len(styled)-1] = styled[len(styled)-1].TrimEnd(1) // trim comma token style
		}
	}

	tbs = tabs(tabSize, nestLevel)
	json.WriteString(spf("%s],\n", tbs))
	styled = append(styled, styles.LineStyle{}.AppendRuneLG(tokenSt.PaddingLeft(len(tbs))).AppendRuneLG(tokenSt))

	return json.String(), styled
}

func emptyBrackets(brackets string, tokenStyle lipgloss.Style) (string, styles.ObjectStyle) {
	return spf("%s,\n", brackets), styles.ObjectStyle{styles.LineStyle{}.AppendStringLG(brackets, tokenStyle).AppendRuneLG(tokenStyle)}
}

func pJSONBool(bl bool, tokenStyle, contentStyle lipgloss.Style) (string, styles.LineStyle) {
	b := spf("%t", bl)
	styled := styles.LineStyle{}.AppendStringLG(b, contentStyle).AppendRuneLG(tokenStyle)
	return spf(jsonFmt, b, ","), styled
}

func pJSONBytes(bt []byte, tokenStyle, contentStyle lipgloss.Style) (string, styles.LineStyle) {
	bytesFmt := "<bytes>(len=%d)"
	b := spf(bytesFmt, len(bt))
	styled := styles.LineStyle{}.AppendStringLG(b, contentStyle).AppendRuneLG(tokenStyle)
	return spf(jsonFmt, b, ","), styled
}

func pJSONNULL(n string, tokenStyle, contentStyle lipgloss.Style) (string, styles.LineStyle) {
	s := spf("%s", n)
	styled := styles.LineStyle{}.AppendStringLG(s, contentStyle).AppendRuneLG(tokenStyle)
	return spf(jsonFmt, s, ","), styled
}

func pJSONERR(err string, tokenStyle, contentStyle lipgloss.Style) (string, styles.LineStyle) {
	s := spf("%q", err)
	styled := styles.LineStyle{}.AppendStringLG(s, contentStyle).AppendRuneLG(tokenStyle)
	return spf(jsonFmt, s, ","), styled
}

func pJSONString(str string, tokenStyle, contentStyle lipgloss.Style) (string, styles.LineStyle) {
	s := spf("%q", str)
	styled := styles.LineStyle{}.AppendStringLG(s, contentStyle).AppendRuneLG(tokenStyle)
	return spf(jsonFmt, s, ","), styled
}

func pJSONNum(num string, tokenStyle, contentStyle lipgloss.Style) (string, styles.LineStyle) {
	n := spf("%s", num)
	styled := styles.LineStyle{}.AppendStringLG(n, contentStyle).AppendRuneLG(tokenStyle)
	return spf(jsonFmt, n, ","), styled
}
