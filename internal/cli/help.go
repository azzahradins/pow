package cli

import "fmt"

func PrintHelp() {
	fmt.Println(`pow - command-line tool for analyzing activity in your system.
Usage:
	pow <command> [flags]

Commands:
	init	 Initialize file/repository tracking
	audit    Perform a daily audit
	help     Show all available commands

Run "pow <command> -h" for command-specific flags.`)
}
