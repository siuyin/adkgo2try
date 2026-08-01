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
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/geminitool"
)

func main() {
	modelName := dflt.EnvString("MODEL", "gemma-4-31b-it")
	ctx := context.Background()
	model, _ := gemini.NewModel(ctx, modelName, nil)
	a, _ := llmagent.New(llmagent.Config{
		Name:        "researcher",
		Model:       model,
		Instruction: "You help users research topics thoroughly.",
		Tools:       []tool.Tool{geminitool.GoogleSearch{}},
	})
	config := &launcher.Config{
		AgentLoader: agent.NewSingleLoader(a),
	}

	l := full.NewLauncher()
	if err := l.Execute(ctx, config, os.Args[1:]); err != nil {
		log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}
}
