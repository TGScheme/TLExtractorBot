package logging

import (
	"TLExtractor/consts"
	"TLExtractor/logging/types"
	"fmt"
	"path"
	"runtime"
	"strconv"
	"strings"
)

func getInfo(skips int) (*types.CallerInfo, error) {
	var callerInfo types.CallerInfo
	pc, file, line, _ := runtime.Caller(skips + 1)
	callerInfo.Line = line
	callerInfo.FileName = path.Base(file)
	callerInfo.FuncName = runtime.FuncForPC(pc).Name()
	if strings.HasPrefix(callerInfo.FuncName, "runtime.") {
		return nil, fmt.Errorf("runtime function")
	}
	funcInfo := consts.GetFunctionInfoRgx.FindStringSubmatch(callerInfo.FuncName)
	callerInfo.PackageName = strings.ReplaceAll(funcInfo[1], "/", ".")
	callerInfo.FilePath = path.Join(path.Join(strings.Split(funcInfo[1], "/")[1:]...), path.Base(file))
	callerInfo.FuncName = funcInfo[2]
	if lambdaMatches := consts.LambdaNameRgx.FindAllStringSubmatch(callerInfo.FuncName, -1); len(lambdaMatches) > 0 {
		lambdaDetails, _ := getInfo(skips + 2)
		numFunc, _ := strconv.Atoi(lambdaMatches[0][1])
		callerInfo.FuncName = fmt.Sprintf("lambda$%s$%d", lambdaDetails.FuncName, numFunc-1)
	}
	callerInfo.FuncName = strings.ReplaceAll(callerInfo.FuncName, ".", "")
	return &callerInfo, nil
}
