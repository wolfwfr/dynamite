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
	// token := p.Styles.TokenStyle.Render
	json, styled, keyValues := p.pJSON(item, hashkey, rangekey, 0)
	trim := func(s, token string) string { return strings.TrimSuffix(s, spf("%s\n", token)) }
	if len(styled) > 0 {
		styled[len(styled)-1] = styled[len(styled)-1].TrimEnd(1)
	}
	return trim(json, ","), styled, keyValues
}

func (p JSONParser) ParseItemToJSON(item map[string]types.AttributeValue, hashkey string, rangekey *string) (string, []styles.JSONLineStyling) {
	// token := p.Styles.TokenStyle.Render
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

	// token := p.Styles.TokenStyle.Render
	tokenSt := p.Styles.TokenStyle
	// field := p.Styles.FieldNameStyle.Render
	fieldSt := p.Styles.FieldNameStyle

	hasContent := len(keysSorted) > 0

	// write prefix token
	prefix := func(token string) string { return spf("%s%s", token, newLineIf(hasContent)) }
	raw.WriteString(prefix("{"))
	styled = append(styled, styles.JSONLineStyling{}.AppendRuneStyleLG((tokenSt.PaddingLeft(len(tabs(nestLevel - 1))))))
	// styled.WriteString(prefix(token("{")))

	for i, k := range keysSorted {
		v := elements[k]
		isLast := i == len(keysSorted)-1

		// write field-name
		fieldName := func(quotedName, colon string) string {
			return spf("%s%s%s ", tabs(nestLevel), quotedName, colon)
		}
		quotedName := spf("\"%s\"", k)
		raw.WriteString(fieldName(quotedName, ":"))

		// styled = append(styled, styles.JSONLineStyling{}.AppendStringStyleLG(quotedName, fieldSt).AppendRuneStyleLG(tokenSt))
		styled = append(styled, styles.JSONLineStyling{}.AppendStringStyleLG(quotedName, fieldSt, styles.WithStringInitialPadding(len(tabs(nestLevel)))).AppendRuneStyleLG(tokenSt).AppendRuneStyleLG(tokenSt))
		// styled.WriteString(fieldName(field(quotedName), token(":")))

		// obtain block content
		content, styledContent := p.switchAttrValueJSON(v, hashkey, rangekey, nestLevel)

		// prepare table keys
		if isRootLevel {
			// kv[i] = apitypes.KeyValue{Key: k, Value: flatten(content, ","), StyledValue: flatten(styledContent, token(","))}
			kv[i] = apitypes.KeyValue{Key: k, Value: flatten(content, ","), ValueStyling: flattenStyles(styledContent)}
		}

		// write comma & newline, unless last element
		withSuffix := func(s, comma string) string {
			return spf("%s", suffixIf(trimSuffixIf(s, spf("%s\n", comma), isLast), "\n", isLast)) // no trailing commas
		}
		raw.WriteString(withSuffix(content, ","))
		styled[len(styled)-1] = styled[len(styled)-1].AppendLineStyle(styledContent[0])
		if len(styledContent) > 1 {
			styled = append(styled, styledContent[1:]...)
		}
		styled[len(styled)-1] = styled[len(styled)-1].AppendRuneStyleLG(tokenSt)
		// styled.WriteString(withSuffix(styledContent, token(",")))
	}

	//write suffix tokens
	suffix := func(token, comma string) string {
		return spf("%s%s%s\n", prefixIf("", tabs(nestLevel-1), hasContent), token, comma)
	}
	raw.WriteString(suffix("}", ","))
	// TODO: handle hasContent earlier to ensure '},' is not treated as new line
	// if there was no content
	styled = append(styled, styles.JSONLineStyling{}.AppendRuneStyleLG(tokenSt.PaddingLeft(len(tabs(nestLevel-1)))).AppendRuneStyleLG(tokenSt))
	// styled.WriteString(suffix(token("}"), token(",")))

	return raw.String(), styled, kv
}

func (p JSONParser) switchAttrValueJSON(v types.AttributeValue, hashkey string, rangekey *string, nestLevel int) (string, []styles.JSONLineStyling) {
	// str := p.Styles.StringStyle.Render
	strSt := p.Styles.StringStyle
	// num := p.Styles.NumberStyle.Render
	numSt := p.Styles.NumberStyle
	// bol := p.Styles.BoolStyle.Render
	bolSt := p.Styles.BoolStyle
	// byt := p.Styles.BytesStyle.Render
	bytSt := p.Styles.BytesStyle
	// tok := p.Styles.TokenStyle.Render
	tokSt := p.Styles.TokenStyle
	// nul := p.Styles.NULLStyle.Render
	nulSt := p.Styles.NULLStyle
	// err := p.Styles.ErrorStyle.Render
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
		// raw, styled := strings.Builder{}, strings.Builder{}
		raw := strings.Builder{}
		str, st, _ := p.pJSON(vv.Value, hashkey, rangekey, nestLevel)
		fmt.Fprintf(&raw, "%s", str)
		// fmt.Fprintf(&styled, "%s", st)
		return raw.String(), st
	case *types.AttributeValueMemberN:
		return pJSONNum(vv.Value, tokSt, numSt)
	case *types.AttributeValueMemberNS:
		return stringableAsListJSON(p.Styles, vv.Value, nestLevel, func(s string) (string, []styles.JSONLineStyling) { return pJSONNum(s, tokSt, numSt) })
	case *types.AttributeValueMemberNULL: // TODO: ignore?
		v := util.Ternary("NULL", "NOT NULL", vv.Value)
		return pJSONNULL(v, tokSt, nulSt)
		// return spf(jsonFmt, v, ","), spf(jsonFmt, nul(v), tok(","))
	case *types.AttributeValueMemberS:
		return pJSONString(vv.Value, tokSt, strSt)
	case *types.AttributeValueMemberSS:
		return stringableAsListJSON(p.Styles, vv.Value, nestLevel, func(s string) (string, []styles.JSONLineStyling) { return pJSONString(s, tokSt, strSt) })
	default:
		fm := "<failed to parse>"
		return pJSONERR(fm, tokSt, errSt)
		// return spf(jsonFmt, fm, ","), spf(jsonFmt, err(fm), tok(",")) // TODO: error?
	}
}

