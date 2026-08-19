package telegraph

import (
	"fmt"
	"github.com/TGScheme/TLExtractorBot/internal/consts"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/telegraph/types"
	"slices"
)

func parseHtml(html string) ([]types.Node, error) {
	chars := []rune(html)
	var dom []types.Node
	var rawNodes []types.RawNode
	var buildTag, closingTag, buildAttrName, findValue, buildSpecialChar, buildAttrValue bool
	attributes := make(map[string]string)
	var openedTags []string
	var varName, varValue, specialChar string
	var tag string
	var line int
	for i := 0; i < len(chars); i++ {
		if chars[i] == '\n' {
			line++
		}
		if chars[i] == '"' && findValue {
			buildAttrValue = !buildAttrValue
			findValue = buildAttrValue
			if !buildAttrValue {
				attributes[varName] = varValue
			}
		} else if buildAttrValue {
			varValue += string(chars[i])
		} else if chars[i] == '<' {
			if buildTag {
				return nil, fmt.Errorf("unexpected < at line %d", line)
			}
			buildTag = true
		} else if chars[i] == '>' {
			if closingTag && chars[i-1] != '/' {
				if len(openedTags) == 0 {
					return nil, fmt.Errorf("unexpected closing tag at line %d", line)
				}
				openedTag := openedTags[len(openedTags)-1]
				if openedTag != tag {
					return nil, fmt.Errorf("closing tag %s does not match opening tag %s at line %d", tag, openedTag, line)
				}
				if len(rawNodes) > 0 && rawNodes[len(rawNodes)-1].Nesting == len(openedTags) && slices.Contains(consts.TagUnsupportedChildren, openedTags[len(openedTags)-1]) {
					return nil, fmt.Errorf("tag %s cannot contain children at line %d", openedTags[len(openedTags)-1], line)
				}
				closingTag = false
				openedTags = openedTags[:len(openedTags)-1]
				if len(openedTags) == 0 {
					dom = append(dom, mergeNodes(rawNodes)...)
					rawNodes = nil
				}
			} else {
				if !slices.Contains(consts.SupportedHtmlTags, tag) {
					return nil, fmt.Errorf("unsupported tag %s at line %d", tag, line)
				}
				for attrName := range attributes {
					if !slices.Contains(consts.TagRequiredAttrs[tag], attrName) {
						return nil, fmt.Errorf("unsupported attribute %s for tag %s at line %d", attrName, tag, line)
					}
				}
				for _, attrName := range consts.TagRequiredAttrs[tag] {
					if _, ok := attributes[attrName]; !ok {
						return nil, fmt.Errorf("missing attribute %s for tag %s at line %d", attrName, tag, line)
					}
				}
				if findValue {
					return nil, fmt.Errorf("unexpected > at line %d", line)
				}
				rawNodes = append(rawNodes, types.RawNode{
					Nesting: len(openedTags),
					Tag:     tag,
					Attrs:   attributes,
				})
				attributes = make(map[string]string)
				if closingTag {
					closingTag = false
				} else if !slices.Contains(consts.UnclosedTags, tag) {
					openedTags = append(openedTags, tag)
				}
				buildAttrName = false
			}
			tag = ""
			buildTag = false
		} else if chars[i] == '/' && buildTag && !closingTag {
			closingTag = true
		} else if chars[i] == ' ' && buildTag && !closingTag {
			buildAttrName = true
			varName = ""
			varValue = ""
		} else if chars[i] == '=' && buildTag && !closingTag {
			if len(varName) == 0 {
				return nil, fmt.Errorf("unexpected = at line %d", line)
			}
			buildAttrName = false
			findValue = true
		} else if buildAttrName {
			varName += string(chars[i])
		} else if buildTag {
			tag += string(chars[i])
		} else {
			if len(rawNodes) == 0 || rawNodes[len(rawNodes)-1].Tag != "text" || rawNodes[len(rawNodes)-1].Nesting != len(openedTags) {
				rawNodes = append(rawNodes, types.RawNode{
					Nesting: len(openedTags),
					Tag:     "text",
					Attrs:   map[string]string{"value": ""},
				})
			}
			rawNodes[len(rawNodes)-1].Attrs["value"] += string(chars[i])
			if chars[i] == '&' {
				buildSpecialChar = true
				specialChar = ""
			} else if buildSpecialChar {
				if chars[i] == ';' {
					buildSpecialChar = false
					if sChar, ok := consts.SpecialChars[specialChar]; ok {
						content := rawNodes[len(rawNodes)-1].Attrs["value"]
						rawNodes[len(rawNodes)-1].Attrs["value"] = content[:len(content)-len(specialChar)-2] + sChar
					}
				} else {
					specialChar += string(chars[i])
				}
			}
		}
	}
	if len(openedTags) > 0 {
		return nil, fmt.Errorf("unclosed tag %s", openedTags[len(openedTags)-1])
	}
	if len(rawNodes) > 0 {
		dom = append(dom, mergeNodes(rawNodes)...)
	}
	return dom, nil
}
