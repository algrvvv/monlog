package log

import (
	"fmt"
	"log"
	"os"
)

const (
	levelInfo  = "[\033[38;5;2mINFO\033[0m]"
	levelDebug = "[\033[38;5;33mDEBUG\033[0m]"
	levelWarn  = "[\033[38;5;214mWARN\033[0m]"
	levelError = "[\033[38;5;9mERROR\033[0m]"
)

func printLog(level string, args ...any) {
	log.SetPrefix(level + " ")
	log.Println(args...)
	log.SetPrefix("")
}

func PrintInfof(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	printLog(levelInfo, message)
}

func PrintInfo(args ...any) {
	printLog(levelInfo, args...)
}

func PrintFatal(args ...any) {
	printLog(levelError, args...)
	os.Exit(1)
}

func PrintFatalf(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	printLog(levelError, message)
	os.Exit(1)
}

func PrintError(args ...any) {
	printLog(levelError, args...)
}

func PrintErrorf(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	printLog(levelError, message)
}

func PrintWarn(args ...any) {
	printLog(levelWarn, args...)
}

func PrintWarnf(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	printLog(levelWarn, message)
}
