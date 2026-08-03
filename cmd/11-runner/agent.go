package main

import (
	"context"
	"fmt"
	"log"

	"github.com/siuyin/dflt"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model/gemini"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"
)

const (
	appName   = "my_first_adk_v2_app"
	userID    = "user_123"
	sessionID = "session_456"
)

func main() {

	modelName := dflt.EnvString("MODEL", "gemini-3.5-flash-lite")
	ctx := context.Background()
	model, err := gemini.NewModel(ctx, modelName, nil)
	if err != nil {
		log.Fatal("could not create model: ", modelName, err)
	}

	a, err := llmagent.New(llmagent.Config{
		Name:        "researcher",
		Model:       model,
		Instruction: "You help users research topics thoroughly.",
	})
	if err != nil {
		log.Fatal("could not create agent: ", err)
	}

	sessionSvc := session.InMemoryService()

	_, err = sessionSvc.Create(ctx, &session.CreateRequest{
		AppName:   appName,
		UserID:    userID,
		SessionID: sessionID,
	})
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	myRunner, err := runner.New(runner.Config{
		AppName:        appName,
		Agent:          a,
		SessionService: sessionSvc,
	})
	if err != nil {
		log.Fatalf("Failed to create runner: %v", err)
	}

	userMsg := &genai.Content{
		Role:  genai.RoleUser,
		Parts: []*genai.Part{genai.NewPartFromText("Explain AI in 1 sentence.")},
	}

	events := myRunner.Run(ctx, userID, sessionID, userMsg, agent.RunConfig{})

	fmt.Print("Agent Response: ")
	for event, err := range events {
		if err != nil {
			log.Printf("\nError during stream: %v", err)
			break
		}

		if event.Content != nil {
			for _, part := range event.Content.Parts {
				if part.Text != "" {
					fmt.Print(part.Text)
				}
			}
		}
	}
	fmt.Println()
}
