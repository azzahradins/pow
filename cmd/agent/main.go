package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"id.azz/pow/internal/cli"
	"id.azz/pow/pkg/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := logger.New(os.Getenv("DEBUG_MODE") == "true", os.Getenv("LOG_FOLDER")); err != nil {
		panic("System failed to logging | implication: missing report & usage history")
	}
	defer logger.Close()

	if len(os.Args) < 1 {
		cli.PrintHelp()
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "audit":
		err = cli.RunAudit(ctx)
	case "help", "-h", "--help":
		cli.PrintHelp()
		return
	case "init":
		err = cli.RunInit(ctx, os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		cli.PrintHelp()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
