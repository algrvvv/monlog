package logger

import (
	"fmt"
	"strings"

	"github.com/algrvvv/monlog/internal/config"
)

const banner = "\n███╗   ███╗ ██████╗ ███╗   ██╗██╗      ██████╗  ██████╗ \n████╗ ████║██╔═══██╗████╗  ██║██║     ██╔═══██╗██╔════╝ \n██╔████╔██║██║   ██║██╔██╗ ██║██║     ██║   ██║██║  ███╗\n██║╚██╔╝██║██║   ██║██║╚██╗██║██║     ██║   ██║██║   ██║\n██║ ╚═╝ ██║╚██████╔╝██║ ╚████║███████╗╚██████╔╝╚██████╔╝\n╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═══╝╚══════╝ ╚═════╝  ╚═════╝ \n                                                        \n"

func PrintLogo() {
	fmt.Print(banner)

	line := strings.Repeat("─", 51)
	top := fmt.Sprintf("┌%s┐", line)
	bottom := fmt.Sprintf("└%s┘", line)

	content := []string{
		"Monlog v0.9.7",
		fmt.Sprintf("http://127.0.0.1:%d", config.Cfg.App.Port),
		fmt.Sprintf("(bound on host 0.0.0.0 and port %d)", config.Cfg.App.Port),
	}

	fmt.Println(top)
	for _, line := range content {
		fmt.Printf("│%-51s│\n", center(line, 51))
	}

	fmt.Println(bottom)
}

func center(s string, width int) string {
	if len(s) >= width {
		return s
	}
	padLeft := (width - len(s)) / 2
	padRight := width - len(s) - padLeft
	return strings.Repeat(" ", padLeft) + s + strings.Repeat(" ", padRight)
}
