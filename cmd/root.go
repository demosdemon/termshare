package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	Name        = `termshare`
	colorGreen  = `32`
	colorYellow = `33`
	colorRed    = `31`
)

var root = cobra.Command{
	Use: Name,
}

func AddCommand(cmds ...*cobra.Command) {
	root.AddCommand(cmds...)
}

func Execute(args []string) error {
	root.SetArgs(args)
	return root.Execute()
}

func colorize(code, text string, v []interface{}) string {
	text = fmt.Sprintf(text, v...)
	return fmt.Sprintf("\x1b[0;%sm%s: %s\x1b[0m", code, Name, text)
}

func Info(text string, v ...interface{}) string {
	return colorize(colorGreen, text, v)
}

func Warning(text string, v ...interface{}) string {
	return colorize(colorYellow, text, v)
}

func Error(text string, v ...interface{}) string {
	return colorize(colorRed, text, v)
}
