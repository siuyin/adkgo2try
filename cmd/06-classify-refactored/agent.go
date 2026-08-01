package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/workflowagent"
	"google.golang.org/adk/v2/cmd/launcher"
	"google.golang.org/adk/v2/cmd/launcher/full"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"
	"google.golang.org/genai"
)

func main() {
	a, err := newProcessPipeline()
	if err != nil {
		log.Fatal("could not create sequential agent workflow: ", err)
	}
	config := &launcher.Config{
		AgentLoader: agent.NewSingleLoader(a),
	}

	l := full.NewLauncher()
	if err = l.Execute(context.Background(), config, os.Args[1:]); err != nil {
		log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}
}

func newProcessPipeline() (agent.Agent, error) {
	// workflow.Concat merges the sequential chain with the conditional edges.
	// Each workflow.Edge carries a workflow.StringRoute matcher that the engine
	// checks against ev.Routes emitted by classifyNode.
	cn := classifyNode()
	edges := workflow.Concat(
		workflow.Chain(workflow.Start, cn),
		[]workflow.Edge{
			{From: cn, To: bugNode(), Route: workflow.StringRoute("BUG")},
			{From: cn, To: supportNode(), Route: workflow.StringRoute("CUSTOMER_SUPPORT")},
			{From: cn, To: logisticsNode(), Route: workflow.StringRoute("LOGISTICS")},
		},
	)

	return workflowagent.New(workflowagent.Config{
		Name:        "routing_workflow",
		Description: "Classifies a message and routes it to the appropriate handler.",
		Edges:       edges,
	})
}

func classifyNode() *workflow.FunctionNode {
	return workflow.NewEmittingFunctionNode(
		"process_message", classifyMessage, workflow.NodeConfig{},
	)
}

type nilOutput any

func classifyMessage(ctx agent.Context, msg string, emit func(*session.Event) error) (nilOutput, error) {
	category := categorize(msg)

	ev := session.NewEvent(ctx, ctx.InvocationID())
	ev.Routes = []string{category} // drives edge dispatch
	ev.Content = genai.NewContentFromText(fmt.Sprintf("classifying %q\n", msg), genai.RoleModel)
	ev.Output = msg // forward original message to the chosen handler

	if err := emit(ev); err != nil {
		return nil, err
	}

	return nil, nil // nil suppresses the automatic terminal event
}
func categorize(msg string) string {
	category := "LOGISTICS"
	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(lower, "bug") || strings.Contains(lower, "error"):
		category = "BUG"
	case strings.Contains(lower, "help") || strings.Contains(lower, "support"):
		category = "CUSTOMER_SUPPORT"
	}

	return category
}

func bugNode() *workflow.FunctionNode {
	return workflow.NewFunctionNode("response_1_bug",
		func(_ agent.Context, msg string) (string, error) {
			return "Handling bug...: " + msg, nil
		},
		workflow.NodeConfig{},
	)
}
func supportNode() *workflow.FunctionNode {
	return workflow.NewFunctionNode("response_2_support",
		func(_ agent.Context, msg string) (string, error) {
			return "Handling customer support...: " + msg, nil
		},
		workflow.NodeConfig{},
	)
}
func logisticsNode() *workflow.FunctionNode {
	return workflow.NewFunctionNode("response_3_logistics",
		func(_ agent.Context, msg string) (string, error) {
			return "Handling logistics...: " + msg, nil
		},
		workflow.NodeConfig{},
	)
}
