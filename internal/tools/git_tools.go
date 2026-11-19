package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type GitHistoryArgs struct {
	// Path  string `json:"path"`
	Limit int `json:"limit"`
}

type GitCommit struct {
	Hash    string `json:"hash"`
	Author  string `json:"author"`
	Date    string `json:"date"`
	Message string `json:"message"`
}

type GitHistoryResult struct {
	Commits []GitCommit `json:"commits"`
}

func ReadGitHistory(
	ctx context.Context,
	args GitHistoryArgs,
) ([]byte, error) {
	// 98d2579ef526f133440171a8adfcb57761cd3309^_2026-11-16^_init: project ideation
	if args.Limit <= 0 || args.Limit > 50 {
		return nil, errors.New("limit must be between 1 and 50")
	}

	// command: git -C . log -n 20 --format=%H%x1f%ae%x1f%ad%x1f%s --author=%whois% --date=short
	whois := fmt.Sprintf("--author=%s", os.Getenv("GIT_EMAIL"))
	cmd := exec.CommandContext(
		ctx,
		"git",
		"-C", ".", // directory
		"log", "-n", strconv.Itoa(args.Limit), // display commit history with limit
		"--format=%H%x1f%ae%x1f%ad%x1f%s",
		whois,
		"--date=short",
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to read git history: %w", err)
	}

	var result GitHistoryResult

	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line == "" {
			continue
		}

		fields := strings.SplitN(line, "\x1f", 4)
		if len(fields) != 4 {
			return nil, errors.New("invalid git log output")
		}

		result.Commits = append(result.Commits, GitCommit{
			Hash:    fields[0],
			Author:  fields[1],
			Date:    fields[2],
			Message: fields[3],
		})
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return resultJSON, nil
}
