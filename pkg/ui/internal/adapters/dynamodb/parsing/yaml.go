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

func (p YAMLParser) ParseItemToYAML(item map[string]types.AttributeValue, hashkey string, rangekey *string) (string, styles.ObjectStyle) {
	return p.pYAML(item, hashkey, rangekey, 0)
}

// nestLevel determines indentations pYAML is an internal, recursive function
// that takes a dynamo-db item and parses it to a yaml-formatted string.
func (p YAMLParser) pYAML(elements map[string]types.AttributeValue, hashkey string, rangekey *string, nestLevel int) (string, styles.ObjectStyle) {
	raw := strings.Builder{}
	styled := styles.ObjectStyle{}

	// obtain sorted keys
	keysSorted := getSortedKeys(hashkey, rangekey, elements, nestLevel == 0)

	fieldNameSt := p.Styles.FieldNameStyle
	tokenSt := p.Styles.TokenStyle

	tbs := tabs(nestLevel)
	for _, k := range keysSorted {
		fmt.Fprintf(&raw, "%s%s: ", tbs, k)
		styled = append(styled, styles.LineStyle{}.
			AppendStringLG(k, fieldNameSt, styles.WithStringInitialPadding(len(tbs))).
			AppendRuneLG(tokenSt). // ':'
			AppendRuneLG(tokenSt)) // '_' (space)
		v := elements[k]
		r, contentStyling := p.switchAttrValueYAML(v, hashkey, rangekey, nestLevel, false)
		raw.WriteString(r)

		if len(contentStyling) == 0 {
			continue
		}
		// always append the first line of the result to the key; should be
		// empty for objects or lists
		styled[len(styled)-1] = styled[len(styled)-1].AppendLine(contentStyling[0])
		if len(contentStyling) > 1 {
			styled = append(styled, contentStyling[1:]...) // then append the rest of the lines
		}
	}

	return raw.String(), styled
}

// switchAttrValueYAML should always return a styling object in the assumption
// that the first line of the object will be appended to the line of its key. In
// case of lists and objects, which start with a newline ("\n"), the first line
// in the styling-object must not refer to the object/list contents. Its true
// contents are irrelevent as they will not get rendered anyway.
func (p YAMLParser) switchAttrValueYAML(v types.AttributeValue, hashkey string, rangekey *string, nestLevel int, isListItem bool) (string, styles.ObjectStyle) {
	strSt := p.Styles.StringStyle
	numSt := p.Styles.NumberStyle
	bolSt := p.Styles.BoolStyle
	bytSt := p.Styles.BytesStyle
	nulSt := p.Styles.NULLStyle
	errSt := p.Styles.ErrorStyle

	obj := parseListHelper

	switch vv := v.(type) {
	case *types.AttributeValueMemberB:
		return obj(pYAMLBytes(vv.Value, bytSt))
	case *types.AttributeValueMemberBOOL:
		return obj(pYAMLBool(vv.Value, bolSt))
	case *types.AttributeValueMemberBS:
		return stringableAsListYAML(p.Styles, vv.Value, nestLevel, func(s []byte) (string, styles.ObjectStyle) { return obj(pYAMLBytes(s, bytSt)) })
	case *types.AttributeValueMemberL:
		return stringableAsListYAML(p.Styles, vv.Value, nestLevel, func(s types.AttributeValue) (string, styles.ObjectStyle) {
			return p.switchAttrValueYAML(s, hashkey, rangekey, nestLevel, true)
		})
	case *types.AttributeValueMemberM:
		b := strings.Builder{}
		s := styles.ObjectStyle{styles.LineStyle{}} // start with empty style
		raw, contentStyling := p.pYAML(vv.Value, hashkey, rangekey, nestLevel+1)
		if isListItem {
			s = styles.ObjectStyle{} // remove first empty styling if both list & object
			// add 2 spaces (in lieu of '- ') after each newline; remove 2 spaces at the end
			raw = strings.TrimSuffix(strings.ReplaceAll(raw, "\n", "\n  "), "  ")
			for i, st := range contentStyling {
				if i == 0 {
					contentStyling[i] = st.SetLeftPaddingFirst(0) // no padding on first line (which is appended to list's '- ')
					continue
				}
				// extra padding (in lieu of '- ') for remaining lines of the list item
				first, _ := st.GetAt(0)
				contentStyling[i] = st.Override(0, first.PaddingLeft(first.GetPaddingLeft()+2))
			}
		}
		tbs := tabs(nestLevel + 1)
		prefix := func(s string) string {
			// if not a list item (then newline already prepended) && object not empty, then prepend with newline
			// if list-item, remove prepended tabs (list does that for first line)
			// if object is empty, end with newline
			return spf("%s%s%s", newLineIf(!isListItem && s != ""), trimPrefixIf(s, tbs, isListItem), newLineIf(s == ""))
		}
		b.WriteString(prefix(raw))
		s = append(s, contentStyling...)
		return b.String(), s
	case *types.AttributeValueMemberN:
		return obj(pYAMLNum(vv.Value, numSt))
	case *types.AttributeValueMemberNS:
		return stringableAsListYAML(p.Styles, vv.Value, nestLevel, func(s string) (string, styles.ObjectStyle) { return obj(pYAMLNum(s, numSt)) })
	case *types.AttributeValueMemberNULL: // TODO: ignore?
		v := util.Ternary("NULL", "NOT NULL", vv.Value)
		return obj(pYAMLNULL(v, nulSt))
	case *types.AttributeValueMemberS:
		return obj(pYAMLString(vv.Value, strSt))
	case *types.AttributeValueMemberSS:
		return stringableAsListYAML(p.Styles, vv.Value, nestLevel, func(s string) (string, styles.ObjectStyle) { return obj(pYAMLString(s, strSt)) })
	default:
		fm := "<failed to parse>"
		return obj(pYAMLERR(fm, errSt)) // TODO: error?
	}
}

