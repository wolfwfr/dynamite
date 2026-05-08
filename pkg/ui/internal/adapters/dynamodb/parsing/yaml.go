package parsing

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
	"github.com/wolfwfr/dynamite/pkg/util"
)

const (
	yamlFmt = "%s\n"
)

type YAMLParser struct {
	Styles yamlParserStyles
}

type yamlParserStyles struct {
	FieldNameStyle lipgloss.Style
	NumberStyle    lipgloss.Style
	BoolStyle      lipgloss.Style
	BytesStyle     lipgloss.Style
	NULLStyle      lipgloss.Style
	StringStyle    lipgloss.Style
	TokenStyle     lipgloss.Style
	ErrorStyle     lipgloss.Style
}

func NewYAMLParser() YAMLParser {
	p := YAMLParser{}
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

func (p YAMLParser) ParseItemToYAML(item map[string]types.AttributeValue, hashkey string, rangekey *string) (string, string) {
	return p.pYAML(item, hashkey, rangekey, 0)
}

// nestLevel determines indentations pYAML is an internal, recursive function
// that takes a dynamo-db item and parses it to a yaml-formatted string.
func (p YAMLParser) pYAML(elements map[string]types.AttributeValue, hashkey string, rangekey *string, nestLevel int) (string, string) {
	raw, styled := strings.Builder{}, strings.Builder{}

	// obtain sorted keys
	keysSorted := getSortedKeys(hashkey, rangekey, elements, nestLevel == 0)

	fieldName := p.Styles.FieldNameStyle.Render

	for _, k := range keysSorted {
		fmt.Fprintf(&raw, "%s%s: ", tabs(nestLevel), k)
		fmt.Fprintf(&styled, "%s%s: ", tabs(nestLevel), fieldName(k))
		v := elements[k]
		r, s := p.switchAttrValueYAML(v, hashkey, rangekey, nestLevel, false)
		raw.WriteString(r)
		styled.WriteString(s)
	}

	return raw.String(), styled.String()
}

func (p YAMLParser) switchAttrValueYAML(v types.AttributeValue, hashkey string, rangekey *string, nestLevel int, isListItem bool) (string, string) {
	str := p.Styles.StringStyle.Render
	num := p.Styles.NumberStyle.Render
	bol := p.Styles.BoolStyle.Render
	byt := p.Styles.BytesStyle.Render
	nul := p.Styles.NULLStyle.Render
	err := p.Styles.ErrorStyle.Render

	switch vv := v.(type) {
	case *types.AttributeValueMemberB:
		return pYAMLBytes(vv.Value, byt)
	case *types.AttributeValueMemberBOOL:
		return pYAMLBool(vv.Value, bol)
	case *types.AttributeValueMemberBS:
		return stringableAsListYAML(p.Styles, vv.Value, nestLevel, func(s []byte) (string, string) { return pYAMLBytes(s, byt) })
	case *types.AttributeValueMemberL:
		return stringableAsListYAML(p.Styles, vv.Value, nestLevel, func(s types.AttributeValue) (string, string) {
			return p.switchAttrValueYAML(s, hashkey, rangekey, nestLevel, true)
		})
	case *types.AttributeValueMemberM:
		b, s := strings.Builder{}, strings.Builder{}
		raw, styled := p.pYAML(vv.Value, hashkey, rangekey, nestLevel+1)
		if isListItem {
			trim := func(s string) string { return strings.TrimSuffix(strings.ReplaceAll(s, "\n", "\n  "), "  ") }
			raw = trim(raw)
			styled = trim(styled)
		}
		pr := func(s string) string {
			return spf("%s%s%s", newLineIf(!isListItem && s != ""), trimPrefixIf(s, tabs(nestLevel+1), isListItem), newLineIf(s == ""))
		}
		b.WriteString(pr(raw))
		s.WriteString(pr(styled))
		return b.String(), s.String()
	case *types.AttributeValueMemberN:
		return pYAMLNum(vv.Value, num)
	case *types.AttributeValueMemberNS:
		return stringableAsListYAML(p.Styles, vv.Value, nestLevel, func(s string) (string, string) { return pYAMLNum(s, num) })
	case *types.AttributeValueMemberNULL: // TODO: ignore?
		v := util.Ternary("NULL", "NOT NULL", vv.Value)
		return spf(yamlFmt, v), spf(yamlFmt, nul(v))
	case *types.AttributeValueMemberS:
		return pYAMLString(vv.Value, str)
	case *types.AttributeValueMemberSS:
		return stringableAsListYAML(p.Styles, vv.Value, nestLevel, func(s string) (string, string) { return pYAMLString(s, str) })
	default:
		fm := "<failed to parse>"
		return spf(yamlFmt, fm), spf(yamlFmt, err(fm)) // TODO: error?
	}
}

func stringableAsListYAML[S []E, E any](styles yamlParserStyles, s S, nestLevel int, tr func(E) (string, string)) (string, string) {
	token := styles.TokenStyle.Render

	raw := strings.Builder{}
	styled := strings.Builder{}

	raw.WriteString("\n")
	styled.WriteString("\n")

	line := func(token, ele string) string { return spf("%s%s %s", tabs(nestLevel+1), token, ele) }

	for _, v := range s {
		r, st := tr(v)
		raw.WriteString(line("-", r))
		styled.WriteString(line(token("-"), st))
	}

	return raw.String(), styled.String()
}

func pYAMLBool(bl bool, render func(...string) string) (string, string) {
	b := spf("%t", bl)
	return spf(yamlFmt, b), spf(yamlFmt, render(b))
}

func pYAMLBytes(bt []byte, render func(...string) string) (string, string) {
	bytesFmt := "<bytes>(len=%d)"
	b := spf(bytesFmt, len(bt))
	return spf(yamlFmt, b), spf(yamlFmt, render(b))
}

func pYAMLString(str string, render func(...string) string) (string, string) {
	s := spf("%q", str)
	return spf(yamlFmt, s), spf(yamlFmt, render(s))
}

func pYAMLNum(num string, render func(...string) string) (string, string) {
	n := spf("%s", num)
	return spf(yamlFmt, n), spf(yamlFmt, render(n))
}
