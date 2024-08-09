package logging

import (
	"TLExtractor/consts"
	"TLExtractor/logging/types"
	"TLExtractor/utils"
	"fmt"
	"github.com/fatih/color"
	"golang.org/x/term"
	"log"
	"math"
	"regexp"
	"strings"
)

const (
	FixedTimeStampWidth = 19
	TagWidth            = 8
	PackageWidth        = 15
)

var (
	debugLevel = types.LogLevelInfo{
		Icon:       'D',
		Background: color.New(color.ResetBold, 48, 2, 47, 93, 119),
		Foreground: color.New(color.ResetBold, 38, 2, 187, 187, 187),
		TextColor:  color.New(color.Reset, 38, 2, 39, 153, 153),
	}
	infoLevel = types.LogLevelInfo{
		Icon:       'I',
		Background: color.New(color.ResetBold, 48, 2, 105, 135, 89),
		Foreground: color.New(color.Bold, 38, 2, 233, 245, 230),
		TextColor:  color.New(color.Reset, 38, 2, 171, 192, 34),
	}
	warnLevel = types.LogLevelInfo{
		Icon:       'W',
		Background: color.New(color.ResetBold, 48, 2, 187, 181, 39),
		TextColor:  color.New(color.Reset, 38, 2, 187, 181, 39),
	}
	errorLevel = types.LogLevelInfo{
		Icon:       'E',
		Background: color.New(color.ResetBold, 48, 2, 207, 91, 86),
		TextColor:  color.New(color.Reset, 38, 2, 207, 91, 86),
	}
)

func internalLog(levelInfo types.LogLevelInfo, fatal bool, message ...any) {
	var errMessage string
	for _, x := range message {
		switch x.(type) {
		case error:
			errMessage += x.(error).Error() + " "
		case string:
			errMessage += x.(string) + " "
		default:
			if utils.IsNil(x) {
				continue
			}
			errMessage += fmt.Sprintf("%v", message) + " "
		}
	}
	if len(errMessage) == 0 {
		return
	}
	errMessage = strings.TrimSpace(errMessage)
	if levelInfo.Foreground == nil {
		levelInfo.Foreground = color.New(color.Bold, 38, 2, 0, 0, 0)
	}
	textColor := levelInfo.TextColor.SprintFunc()
	fileColor := color.New(color.Underline, 38, 2, 97, 175, 225).SprintFunc()
	mainColor := color.New(color.Bold, 38, 2, 171, 145, 186).SprintFunc()
	classColor := color.New(color.Bold, 38, 2, 98, 198, 183).SprintFunc()
	printFunc := log.Println
	if fatal {
		printFunc = log.Fatal
	}
	var mainDetails *types.CallerInfo
	startSkips := 2
	for {
		details, err := getInfo(startSkips)
		if err == nil {
			mainDetails = details
			break
		}
		if startSkips > 10 {
			break
		}
		startSkips++
	}

	matches := regexp.MustCompile(`^(([[:lower:]^:]{2,10}): )?([\S \n]+)`).FindStringSubmatch(errMessage)
	class := utils.Capitalize(matches[2])
	if len(class) == 0 {
		class = strings.Split(mainDetails.PackageName, ".")[0]
	}
	if class == "main" {
		class = utils.Capitalize(class)
		classColor = mainColor
	}
	termWidth, _, _ := term.GetSize(0)
	termWidth = int(math.Max(float64(termWidth), consts.MinTermWidth))
	availableTagWidth := int(math.Max(float64(termWidth/100*TagWidth)-3, 3))
	TagSpaces := strings.Repeat(" ", int(math.Max(float64(availableTagWidth-len(class)), 0)))
	if len(class) >= availableTagWidth {
		class = class[:availableTagWidth-3] + "..."
		TagSpaces = " "
	}
	availablePackageWidth := int(math.Max(float64(termWidth/100*PackageWidth)-3, 3))
	PackageSpaces := strings.Repeat(" ", int(math.Max(float64(availablePackageWidth-len(mainDetails.PackageName)), 0)))
	if len(mainDetails.PackageName) > availablePackageWidth {
		mainDetails.PackageName = mainDetails.PackageName[:availablePackageWidth-3] + "..."
		PackageSpaces = " "
	}
	totalIndent := FixedTimeStampWidth + len(class) + len(TagSpaces) + len(mainDetails.PackageName) + len(PackageSpaces) + 5
	errMess := strings.Join(strings.Split(matches[3], "\n"), "\n"+strings.Repeat(" ", totalIndent))
	description := fmt.Sprintf(
		"%s%s%s%s%s %s",
		classColor(class),
		TagSpaces,
		mainDetails.PackageName,
		PackageSpaces,
		levelInfo.Background.SprintFunc()(levelInfo.Foreground.SprintFunc()(fmt.Sprintf(" %c ", levelInfo.Icon))),
		textColor(utils.Capitalize(errMess)),
	)
	var lines string
	skips := startSkips - 1
	if fatal {
		for {
			skips++
			subDetails, runtimeErr := getInfo(skips)
			if runtimeErr != nil {
				break
			}
			if mainDetails.PackageName != subDetails.PackageName {
				subDetails.FuncName = subDetails.PackageName + "." + subDetails.FuncName
			}
			lines += fmt.Sprintf(
				"\n%s%s%s%s",
				strings.Repeat(" ", totalIndent+2),
				textColor(
					fmt.Sprintf(
						"at %s(",
						subDetails.FuncName,
					),
				),
				fileColor(
					fmt.Sprintf(
						"\u001B]8;;%s\u001B\\%s:%d\033]8;;\033\\",
						fmt.Sprintf("%s:%d", subDetails.FilePath, subDetails.Line),
						subDetails.FileName,
						subDetails.Line,
					),
				),
				textColor(")"),
			)
		}
	}
	printFunc(
		fmt.Sprintf(
			"%s%s",
			description,
			lines,
		),
	)
}
