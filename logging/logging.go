package logging

func Debug(message ...any) {
	internalLog(debugLevel, false, message...)
}

func Info(message ...any) {
	internalLog(infoLevel, false, message...)
}

func Warn(message ...any) {
	internalLog(warnLevel, false, message...)
}

func Error(message ...any) {
	internalLog(errorLevel, false, message...)
}

func Fatal(message any) {
	internalLog(errorLevel, true, message)
}
