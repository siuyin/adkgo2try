package main

import (
	"context"
	"log"
	"os"

	"github.com/siuyin/dflt"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/cmd/launcher"
	"google.golang.org/adk/v2/cmd/launcher/full"
	"google.golang.org/adk/v2/model/gemini"
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
		GenerateContentConfig: &genai.GenerateContentConfig{
			ThinkingConfig: &genai.ThinkingConfig{
				IncludeThoughts: true,
				ThinkingLevel:   genai.ThinkingLevelUnspecified},
		},
	})
	if err != nil {
		log.Fatal("could not create agent: ", err)
	}

	sessionSvc := session.InMemoryService()
	if _, err := sessionSvc.Create(ctx, &session.CreateRequest{
		AppName:   appName,
		UserID:    userID,
		SessionID: sessionID,
	}); err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	config := &launcher.Config{
		SessionService: sessionSvc,
		AgentLoader:    agent.NewSingleLoader(a),
	}

	l := full.NewLauncher()
	if err := l.Execute(ctx, config, os.Args[1:]); err != nil {
		log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}

}
