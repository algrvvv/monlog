package utils

func ReaderCallback() func([]byte) []string {
	var lines []string
	return func(data []byte) []string {
		lines = append(lines, string(data))
		return lines
	}
}
