package state

import (
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"github.com/algrvvv/monlog/internal/config"
	"github.com/algrvvv/monlog/internal/logger"
	"github.com/algrvvv/monlog/internal/notify"
	"github.com/algrvvv/monlog/internal/utils"
)

func generateDateTimeRegex(format string) string {
	replacements := map[string]string{
		"YYYY": `\d{4}`,
		"YY":   `\d{2}`,
		"MM":   `\d{2}`,
		"DD":   `\d{2}`,
		"hh":   `\d{2}`,
		"mm":   `\d{2}`,
		"ss":   `\d{2}`,
	}

	for key, value := range replacements {
		format = strings.ReplaceAll(format, key, value)
	}

	format = regexp.MustCompile(`[-/.:\s]`).ReplaceAllString(format, `\$0`)

	return fmt.Sprintf("(?P<TIME>%s)", format)
}

func generateRegexFromLayout(layout, timeFormat string) (*regexp.Regexp, error) {
	timeRegex := generateDateTimeRegex(timeFormat)

	replacements := map[string]string{
		"%TIME%":    timeRegex,
		"%LEVEL%":   `(?P<LEVEL>[A-Z]+)`,
		"%MESSAGE%": `(?P<MESSAGE>.+)`,
		"...":       `.*?`,
	}

	for key, value := range replacements {
		layout = strings.ReplaceAll(layout, key, value)
	}

	return regexp.Compile(layout)
}

func ParseLineAndSendNotify(sid int, line string) {
	if n := utils.ValidateServerID(strconv.Itoa(sid)); n == -1 {
		logger.Error("got invalid server id", nil)
		return
	}
	sl := config.Cfg.Servers[sid]
	regex, err := generateRegexFromLayout(sl.LogLayout, sl.LogTimeFormat)
	if err != nil {
		logger.Warn(
			"Failed to generate regex from layout: "+err.Error(),
			slog.Any("layout", sl.LogLayout),
		)
		return
	}

	matches := regex.FindStringSubmatch(line)
	if matches == nil {
		logger.Debug("No matches found for line: "+line, slog.Any("layout", sl.LogLayout),
			slog.Any("regex", regex))
		return
	}

	values := make(map[string]string)
	names := regex.SubexpNames()
	for i, match := range matches {
		if i > 0 && i < len(names) && match != "" {
			values[names[i]] = match
			// fmt.Printf("%s: %s\n", names[i], match)
		}
	}

	if (sl.Notify != "" && sl.Notify != "none") && isAlertLevel(values["LEVEL"], sl.LogLevel) &&
		CompareLastNotifyTime(sl.ID, values["TIME"]) {
		notification := &notify.Notification{
			Server: &sl, Log: line, Message: values["MESSAGE"],
			Time: values["TIME"], Level: values["LEVEL"],
		}

		if err = notify.Notify(notification); err != nil {
			logger.Errorf("Error sending notification: %s", err)
		}
		if err = UpdateLastNotifyTime(sl.ID, values["TIME"]); err != nil {
			logger.Error(
				"Error updating last notify time: "+err.Error(),
				err,
				slog.Any("time", values["TIME"]),
			)
		}
	}
}

func isAlertLevel(level string, alertLevels string) bool {
	alertLevelRegex := regexp.MustCompile(fmt.Sprintf("^(%s)$", alertLevels))
	return alertLevelRegex.MatchString(level)
}
