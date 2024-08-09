package android

import (
	"TLExtractor/consts"
	"TLExtractor/environment"
	"TLExtractor/java"
	javaTypes "TLExtractor/java/types"
	"TLExtractor/logging"
	"TLExtractor/telegram/scheme/types"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

func extractParams(class *javaTypes.RawClass, declarationPos int) ([]types.Parameter, error) {
	var params []types.Parameter
	var openedFlags, fromIf, fromLoop bool
	var flagNesting, forNesting int
	var addedFlags []string
	var flagName string
	flagValue := -1
	fastCheck := regexp.MustCompile("this\\.\\w+")
	compileVars := regexp.MustCompile("\\(?(this|tLRPC[^.]+)\\.([^. ]+)( \\?|\\.add|\\.get|\\.serialize|\\)| !| = (abstractSerializedData|i[0-9+]*;|read|TLdeserialize;|\\(|\\w+\\$\\w+\\.\\w+deserialize))\\)?")
	compileVarBuffer := regexp.MustCompile("^(this|tLRPC\\$[^.]+)*\\.*\\w* *=* *(abstractSerializedData[0-9]*)?(\\.write|\\.read|TLRPC\\$)([^(.]+).*?\\);")
	compileVarFlag := regexp.MustCompile("this\\.flags[0-9]* = readInt[0-9]+;")
	compileVarBool := regexp.MustCompile("this\\.\\w+ = \\([^)]*readInt32[0-9]*[^)]*\\)")
	compileFlags := regexp.MustCompile("[\\w =]+[|& ][ (]([0-9]+)")
	compileFlagName := regexp.MustCompile("flags[0-9]*")
	compileUnVector := regexp.MustCompile("Vector<(.*?)>")
	for pos, line := range class.Content {
		if pos > declarationPos && declarationPos != 0 && line.Nesting >= 2 {
			if matches := compileFlags.FindAllStringSubmatch(line.Line, -1); len(matches) > 0 {
				flagNum, err := strconv.Atoi(matches[0][1])
				if err != nil {
					return nil, err
				}
				flagValue = int(math.Log2(float64(flagNum)))
				openedFlags = true
				flagNesting = line.Nesting
				if name := compileFlagName.FindAllString(line.Line, -1); name != nil {
					flagName = name[0]
				} else if name = compileFlagName.FindAllString(class.Content[pos+1].Line, -1); name != nil {
					flagName = name[0]
				} else if name = compileFlagName.FindAllString(class.Content[pos-1].Line, -1); name != nil {
					flagName = name[0]
				}
				if fromIf = strings.HasPrefix(line.Line, "if"); fromIf {
					flagNesting--
					continue
				} else if fromIf = strings.HasPrefix(line.Line, "boolean"); fromIf {
					continue
				}
			}
			if line.Nesting == flagNesting && openedFlags {
				if fromIf && strings.HasPrefix(line.Line, "}") {
					openedFlags = false
					flagValue = -1
				}
			}
			if line.Nesting == forNesting && fromLoop {
				fromLoop = false
			}
			if strings.HasPrefix(line.Line, "for") || strings.HasPrefix(line.Line, "while") {
				fromLoop = true
				forNesting = line.Nesting - 1
			}
			if matches := compileVars.FindAllStringSubmatch(line.Line, -1); len(matches) > 0 {
				var parameter types.Parameter
				var fromBuffer bool
				parameter.Name = matches[0][2]
				if matchedType := compileVarBuffer.FindAllStringSubmatch(line.Line, -1); len(matchedType) > 0 {
					parameter.Type = java.ParseType(matchedType[0][4])
					fromBuffer = true
				} else if declaredType, ok := class.Vars[matches[0][2]]; ok {
					parameter.Type = java.ParseType(declaredType)
				} else if compileVarFlag.MatchString(line.Line) {
					parameter.Type = "int"
				} else if compileVarBool.MatchString(line.Line) {
					parameter.Type = "bool"
				} else if strings.HasPrefix(matches[0][1], "tLRPC") {
					escapedVar := regexp.QuoteMeta(matches[0][1])
					compileReverseName := regexp.MustCompile(fmt.Sprintf("(%s =|this\\.)(\\w+)(;| = %s)", escapedVar, escapedVar))
					compileReverseType := regexp.MustCompile(fmt.Sprintf("(\\S+) %s =", escapedVar))
					for tries := 0; tries < 20 && pos >= tries; tries++ {
						currLine := class.Content[pos-tries].Line
						if varAssMatches := compileReverseName.FindAllStringSubmatch(currLine, -1); len(varAssMatches) > 0 {
							parameter.Name = varAssMatches[0][2]
						}
						if varTypeMatches := compileReverseType.FindAllStringSubmatch(currLine, -1); len(varTypeMatches) > 0 {
							parameter.Type = varTypeMatches[0][1]
							break
						}
					}
				} else {
					return nil, consts.UnknownType
				}
				if slices.Contains(addedFlags, parameter.Name) {
					continue
				}
				if strings.HasPrefix(parameter.Name, "flags") {
					parameter.Type = "#"
					addedFlags = append(addedFlags, parameter.Name)
				}
				class.Vars[parameter.Name] = parameter.Type
				parameter.Type, _ = java.FormatType(parameter.Type, true)
				if strings.HasPrefix(parameter.Type, "Vector") && !fromLoop {
					if vectorRes := compileUnVector.FindAllStringSubmatch(parameter.Type, -1); len(matches) > 0 {
						parameter.Type = vectorRes[0][1]
					}
				}
				if openedFlags {
					if flagValue == -1 {
						return nil, consts.FlagNotFound
					}
					if !fromBuffer && parameter.Type == "Bool" {
						parameter.Type = "true"
					}
					parameter.Type = fmt.Sprintf("%s.%d?%s", flagName, flagValue, parameter.Type)
					if !slices.Contains(addedFlags, flagName) {
						addedFlags = append(addedFlags, flagName)
						params = append(params, types.Parameter{
							Name: flagName,
							Type: "#",
						})
					}
					if line.Nesting == flagNesting {
						flagValue = -1
						openedFlags = false
					}
				}
				parameter.Name = fixParamName(parameter.Name)
				params = append(params, parameter)
			} else if fastCheck.MatchString(line.Line) && environment.Debug {
				logging.Debug("Fast check:", line.Line, class.FullName())
			}
		}
		if pos > declarationPos && line.Nesting == 1 {
			break
		}
	}
	return params, nil
}
