package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"id.azz/pow/pkg/logger"
	"unknwon.dev/ini/v2"
)

// go run pow init
func RunInit(ctx context.Context, args []string) (err error) {
	if gitActive := checkGitRepo(); gitActive == false {
		return fmt.Errorf("Error - This project does not tracked by any git provider")
	}
	logger.Println("Installation - start to track your project progress!")
	powDir := ".pow"
	configPath := filepath.Join(powDir, "config")
	createdDir := false

	// rollback panic or error happens -> delete file and folder
	defer rollbackOnError(&err, powDir, &createdDir)

	// check file already available
	if _, err := os.Stat(configPath); err == nil {
		logger.Println("Your pow already installed and ready to use!\nIf you want to change branch, modify the config file", configPath)
		return nil
	}

	// tool instance
	cfg := ini.Empty()
	info, _ := cfg.NewSection("information")
	scanner := bufio.NewScanner(os.Stdin)
	// 1: What project you are on?
	defaultProject := detectRepoName()
	fmt.Printf("⊗ Project name (%s) = ", defaultProject)
	scanner.Scan()
	projectName := strings.TrimSpace(scanner.Text())
	if projectName == "" {
		projectName = defaultProject
	}
	info.NewKey("project", projectName)

	// 2: project summary?
	fmt.Printf("⊗ Project summary (empty) = ")
	scanner.Scan()
	projectSummary := strings.TrimSpace(scanner.Text())
	info.NewKey("summary", projectSummary)

	// 3: choose tracking git remote or local?
	branchExist := false
	defaultBranch := detectCurrentBranch()

	for branchExist == false {
		fmt.Printf("⊗ branch? (%s) = ", defaultBranch)
		scanner.Scan()
		trackBranch := strings.TrimSpace(scanner.Text())
		if trackBranch == "" {
			trackBranch = defaultBranch
		}
		if branchExist = checkBranchExist(ctx, trackBranch); branchExist {
			info.NewKey("branch", trackBranch)
		} else {
			fmt.Println("Error: branch does not exist!")
		}
	}

	// save values to file
	if err := os.MkdirAll(powDir, 0755); err != nil {
		return fmt.Errorf("failed to create %s directory: %w", powDir, err)
	}
	if err := cfg.SaveTo(configPath); err != nil {
		return err
	}

	//adjust gitignore
	if err := ensureGitIgnore(".pow/*"); err != nil {
		return err
	}

	logger.Println("OK - Success pow initialization directory in ", powDir)
	return nil
}

func rollbackOnError(err *error, powDir string, createdDir *bool) {
	if *err != nil && *createdDir {
		_ = os.RemoveAll(powDir)
	}
}

func detectRepoName() string {
	fallback := "project-1"
	// 1. Try reading remote origin URL (e.g., https://github.com/user/my-repo.git -> my-repo)
	if out, err := exec.Command("git", "config", "--get", "remote.origin.url").Output(); err == nil {
		raw := strings.TrimSpace(string(out))
		if raw != "" {
			base := filepath.Base(raw)
			return strings.TrimSuffix(base, ".git")
		}
	}

	// 2. Fallback: Git top-level folder name
	if out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output(); err == nil {
		top := strings.TrimSpace(string(out))
		if top != "" {
			return filepath.Base(top)
		}
	}

	// 3. Final fallback: current folder name
	return fallback
}

func detectCurrentBranch() string {
	if out, err := exec.Command("git", "branch", "--show-current").Output(); err == nil {
		branch := strings.TrimSpace(string(out))
		if branch != "" {
			return branch
		}
	}
	return "main"
}

func checkBranchExist(ctx context.Context, name string) bool {
	// Check local branch: refs/heads/<name>
	localCmd := exec.CommandContext(ctx, "git", "show-ref", "--verify", "--quiet", "refs/heads/"+name)
	if localCmd.Run() == nil {
		return true
	}

	return false
}

func ensureGitIgnore(entry string) error {
	gitignorePath := ".gitignore"

	// 1. Read existing .gitignore if it exists
	contentBytes, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read .gitignore: %w", err)
	}

	content := string(contentBytes)

	// 2. Check if ".pow/" (or ".pow") is already ignored to avoid duplicates
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == entry || trimmed == strings.TrimSuffix(entry, "/") {
			return nil // Already ignored
		}
	}

	// 3. Open .gitignore in append/create mode and write ".pow/"
	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open .gitignore: %w", err)
	}
	defer f.Close()

	// Add a leading newline if the existing file doesn't end with one
	prefix := ""
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		prefix = "\n"
	}

	if _, err := f.WriteString(prefix + entry + "\n"); err != nil {
		return fmt.Errorf("failed to update .gitignore: %w", err)
	}

	return nil
}

func checkGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	return cmd.Run() == nil
}
