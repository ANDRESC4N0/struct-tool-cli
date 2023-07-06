package logs

import (
	"fmt"

	"github.com/fatih/color"
)

func Error(msg string) {
	color.Red("==> Error %s", msg)
}

func Chapter(msg string) {
	color.Green("==> %s", msg)
}

func Echo(msg string) {
	fmt.Println(msg)
}
