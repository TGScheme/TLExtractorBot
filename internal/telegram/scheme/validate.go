package scheme

import (
	"fmt"
	"hash/crc32"
	"regexp"
	"strconv"
	"strings"

	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
)

type Problem struct {
	File   string
	Line   int
	Source string
	Reason string
	Fatal  bool
}

func (p Problem) String() string {
	return fmt.Sprintf("%s:%d: %s (%s)", p.File, p.Line, p.Reason, p.Source)
}

func FatalProblems(problems []Problem) []Problem {
	var fatal []Problem
	for _, problem := range problems {
		if problem.Fatal {
			fatal = append(fatal, problem)
		}
	}
	return fatal
}

type tlObject struct {
	file     string
	line     int
	source   string
	name     string
	id       string
	generics map[string]bool
	params   []types.Parameter
	result   string
	isMethod bool
}

var (
	tlNameRe        = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*(\.[a-zA-Z][a-zA-Z0-9_]*)?$`)
	tlIDRe          = regexp.MustCompile(`^[0-9a-fA-F]{1,8}$`)
	tlGenericRe     = regexp.MustCompile(`^\{([a-zA-Z][a-zA-Z0-9_]*):Type\}$`)
	tlFlagRe        = regexp.MustCompile(`^(flags\d*)\.(\d+)\?(.+)$`)
	tlOptionalTrue  = regexp.MustCompile(` \w+:\w+\.\d+\?true`)
	tlLayerSuffixRe = regexp.MustCompile(`^(.+?)\d+$`)
)

var tlBuiltinTypes = map[string]bool{
	"Type": true, "Object": true, "Vector": true, "True": true,
}

type TLFile struct {
	Name string
	Text string
}

func Validate(scheme *types.TLFullScheme) []Problem {
	return ValidateFiles([]TLFile{
		{Name: "main_api.tl", Text: ToString(scheme.MainApi, scheme.Layer, false)},
		{Name: "e2e.tl", Text: ToString(scheme.E2EApi, scheme.Layer, false)},
	})
}

func ValidateFiles(files []TLFile) []Problem {
	var problems []Problem
	var objects []tlObject
	for _, file := range files {
		if strings.TrimSpace(file.Text) == "" {
			continue
		}
		parsed, parseProblems := parseTLText(file.Name, file.Text)
		objects = append(objects, parsed...)
		problems = append(problems, parseProblems...)
		problems = append(problems, checkDuplicates(parsed)...)
	}
	return append(problems, checkObjects(objects)...)
}

func checkDuplicates(objects []tlObject) []Problem {
	var problems []Problem
	byID := map[string]tlObject{}
	byName := map[string]tlObject{}
	for _, object := range objects {
		if previous, clash := byID[object.id]; clash {
			problems = append(problems, object.problem(true, fmt.Sprintf(
				"constructor id #%s is already used by %s at line %d",
				object.id, previous.name, previous.line,
			)))
		} else {
			byID[object.id] = object
		}
		if previous, clash := byName[object.name]; clash {
			problems = append(problems, object.problem(true, fmt.Sprintf(
				"object name %s is already defined at line %d", object.name, previous.line,
			)))
		} else {
			byName[object.name] = object
		}
	}
	return problems
}

func parseTLText(file, text string) ([]tlObject, []Problem) {
	var objects []tlObject
	var problems []Problem
	isMethod := false
	for i, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		switch line {
		case "---types---":
			isMethod = false
			continue
		case "---functions---":
			isMethod = true
			continue
		}
		object, err := parseTLLine(line)
		if err != "" {
			problems = append(problems, Problem{File: file, Line: i + 1, Source: line, Reason: err, Fatal: true})
			continue
		}
		object.file = file
		object.line = i + 1
		object.source = line
		object.isMethod = isMethod
		objects = append(objects, object)
	}
	return objects, problems
}

func parseTLLine(line string) (tlObject, string) {
	object := tlObject{generics: map[string]bool{}}
	if !strings.HasSuffix(line, ";") {
		return object, "missing the trailing semicolon"
	}
	body := strings.TrimSpace(strings.TrimSuffix(line, ";"))
	separator := strings.LastIndex(body, "=")
	if separator < 0 {
		return object, "missing the result type"
	}
	object.result = strings.TrimSpace(body[separator+1:])
	if object.result == "" {
		return object, "missing the result type"
	}
	fields := strings.Fields(strings.TrimSpace(body[:separator]))
	if len(fields) == 0 {
		return object, "missing the object name"
	}
	name, id, found := strings.Cut(fields[0], "#")
	if !found {
		return object, "missing the constructor id"
	}
	object.name, object.id = name, id
	if !tlNameRe.MatchString(name) {
		return object, fmt.Sprintf("invalid object name %q", name)
	}
	if !tlIDRe.MatchString(id) {
		return object, fmt.Sprintf("invalid constructor id %q", id)
	}
	if parsed, err := strconv.ParseUint(id, 16, 32); err == nil && parsed == 0 {
		return object, "constructor id is zero"
	}
	inArray := false
	for _, field := range fields[1:] {
		if generic := tlGenericRe.FindStringSubmatch(field); generic != nil {
			object.generics[generic[1]] = true
			continue
		}
		switch field {
		case "[":
			inArray = true
			continue
		case "]":
			inArray = false
			continue
		case "#":
			continue
		}
		if inArray {
			continue
		}
		paramName, paramType, valid := strings.Cut(field, ":")
		if !valid || paramName == "" || paramType == "" {
			return object, fmt.Sprintf("invalid parameter %q", field)
		}
		object.params = append(object.params, types.Parameter{Name: paramName, Type: paramType})
	}
	return object, ""
}

func checkObjects(objects []tlObject) []Problem {
	var problems []Problem
	defined := map[string]bool{}
	for _, object := range objects {
		if object.isMethod {
			continue
		}
		for _, atom := range typeAtoms(object.result) {
			defined[atom] = true
		}
	}
	for _, object := range objects {
		problems = append(problems, object.checkParams(defined)...)
		for _, atom := range typeAtoms(object.result) {
			if !isKnownType(atom, defined, object.generics) {
				problems = append(problems, object.problem(false, fmt.Sprintf("unknown result type %q", atom)))
			}
		}
		if !object.matchesInferredID() {
			problems = append(problems, object.problem(false, fmt.Sprintf(
				"constructor id #%s does not match the declaration of %s", object.id, object.name,
			)))
		}
	}
	return problems
}

func (object tlObject) matchesInferredID() bool {
	declared, err := strconv.ParseUint(object.id, 16, 32)
	if err != nil {
		return true
	}
	body := strings.TrimSpace(strings.TrimSuffix(object.source, ";"))
	name, rest, found := strings.Cut(body, "#")
	if !found {
		return true
	}
	if _, rest, found = strings.Cut(rest, " "); !found {
		rest = ""
	}
	if inferID(name, rest) == uint32(declared) {
		return true
	}
	if stripped := tlLayerSuffixRe.FindStringSubmatch(name); stripped != nil {
		return inferID(stripped[1], rest) == uint32(declared)
	}
	return false
}

func inferID(name, params string) uint32 {
	return inferIDFromText(strings.TrimRight(name+" "+params, " "))
}

func inferIDFromText(representation string) uint32 {
	representation = strings.ReplaceAll(representation, ":bytes ", ":string ")
	representation = strings.ReplaceAll(representation, "?bytes ", "?string ")
	representation = strings.ReplaceAll(representation, "<", " ")
	representation = strings.ReplaceAll(representation, ">", "")
	representation = strings.ReplaceAll(representation, "{", "")
	representation = strings.ReplaceAll(representation, "}", "")
	return crc32.ChecksumIEEE([]byte(tlOptionalTrue.ReplaceAllString(representation, "")))
}

func (object tlObject) checkParams(defined map[string]bool) []Problem {
	var problems []Problem
	flagFields := map[string]bool{}
	reported := map[string]bool{}
	seen := map[string]bool{}
	for _, param := range object.params {
		if seen[param.Name] {
			problems = append(problems, object.problem(true, fmt.Sprintf("duplicated parameter %q", param.Name)))
		}
		seen[param.Name] = true
		if param.Type == "#" {
			flagFields[param.Name] = true
			continue
		}
		paramType := param.Type
		if flag := tlFlagRe.FindStringSubmatch(paramType); flag != nil {
			if !flagFields[flag[1]] && !reported[flag[1]] {
				reported[flag[1]] = true
				problems = append(problems, object.problem(true, fmt.Sprintf(
					"parameter %q depends on %q, which is not declared before it", param.Name, flag[1],
				)))
			}
			if bit, err := strconv.Atoi(flag[2]); err != nil || bit > 31 {
				problems = append(problems, object.problem(true, fmt.Sprintf(
					"parameter %q uses the out of range flag bit %s", param.Name, flag[2],
				)))
			}
			paramType = flag[3]
		}
		for _, atom := range typeAtoms(paramType) {
			if !isKnownType(atom, defined, object.generics) {
				problems = append(problems, object.problem(false, fmt.Sprintf(
					"parameter %q has the unknown type %q", param.Name, atom,
				)))
			}
		}
	}
	return problems
}

func (object tlObject) problem(fatal bool, reason string) Problem {
	return Problem{File: object.file, Line: object.line, Source: object.source, Reason: reason, Fatal: fatal}
}

func typeAtoms(rawType string) []string {
	var atoms []string
	for _, field := range strings.Fields(rawType) {
		field = strings.TrimLeft(field, "!%")
		for strings.HasPrefix(field, "Vector<") && strings.HasSuffix(field, ">") {
			atoms = append(atoms, "Vector")
			field = strings.TrimLeft(field[len("Vector<"):len(field)-1], "!%")
		}
		if field != "" {
			atoms = append(atoms, field)
		}
	}
	return atoms
}

func isKnownType(atom string, defined map[string]bool, generics map[string]bool) bool {
	return tlPrimitives[atom] || tlBuiltinTypes[atom] || defined[atom] || generics[atom]
}
