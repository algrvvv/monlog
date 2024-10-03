package logger

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/lmittmann/tint"
	"github.com/mdobak/go-xerrors"
	slogmulti "github.com/samber/slog-multi"
)

type stackFrame struct {
	Func   string `json:"func"`
	Source string `json:"source"`
	Line   int    `json:"line"`
}

func marshalStack(err error) []stackFrame {
	trace := xerrors.StackTrace(err)
	if len(trace) == 0 {
		return nil
	}

	frames := trace.Frames()
	s := make([]stackFrame, len(frames))

	for i, v := range frames {
		f := stackFrame{
			Source: filepath.Join(
				filepath.Base(filepath.Dir(v.File)),
				filepath.Base(v.File),
			),
			Func: filepath.Base(v.Function),
			Line: v.Line,
		}

		s[i] = f
	}

	return s
}

func fmtError(err error) slog.Value {
	var groupValues []slog.Attr
	groupValues = append(groupValues, slog.String("msg", err.Error()))
	frames := marshalStack(err)

	if frames != nil {
		groupValues = append(groupValues, slog.Any("trace", frames))
	}

	return slog.GroupValue(groupValues...)
}

func replaceAttr(_ []string, a slog.Attr) slog.Attr {
	// nolint
	switch a.Value.Kind() {
	case slog.KindAny:
		switch v := a.Value.Any().(type) {
		case error:
			a.Value = fmtError(v)
		}
	case slog.KindTime:
		t := a.Value.Time()
		a.Value = slog.StringValue(t.Format("02.01.06 15:04:05"))
	}

	return a
}

func tintReplaceAttr(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.LevelKey {
		// Определяем цвет для каждого уровня
		switch a.Value.String() {
		case "DEBUG":
			a.Value = slog.StringValue("\033[38;5;33mDEBUG\033[0m") // Синий для DEBUG
		case "INFO":
			a.Value = slog.StringValue("\033[38;5;2mINFO\033[0m") // Зеленый для INFO
		case "WARN":
			a.Value = slog.StringValue("\033[38;5;214mWARN\033[0m") // Оранжевый для WARN
		case "ERROR":
			a.Value = slog.StringValue("\033[38;5;9mERROR\033[0m") // Красный для ERROR
		}
	}
	return a
}

var logger *slog.Logger

func NewLogger(logfile string, debugLvl bool) error {
	file, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	lvl := new(slog.LevelVar)
	if debugLvl {
		lvl.Set(slog.LevelDebug)
	}

	opts := &slog.HandlerOptions{
		ReplaceAttr: replaceAttr,
		Level:       lvl,
	}

	fileHandler := slog.NewJSONHandler(file, opts)
	consoleHandler := tint.NewHandler(os.Stdout, &tint.Options{
		Level:       lvl,
		TimeFormat:  "02.01.06 15:04:05",
		ReplaceAttr: tintReplaceAttr,
	})

	buildinfo, _ := debug.ReadBuildInfo()

	log := slog.New(slogmulti.Fanout(
		fileHandler, consoleHandler,
	))

	logger = log.With(
		slog.Group("program_info",
			slog.Int("pid", os.Getpid()),
			slog.String("go_version", buildinfo.GoVersion),
		),
	)

	return nil
}

func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}

func Warn(msg string, args ...any) {
	logger.Warn(msg, args...)
}

func Error(msg string, err error, args ...any) {
	e := xerrors.New(err)
	args = append(args, slog.Any("error", e))
	logger.Error(msg, args...)
}

func Fatal(msg string, err error, args ...any) {
	e := xerrors.New(err)
	args = append(args, slog.Any("error", e))
	logger.Error(msg, args...)
	os.Exit(1)
}

func Infof(message string, args ...any) {
	msg := fmt.Sprintf(message, args...)
	logger.Info(msg)
}

func Debugf(message string, args ...any) {
	msg := fmt.Sprintf(message, args...)
	logger.Debug(msg)
}

func Warnf(message string, args ...any) {
	msg := fmt.Sprintf(message, args...)
	logger.Warn(msg)
}

func Errorf(message string, args ...any) {
	msg := fmt.Sprintf(message, args...)
	logger.Error(msg)
}

func Fatalf(message string, args ...any) {
	msg := fmt.Sprintf(message, args...)
	logger.Error(msg)
	os.Exit(1)
}
