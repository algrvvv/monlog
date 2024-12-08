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

func print(level string, args ...any) {
	log.SetPrefix(level + " ")
	log.Println(args...)
	log.SetPrefix("")
}

func PrintInfof(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	print(levelInfo, message)
}

func PrintInfo(args ...any) {
	print(levelInfo, args...)
}

func PrintFatal(args ...any) {
	print(levelError, args...)
	os.Exit(1)
}

func PrintFatalf(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	print(levelError, message)
	os.Exit(1)
}

func PrintError(args ...any) {
	print(levelError, args...)
}

func PrintErrorf(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	print(levelError, message)
}

func PrintWarn(args ...any) {
	print(levelWarn, args...)
}

func PrintWarnf(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	print(levelWarn, message)
}
