package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	if args.Limit <= 0 || args.Limit > 100 {
		return nil, errors.New("limit must be between 1 and 50")
	}

	cmd := exec.CommandContext(
		ctx,
		"git",
		"-C",
		"D:\\project\\thesis",
		"log",
		"-n",
		strconv.Itoa(args.Limit),
		"--format=%H%x1f%ad%x1f%s",
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

		fields := strings.SplitN(line, "\x1f", 3)
		if len(fields) != 3 {
			return nil, errors.New("invalid git log output")
		}

		result.Commits = append(result.Commits, GitCommit{
			Hash:    fields[0],
			Date:    fields[1],
			Message: fields[2],
		})
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return resultJSON, nil
}