func stringableAsListYAML[S []E, E any](stls yamlParserStyles, s S, nestLevel int, tr func(E) (string, styles.ObjectStyle)) (string, styles.ObjectStyle) {
	tokenSt := stls.TokenStyle

	raw := strings.Builder{}
	styled := styles.ObjectStyle{styles.LineStyle{}} // start with an empty line

	raw.WriteString("\n")

	tbs := tabs(nestLevel + 1)

	for _, v := range s {
		r, contentStyling := tr(v)
		raw.WriteString(spf("%s- %s", tbs, r))
		for i, st := range contentStyling {
			if i == 0 {
				st = styles.LineStyle{}.
					AppendRuneLG(tokenSt.PaddingLeft(len(tbs))). // '-' (dash)
					AppendRuneLG(tokenSt).                       // '_' (space)
					AppendLine(st)
			}
			styled = append(styled, st)
		}
	}

	return raw.String(), styled
}

func pYAMLBool(bl bool, contentStyle lipgloss.Style) (string, styles.LineStyle) {
	b := spf("%t", bl)
	styled := styles.LineStyle{}.AppendStringLG(b, contentStyle)
	return spf(yamlFmt, b), styled
}

func pYAMLBytes(bt []byte, contentStyle lipgloss.Style) (string, styles.LineStyle) {
	bytesFmt := "<bytes>(len=%d)"
	b := spf(bytesFmt, len(bt))
	styled := styles.LineStyle{}.AppendStringLG(b, contentStyle)
	return spf(yamlFmt, b), styled
}

func pYAMLNULL(n string, contentStyle lipgloss.Style) (string, styles.LineStyle) {
	s := spf("%s", n)
	styled := styles.LineStyle{}.AppendStringLG(s, contentStyle)
	return spf(yamlFmt, s), styled
}

func pYAMLERR(err string, contentStyle lipgloss.Style) (string, styles.LineStyle) {
	s := spf("%q", err)
	styled := styles.LineStyle{}.AppendStringLG(s, contentStyle)
	return spf(yamlFmt, s), styled
}

func pYAMLString(str string, contentStyle lipgloss.Style) (string, styles.LineStyle) {
	s := spf("%q", str)
	styled := styles.LineStyle{}.AppendStringLG(s, contentStyle)
	return spf(yamlFmt, s), styled
}

func pYAMLNum(num string, contentStyle lipgloss.Style) (string, styles.LineStyle) {
	n := spf("%s", num)
	styled := styles.LineStyle{}.AppendStringLG(n, contentStyle)
	return spf(yamlFmt, n), styled
}
