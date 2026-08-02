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
	"google.golang.org/adk/v2/workflow"
)

func main() {
	a, err := newNestedWorkflows()
	if err != nil {
		log.Fatal("could not create nested agent workflow: ", err)
	}
	config := &launcher.Config{
		AgentLoader: agent.NewSingleLoader(a),
	}

	l := full.NewLauncher()
	if err = l.Execute(context.Background(), config, os.Args[1:]); err != nil {
		log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}
}

// newNestedWorkflows shows how to nest one workflowagent inside another using
// the v2 graph engine. The inner workflowagent is wrapped with
// workflow.NewAgentNode and placed as a node in the outer graph's edge slice.
// From the outer graph's perspective the inner workflow is a single node that
// runs to completion before the edge to finalNode is followed.
//
// Python equivalent:
//
//	root_agent = Workflow(
//	    name="parent_workflow",
//	    edges=[("START", task_A1, workflow_B, final_node)],
//	)
func newNestedWorkflows() (agent.Agent, error) {
	// --- Inner workflow B ---
	innerStep1 := workflow.NewFunctionNode("inner_step_1",
		func(_ agent.Context, input string) (string, error) {
			return "[ES] " + input, nil // simulate translation to Spanish
		},
		workflow.NodeConfig{},
	)
	innerStep2 := workflow.NewFunctionNode("inner_step_2",
		func(ctx agent.Context, spanish string) (string, error) {
			ctx.State().Set("output", "EN "+spanish)
			return "[EN] " + spanish, nil // simulate translation back to English
		},
		workflow.NodeConfig{},
	)

	// workflowB is a self-contained inner graph.
	workflowB, err := workflowagent.New(workflowagent.Config{
		Name:        "workflow_B",
		Description: "Translates input to Spanish then back to English.",
		Edges:       workflow.Chain(workflow.Start, innerStep1, innerStep2),
	})
	if err != nil {
		return nil, fmt.Errorf("workflowB: %w", err)
	}

	// --- Outer graph ---
	taskA1 := workflow.NewFunctionNode("task_A1",
		func(_ agent.Context, input string) (string, error) {
			return "Summary: " + strings.TrimSpace(input), nil
		},
		workflow.NodeConfig{},
	)

	finalNode := workflow.NewFunctionNode("final_node",
		//func(_ agent.Context, result string) (string, error) {
		//	return "Final: " + result, nil
		func(ctx agent.Context, res string) (string, error) {
			out, _ := ctx.State().Get("output")
			return "Final: " + out.(string), nil
		},
		workflow.NodeConfig{},
	)

	// workflow.NewAgentNode wraps workflowB so it can be placed as a node
	// in the outer graph's edges slice.
	innerNode, err := workflow.NewAgentNode(workflowB, workflow.NodeConfig{})
	if err != nil {
		return nil, fmt.Errorf("NewAgentNode(workflowB): %w", err)
	}

	return workflowagent.New(workflowagent.Config{
		Name:        "parent_workflow",
		Description: "Runs task_A1 then the nested workflow_B then final_node.",
		Edges:       workflow.Chain(workflow.Start, taskA1, innerNode, finalNode),
		SubAgents:   []agent.Agent{workflowB},
	})
}
