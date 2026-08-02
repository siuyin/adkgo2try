package main

import (
	"context"
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
	rootAgent, err := newResumeOrRequestInputAgent()
	if err != nil {
		log.Fatalf("failed to create workflow agent: %v", err)
	}

	launcherCfg := &launcher.Config{
		AgentLoader: agent.NewSingleLoader(rootAgent),
	}
	l := full.NewLauncher()

	if err := l.Execute(context.Background(), launcherCfg, os.Args[1:]); err != nil {
		log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}
}

func newResumeOrRequestInputAgent() (agent.Agent, error) {
	ask := workflow.NewEmittingFunctionNode("ask",
		func(ctx agent.Context, _ any, emit func(*session.Event) error) (*float64, error) {
			reply, err := workflow.ResumeOrRequestInput(ctx, emit, session.RequestInput{
				InterruptID: "myID-" + uuid.NewString(), Message: "Enter a number",
			})
			if err != nil {
				// ErrNodeInterrupted on first pass — workflow pauses here.
				return nil, err
			}

			numP, _ := reply.(*float64)
			return numP, nil
		}, workflow.NodeConfig{})

	rerun := true
	double := workflow.NewFunctionNode("double",
		func(_ agent.Context, inp *float64) (float64, error) {
			return (*inp) * 2.0, nil
		}, workflow.NodeConfig{RerunOnResume: &rerun})

	return workflowagent.New(workflowagent.Config{
		Name:        "hitl_resume_or_request_input",
		Description: "HITL workflow to resume or request for input",
		Edges:       workflow.Chain(workflow.Start, ask, double),
	})
}
