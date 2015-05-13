package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

const (
	LOG_CRIT = iota + 1
	LOG_INFO
	LOG_DEBUG
	LOG_TRACE
)

func Log(text string) {
	fmt.Println(text)
}

func formatHost (s string) string {
	return fmt.Sprintf("[%%-%ds] ", maxHostLength) + s
}

func LogInfo(host, text string) int {
	lenText := 0
	if logLevel < LOG_DEBUG {
		color.Set(color.FgGreen)
		logText := fmt.Sprintf(
			formatHost("# %s"),
			host,
			text,
		)
		lenText = len(logText)
		fmt.Printf(logText)
		color.Unset()
	}
	if logLevel >= LOG_DEBUG {
		fmt.Println("")
	}
	return lenText
}

func LogDebug(host, text string) {
	if logLevel >= LOG_DEBUG {
		color.Yellow(formatHost("%s\n"), host, text)
	}
}

func LogCmd(host, text string) {
	if logLevel >= LOG_DEBUG {
		color.Yellow(formatHost("$ %s\n"), host, text)
	}
}

func LogOut(host, text string) {
	if logLevel >= LOG_TRACE {
		color.White("[%s] %s\n", host, text)
	}
}

func LogError(host string, err error) {
	color.Red("[%s] %s\n", host, err)
}

func logProgress (l int) {
	fmt.Printf(" %s", strings.Repeat(".", 80 - l))
}

func LogOK(l int) {
	if logLevel < LOG_DEBUG {
		logProgress(l)
		color.Green("✔")
	}
}

func LogNG(l int) {
	if logLevel < LOG_DEBUG {
		logProgress(l)
		color.Red("✘")
	}
}

