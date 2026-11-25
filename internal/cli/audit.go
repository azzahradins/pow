package cli

import (
	"context"
	"fmt"
	"os"

	"id.azz/pow/internal/llm"
	"id.azz/pow/internal/storage"
	"id.azz/pow/pkg/logger"
	"id.azz/pow/pkg/utility"
)

func RunAudit(ctx context.Context) error {

	db, err := storage.InitDB("pow.db")
	if err != nil {
		return err
	}
	defer db.Close()

	// GetActiveRulesText, err := db.GetActiveRulesText(ctx)
	// if err != nil {
	// 	return err
	// }

	var maxStep = 5
	agentClient, err := llm.NewClient(os.Getenv("OPENAI_BASEURL"), os.Getenv("OPENAI_API_KEY"))
	if err != nil {
		return err
	}

	var result *llm.AIResponse
	var chatHistory []llm.Message
	// initial prompt:
	systemPrompt, err := utility.LoadPrompt("prompts/system.md")
	if err != nil {
		return err
	}

	userPrompt, err := utility.LoadPromptVariables("prompts/activity_analysis.md", os.Getenv("GIT_EMAIL"))
	if err != nil {
		return err
	}

	chatHistory = append(chatHistory, llm.Message{
		Role:    "system",
		Content: systemPrompt,
	}, llm.Message{
		Role:    "user",
		Content: userPrompt,
	})

	// llm -> action steps
	for step := range maxStep {
		logger.Printf("step: %d", step)

		if result != nil {
			chatHistory = append(chatHistory, result.Choices[0].Message)
			switch result.Choices[0].FinishReason {
			case "tool_calls":
				toolUsage, err := agentClient.ToolCallService(ctx, result.Choices[0])
				if err != nil {
					return err
				}
				chatHistory = append(chatHistory, toolUsage...)
				logger.Printf("string tool builder: %s", toolUsage)
			case "stop":
				fmt.Printf("Audit result:\n %s", result.Choices[0].Message.Content)
				return nil
			}
		}

		result, err = agentClient.CallAIService(ctx, chatHistory...)
		if err != nil {
			return err
		}
	}

	return nil
}
