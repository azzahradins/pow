package utility

import (
	"fmt"
	"os"
	"strings"
)

func LoadPrompt(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(content)), nil
}

func LoadPromptVariables(path string, param ...string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	formatPrompt := fmt.Sprintf(string(content), param)

	return strings.TrimSpace(formatPrompt), nil
}
