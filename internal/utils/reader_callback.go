package utils

import drivers "github.com/algrvvv/monlog/internal/drivers/registry"

func ReaderCallback(driver string) func([]byte) []string {
	var lines []string
	return func(data []byte) []string {
		line := drivers.Handle(driver, string(data))
		lines = append(lines, line)
		return lines
	}
}
