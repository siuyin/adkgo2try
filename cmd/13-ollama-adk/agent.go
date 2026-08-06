// This code only compilies within a go workspace defined thus:
//
// go 1.26.5
//
// use (
// 	/h/exp/adk-go-ollama
// 	/h/exp/adkgo2try
// )
// where adk-go-ollama is the ollama driver which points to https://github.com/siuyin/adk-go-ollama/tree/dev
// That repo was froked from github.com/craigh33/adk-go-ollama/ollama to initially add thinking mode mapping.

package main

import (
	"context"
	"log"
	"os"

	"github.com/craigh33/adk-go-ollama/ollama"
	"github.com/siuyin/dflt"
	"google.golang.org/genai"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/cmd/launcher"
	"google.golang.org/adk/v2/cmd/launcher/full"
)

func main() {
	ctx := context.Background()

	modelName := dflt.EnvString("MODEL", "ssfdre38/gemma4-nano:e2b")
	model, err := ollama.New(modelName)
	if err != nil {
		log.Fatalf("ollama model: %v", err)
	}
	a, err := llmagent.New(llmagent.Config{
		Name:        "local_agent",
		Model:       model,
		Description: "A helpful assistant",
		Instruction: "You reply briefly and clearly.",
		GenerateContentConfig: &genai.GenerateContentConfig{
			ThinkingConfig: &genai.ThinkingConfig{
				IncludeThoughts: true,
				ThinkingLevel:   genai.ThinkingLevel("false")},
			MaxOutputTokens: 512},
	})
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	config := &launcher.Config{
		AgentLoader: agent.NewSingleLoader(a),
	}

	l := full.NewLauncher()
	if err = l.Execute(ctx, config, os.Args[1:]); err != nil {
		log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}
}