func stringableAsListJSON[S []E, E any](stls jsonParserStyles, s S, nestLevel int, tr func(E) (string, []styles.JSONLineStyling)) (string, []styles.JSONLineStyling) {
	// token := stls.TokenStyle.Render
	tokenSt := stls.TokenStyle

	if len(s) == 0 {
		// return "[],\n", spf("%s%s\n", token("[]"), token(","))
		return "[],\n", []styles.JSONLineStyling{styles.JSONLineStyling{}.AppendStringStyleLG("[]", tokenSt).AppendRuneStyleLG(tokenSt)}
	}

	json := strings.Builder{}
	// styled := strings.Builder{}
	styled := []styles.JSONLineStyling{}

	prefix := func(token string) string { return spf("%s\n", token) }
	json.WriteString(prefix("["))
	styled = append(styled, styles.JSONLineStyling{}.AppendRuneStyleLG(tokenSt))
	// styled.WriteString(prefix(token("[")))

	line := func(in, token string, atEnd bool) string {
		return spf("%s%s", tabs(nestLevel+1), suffixIf(trimSuffixIf(in, spf("%s\n", token), atEnd), "\n", atEnd)) // no trailing commas
	}
	for i, v := range s {
		j, styledContent := tr(v)
		json.WriteString(line(j, ",", i == len(s)-1))
		for _, st := range styledContent {
			styled = append(styled, st.SetPaddingAll(nestLevel+1))
		}
		if i == len(s)-1 { // if last line
			styled[len(styled)-1] = styled[len(styled)-1].TrimEnd(1) // trim comma token style
		}
		// styled.WriteString(line(styledContent, token(","), i == len(s)-1))
	}

	suffix := func(token, comma string) string {
		return spf("%s%s%s\n", tabs(nestLevel), token, comma)
	}
	json.WriteString(suffix("]", ","))
	styled = append(styled, styles.JSONLineStyling{}.AppendRuneStyleLG(tokenSt.PaddingLeft(len(tabs(nestLevel)))).AppendRuneStyleLG(tokenSt))
	// styled.WriteString(suffix(token("]"), token(",")))

	// return json.String(), styled.String()
	return json.String(), styled
}

func pJSONBool(bl bool, tokenStyle, contentStyle lipgloss.Style) (string, []styles.JSONLineStyling) {
	b := spf("%t", bl)
	styled := []styles.JSONLineStyling{styles.JSONLineStyling{}.AppendStringStyleLG(b, contentStyle).AppendRuneStyleLG(tokenStyle)}
	// return spf(jsonFmt, b, ","), spf(jsonFmt, render(b), renderedToken)
	return spf(jsonFmt, b, ","), styled
}

func pJSONBytes(bt []byte, tokenStyle, contentStyle lipgloss.Style) (string, []styles.JSONLineStyling) {
	bytesFmt := "<bytes>(len=%d)"
	b := spf(bytesFmt, len(bt))
	styled := []styles.JSONLineStyling{styles.JSONLineStyling{}.AppendStringStyleLG(b, contentStyle).AppendRuneStyleLG(tokenStyle)}
	// return spf(jsonFmt, b, ","), spf(jsonFmt, render(b), renderedToken)
	return spf(jsonFmt, b, ","), styled
}

func pJSONNULL(n string, tokenStyle, contentStyle lipgloss.Style) (string, []styles.JSONLineStyling) {
	s := spf("%s", n)
	styled := []styles.JSONLineStyling{styles.JSONLineStyling{}.AppendStringStyleLG(s, contentStyle).AppendRuneStyleLG(tokenStyle)}
	// return spf(jsonFmt, s, ","), spf(jsonFmt, render(s), renderedToken)
	return spf(jsonFmt, s, ","), styled
}

func pJSONERR(err string, tokenStyle, contentStyle lipgloss.Style) (string, []styles.JSONLineStyling) {
	s := spf("%q", err)
	styled := []styles.JSONLineStyling{styles.JSONLineStyling{}.AppendStringStyleLG(s, contentStyle).AppendRuneStyleLG(tokenStyle)}
	// return spf(jsonFmt, s, ","), spf(jsonFmt, render(s), renderedToken)
	return spf(jsonFmt, s, ","), styled
}

func pJSONString(str string, tokenStyle, contentStyle lipgloss.Style) (string, []styles.JSONLineStyling) {
	s := spf("%q", str)
	styled := []styles.JSONLineStyling{styles.JSONLineStyling{}.AppendStringStyleLG(s, contentStyle).AppendRuneStyleLG(tokenStyle)}
	// return spf(jsonFmt, s, ","), spf(jsonFmt, render(s), renderedToken)
	return spf(jsonFmt, s, ","), styled
}

func pJSONNum(num string, tokenStyle, contentStyle lipgloss.Style) (string, []styles.JSONLineStyling) {
	n := spf("%s", num)
	styled := []styles.JSONLineStyling{styles.JSONLineStyling{}.AppendStringStyleLG(n, contentStyle).AppendRuneStyleLG(tokenStyle)}
	// return spf(jsonFmt, n, ","), spf(jsonFmt, render(n), renderedToken)
	return spf(jsonFmt, n, ","), styled
}
