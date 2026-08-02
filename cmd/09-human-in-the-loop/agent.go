package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/workflowagent"
	"google.golang.org/adk/v2/cmd/launcher"
	"google.golang.org/adk/v2/cmd/launcher/full"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"
)

func main() {
	ctx := context.Background()

	ask := workflow.NewEmittingFunctionNode[any, any]("ask_name",
		func(ctx agent.Context, _ any, emit func(*session.Event) error) (any, error) {
			// Unique InterruptID per request: a reused ID would let the
			// Dev UI treat a later run's prompt as already answered.
			// Mirrors adk-python (RequestInput.interrupt_id defaults to
			// a fresh UUID).
			if err := emit(workflow.NewRequestInputEvent(ctx, session.RequestInput{
				InterruptID: "ask_name-" + uuid.NewString(),
				Message:     "What's your name?",
			})); err != nil {
				return nil, err
			}
			return nil, workflow.ErrNodeInterrupted
		},
		workflow.NodeConfig{},
	)

	// greet receives the user's reply (a string) and returns the
	// greeting. The classic NewFunctionNode is enough — no events
	// to emit beyond the terminal output.
	greet := workflow.NewFunctionNode("greet",
		func(_ agent.Context, name string) (string, error) {
			if name == "" {
				name = "stranger"
			}
			return fmt.Sprintf("Hello, %s!", name), nil
		},
		workflow.NodeConfig{},
	)

	rootAgent, err := workflowagent.New(workflowagent.Config{
		Name:        "hitl_simple",
		Description: "minimal HITL workflow for console launcher verification",
		Edges:       workflow.Chain(workflow.Start, ask, greet),
	})
	if err != nil {
		log.Fatalf("failed to create workflow agent: %v", err)
	}

	log.Printf("hitl_simple sample ready — type anything to start, then answer the prompt")

	launcherCfg := &launcher.Config{
		AgentLoader: agent.NewSingleLoader(rootAgent),
	}
	l := full.NewLauncher()
	if err := l.Execute(ctx, launcherCfg, os.Args[1:]); err != nil {
		log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}
}
