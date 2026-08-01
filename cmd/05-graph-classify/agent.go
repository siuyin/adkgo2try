package main

import (
	"context"
	"log"
	"os"
	"strings"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/workflowagent"
	"google.golang.org/adk/v2/cmd/launcher"
	"google.golang.org/adk/v2/cmd/launcher/full"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"
)

func main() {
	a, err := newProcessPipeline()
	if err != nil {
		log.Fatal("could not create sequential agent workflow")
	}
	config := &launcher.Config{
		AgentLoader: agent.NewSingleLoader(a),
	}

	l := full.NewLauncher()
	if err = l.Execute(context.Background(), config, os.Args[1:]); err != nil {
		log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}
}

// classifyMessage is the router node. It emits ev.Routes to select which
// branch to follow — the Go equivalent of Python's:
//
//	def router(node_input: str):
//	    return Event(route=["BUG"])
func classifyMessage(ctx agent.Context, msg string, emit func(*session.Event) error) (any, error) {
	// In a real workflow this step calls an LLM; here we classify by keyword.
	category := "LOGISTICS"
	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(lower, "bug") || strings.Contains(lower, "error"):
		category = "BUG"
	case strings.Contains(lower, "help") || strings.Contains(lower, "support"):
		category = "CUSTOMER_SUPPORT"
	}

	ev := session.NewEvent(ctx, ctx.InvocationID())
	ev.Routes = []string{category} // drives edge dispatch
	ev.Output = msg                // forward original message to the chosen handler
	if err := emit(ev); err != nil {
		return nil, err
	}
	return nil, nil // nil suppresses the automatic terminal event
}

// newProcessPipeline builds a classification + conditional-routing workflow
// using the v2 graph engine. The classifyMessage emitting node sets
// ev.Routes, and the graph engine dispatches to the matching handler via
// workflow.StringRoute.
//
// This is the Go equivalent of the Python Workflow example:
//
//	root_agent = Workflow(
//	    name="routing_workflow",
//	    edges=[
//	        ("START", process_message, router),
//	        (router, {
//	            "BUG": response_1_bug,
//	            "CUSTOMER_SUPPORT": response_2_support,
//	            "LOGISTICS": response_3_logistics,
//	        }),
//	    ],
//	)
func newProcessPipeline() (agent.Agent, error) {
	classifyNode := workflow.NewEmittingFunctionNode(
		"process_message", classifyMessage, workflow.NodeConfig{},
	)

	bugNode := workflow.NewFunctionNode("response_1_bug",
		func(_ agent.Context, _ any) (string, error) {
			return "Handling bug...", nil
		},
		workflow.NodeConfig{},
	)

	supportNode := workflow.NewFunctionNode("response_2_support",
		func(_ agent.Context, _ any) (string, error) {
			return "Handling customer support...", nil
		},
		workflow.NodeConfig{},
	)

	logisticsNode := workflow.NewFunctionNode("response_3_logistics",
		func(_ agent.Context, _ any) (string, error) {
			return "Handling logistics...", nil
		},
		workflow.NodeConfig{},
	)

	// workflow.Concat merges the sequential chain with the conditional edges.
	// Each workflow.Edge carries a workflow.StringRoute matcher that the engine
	// checks against ev.Routes emitted by classifyNode.
	edges := workflow.Concat(
		workflow.Chain(workflow.Start, classifyNode),
		[]workflow.Edge{
			{From: classifyNode, To: bugNode, Route: workflow.StringRoute("BUG")},
			{From: classifyNode, To: supportNode, Route: workflow.StringRoute("CUSTOMER_SUPPORT")},
			{From: classifyNode, To: logisticsNode, Route: workflow.StringRoute("LOGISTICS")},
		},
	)

	return workflowagent.New(workflowagent.Config{
		Name:        "routing_workflow",
		Description: "Classifies a message and routes it to the appropriate handler.",
		Edges:       edges,
	})
}
