package android

import (
	"TLExtractor/android/types"
	"TLExtractor/consts"
	"TLExtractor/java"
	javaTypes "TLExtractor/java/types"
	schemeTypes "TLExtractor/telegram/scheme/types"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

func extractParams(class *javaTypes.RawClass, declarationPos int) ([]schemeTypes.Parameter, error) {
	var params []schemeTypes.Parameter
	var openedFlags, fromIf, fromLoop, fromSmartFlag bool
	var flagNesting, forNesting int
	var addedFlags []string
	pendingFlags := make(map[string]types.FlagInfo)
	var flagName string
	flagValue := -1
	//fastCheck := regexp.MustCompile(`this\.\w+`)
	compileVars := regexp.MustCompile(`\(?(this|tLRPC[^.]+)\.([^. )]+)( \?|\.\w+Value\(\)|\.add|\.get|\.serialize|\)| !| = (Boolean\.valueOf\(abstractSerializedData|abstractSerializedData|inputSerializedData|i[0-9+]*;|read|TLdeserialize;|Vector\.deserialize|\([^(]|\w+\$\w+\.\w+deserialize))\)?`)
	compileVarBuffer := regexp.MustCompile(`^(this|tLRPC\$[^.]+)*\.*\w* *=* *((Boolean\.valueOf\()?(abstractSerializedData|inputSerializedData)[0-9]*|)?(\.write|\.read|TLRPC\$)([^(.]+).*?\);`)
	compileVarFlag := regexp.MustCompile(`this\.flags[0-9]* = readInt[0-9]+;`)
	compileVarBool := regexp.MustCompile(`this\.\w+ = \([^)]*readInt32[0-9]*[^)]*\)`)
	compileFlags := regexp.MustCompile(`([\w =]+[|& ][ (]|(TLRPC\$|TLObject\.)(setFlag|hasFlag)\((.*?),\s*)([0-9]+)`)
	compileFlagName := regexp.MustCompile(`flags[0-9]*`)
	compileIntegerFlagName := regexp.MustCompile(`i([0-9]*)[^a-zA-Z0-9_]`)
	compileUnVector := regexp.MustCompile(`Vector<(.*?)>`)
	compileUnknownVectorType := regexp.MustCompile(`\(\((.*?)\).*get`)
	dialogResolver := regexp.MustCompile(`DialogObject\..+\(`)
	for pos, line := range class.Content {
		if dialogResolver.MatchString(line.Line) {
			continue
		}
		if pos > declarationPos && declarationPos != 0 && line.Nesting >= 2 {
			if matches := compileFlags.FindAllStringSubmatch(line.Line, -1); len(matches) > 0 {
				flagNum, err := strconv.Atoi(matches[0][5])
				if err != nil {
					return nil, err
				}
				flagValue = int(math.Log2(float64(flagNum)))
				openedFlags = true
				flagNesting = line.Nesting

				linesToCheck := []string{
					line.Line,
					class.Content[pos+1].Line,
					class.Content[pos-1].Line,
				}

				for _, l := range linesToCheck {
					if names := compileFlagName.FindAllString(l, -1); names != nil {
						flagName = names[0]
						break
					}
				}

				if len(flagName) == 0 {
					for _, l := range linesToCheck {
						if rawNums := compileIntegerFlagName.FindAllStringSubmatch(l, -1); len(rawNums) > 0 {
							flagName = fmt.Sprintf("flags%s", rawNums[0][1])
							break
						}
					}
				}
				fromSmartFlag = matches[0][3] == "setFlag" || matches[0][3] == "hasFlag"

				if len(flagName) == 0 {
					if fromSmartFlag {
						//if _, err = strconv.Atoi(matches[0][4]); err == nil {
						//	flagName = "flags"
						//}
						flagName = "flags"
					}
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
				var parameter schemeTypes.Parameter
				var fromBuffer bool
				parameter.Name = matches[0][2]
				if matchedType := compileVarBuffer.FindAllStringSubmatch(line.Line, -1); len(matchedType) > 0 {
					parameter.Type = java.ParseType(matchedType[0][6])
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
				if parameter.Type == "ArrayList" {
					if unkTypeMatches := compileUnknownVectorType.FindAllStringSubmatch(line.Line, -1); len(unkTypeMatches) > 0 {
						parameter.Type = unkTypeMatches[0][1]
					}
				}
				parameter.Type, _ = java.FormatType(parameter.Type, true)
				if strings.HasPrefix(parameter.Type, "Vector") && !strings.Contains(line.Line, "Vector.serialize") && !strings.Contains(line.Line, "Vector.deserialize") && !fromLoop {
					if vectorRes := compileUnVector.FindAllStringSubmatch(parameter.Type, -1); len(matches) > 0 {
						parameter.Type = vectorRes[0][1]
					}
				} else if !strings.HasPrefix(parameter.Type, "Vector") && fromLoop {
					parameter.Type = fmt.Sprintf("Vector<%s>", parameter.Type)
				}
				class.Vars[parameter.Name] = parameter.Type
				if pendingFlag, ok := pendingFlags[parameter.Name]; openedFlags || ok {
					if ok {
						flagName = pendingFlag.Name
						flagValue = pendingFlag.Value
						delete(pendingFlags, parameter.Name)
					}
					if flagValue == -1 {
						return nil, consts.FlagNotFound
					}
					if !fromBuffer && parameter.Type == "Bool" {
						parameter.Type = "true"
					}

					if !fromIf && !fromSmartFlag && !slices.Contains([]string{"true", "#"}, parameter.Type) {
						pendingFlags[parameter.Name] = types.FlagInfo{
							Name:  flagName,
							Value: flagValue,
						}
						continue
					}
					if parameter.Name != flagName {
						parameter.Type = fmt.Sprintf("%s.%d?%s", flagName, flagValue, parameter.Type)
					}
					if !slices.Contains(addedFlags, flagName) {
						addedFlags = append(addedFlags, flagName)
						params = append(params, schemeTypes.Parameter{
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
				parameter.Type = fixParamType(parameter.Type)
				if duplicated := slices.IndexFunc(params, func(oldParameter schemeTypes.Parameter) bool {
					return oldParameter.Name == parameter.Name
				}); duplicated != -1 {
					params = append(params[:duplicated], params[duplicated+1:]...)
				}
				params = append(params, parameter)
			}
		}
		if pos > declarationPos && line.Nesting == 1 {
			break
		}
	}
	return params, nil
}
