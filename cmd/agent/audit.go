package main

import (
	"context"
	"fmt"
	"os"

	"id.azz/pow/internal/llm"
	"id.azz/pow/internal/storage"
	"id.azz/pow/pkg/logger"
)

func audit(ctx context.Context) error {
	if err := logger.New(os.Getenv("DEBUG_MODE") == "true", os.Getenv("LOG_FOLDER")); err != nil {
		return err
	}
	defer logger.Close()

	db, err := storage.InitDB("pow.db")
	if err != nil {
		return err
	}
	defer db.Close()

	// GetActiveRulesText, err := db.GetActiveRulesText(ctx)
	// if err != nil {
	// 	return err
	// }

	// 	userPrompt := fmt.Sprintf(`Audit my PC activities for %s.
	// Acceptance criteria for today:
	// %s
	// Use provided tools to gather evidence, evaluate against the criteria then save the audit results using save_daily_audit and summarize what I accomplished in bullet points.`,
	// 		time.Now().Format("20060102"), GetActiveRulesText)

	// logger.Println(userPrompt)

	var maxStep = 5
	agentClient, err := llm.NewClient(os.Getenv("OPENAI_BASEURL"), os.Getenv("OPENAI_API_KEY"))
	if err != nil {
		return err
	}

	var result *llm.AIResponse
	var tools []llm.Message
	for step := 0; step < maxStep; step++ {
		if result != nil {
			switch result.Choices[0].FinishReason {
			case "tool_calls":
				toolUsage, err := agentClient.ToolCallService(ctx, result.Choices[0])
				if err != nil {
					return err
				}
				tools = toolUsage
				logger.Printf("string tool builder: %s", toolUsage)
			case "stop":
				fmt.Printf("Audit result:\n %s", result.Choices[0].Message)
				return nil
			}
		}

		if step < maxStep || len(tools) > 0 {
			result, err = agentClient.CallAIService(ctx, "summary what last 90 commit i made", tools...)
			if err != nil {
				return err
			}
			logger.Printf("Step: %d | Completions: %v", step, tools)
		} else {
			logger.Fatalf("fatal error: process needs too many tools")
			return nil
		}
	}

	return nil
}
