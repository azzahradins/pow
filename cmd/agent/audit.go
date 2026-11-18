package main

import (
	"context"
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

	agentClient, err := llm.NewClient(os.Getenv("OPENAI_BASEURL"), os.Getenv("OPENAI_API_KEY"))
	if err != nil {
		return err
	}

	result, err := agentClient.CallAIService(ctx, "detect how manyg commits i made")
	if err != nil {
		return err
	}

	logger.Printf("Audit result: %+v", result)

	return nil
}
