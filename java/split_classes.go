package java

import (
	"TLExtractor/utils"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
)

func SplitClasses(className, content string, replaceClasses []string) map[string]string {
	lines := strings.Split(content, "\n")

	compileClass := regexp.MustCompile(`(^ {4}public static .*?class )(.*?) extends (.*?) {`)
	compileClassClose := regexp.MustCompile(`^ {4}}`)
	compileParentClasses := regexp.MustCompile(fmt.Sprintf(`(%s)\.(\w+)`, strings.Join(replaceClasses, "|")))
	compileClassInitializer := regexp.MustCompile(`(^ {12}.*?)(\w+)\$\w+ (\w+) = new \w+\$\w+\(\);`)
	compileClassInitializer2 := regexp.MustCompile(`(\w+)\$\w+ (\w+) = this\.\w+;`)

	var childName string
	opened := false
	classes := make(map[string][]string)

	for _, line := range lines {
		if matches := compileClass.FindAllStringSubmatch(line, -1); len(matches) > 0 {
			childName = fmt.Sprintf("%s$%s", className, matches[0][2])
			opened = true
		}
		if opened {
			classes[childName] = append(classes[childName], line)
		}
		if opened && compileClassClose.MatchString(line) {
			opened = false
		}
	}

	var classNames []string
	for name := range classes {
		classNames = append(classNames, strings.Split(name, "$")[1])
	}
	namesJoined := regexp.MustCompile(fmt.Sprintf(`(\b)(%s)(\b)`, strings.Join(classNames, "|")))
	var dynamicRegex *regexp.Regexp
	replaceNames := make(map[string]string)
	appendName := func(base, name string) {
		replaceNames[name] = fmt.Sprintf("%s$%s", strings.ToLower(base[:1])+base[1:], utils.Capitalize(name))
		dynamicRegex = regexp.MustCompile(fmt.Sprintf(`(^ {12}.*?)(%s)`, strings.Join(slices.Collect(maps.Keys(replaceNames)), "|")))
	}
	for name, classLines := range classes {
		dynamicRegex = nil
		replaceNames = make(map[string]string)
		for i, line := range classLines {
			line = compileParentClasses.ReplaceAllString(line, `$1$$$2`)
			if namesJoined.MatchString(line) {
				line = namesJoined.ReplaceAllString(line, fmt.Sprintf("${1}%s$$$2$3", className))
			}
			if matches := compileClassInitializer2.FindAllStringSubmatch(line, -1); len(matches) > 0 {
				appendName(matches[0][1], matches[0][2])
			}
			if matches := compileClassInitializer.FindAllStringSubmatch(line, -1); len(matches) > 0 {
				appendName(matches[0][2], matches[0][3])
			}
			if dynamicRegex != nil {
				if matches := dynamicRegex.FindAllStringSubmatch(line, -1); len(matches) > 0 {
					line = strings.ReplaceAll(line, matches[0][2], replaceNames[matches[0][2]])
				}
			}
			classes[name][i] = line
		}
	}
	files := make(map[string]string)
	for name, classLines := range classes {
		files[fmt.Sprintf("%s.java", name)] = strings.Join(classLines, "\n")
	}
	return files
}
