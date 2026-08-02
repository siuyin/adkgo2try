package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/workflowagent"
	"google.golang.org/adk/v2/cmd/launcher"
	"google.golang.org/adk/v2/cmd/launcher/full"
	"google.golang.org/adk/v2/workflow"
)

func main() {
	a, err := newParallelFanOut()
	if err != nil {
		log.Fatal("could not create fan out, fan in agent workflow: ", err)
	}
	config := &launcher.Config{
		AgentLoader: agent.NewSingleLoader(a),
	}

	l := full.NewLauncher()
	if err = l.Execute(context.Background(), config, os.Args[1:]); err != nil {
		log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}
}

// newParallelFanOut builds a fan-out / join workflow using the v2 graph engine.
// Three research nodes run in parallel from Start; workflow.NewJoinNode waits
// for all of them to complete and emits a map[nodeName]output to the format
// node, which assembles the results for a synthesis node.
//
// Graph topology:
//
//	START ─┬─> research_A ──┐
//	       ├─> research_B ──┼─> gather (JoinNode) ─> format ─> synthesis
//	       └─> research_C ──┘
//
// Python equivalent:
//
//	edges=[
//	    ("START", research_A, my_join_node),
//	    ("START", research_B, my_join_node),
//	    ("START", research_C, my_join_node),
//	    (my_join_node, format_node),
//	    (format_node, synthesis_node),
//	]
func newParallelFanOut() (agent.Agent, error) {
	researchA := workflow.NewFunctionNode("research_A",
		func(_ agent.Context, _ any) (string, error) {
			return "Fact about renewable energy.", nil
		},
		workflow.NodeConfig{},
	)
	researchB := workflow.NewFunctionNode("research_B",
		func(_ agent.Context, _ any) (string, error) {
			return "Fact about electric vehicles.", nil
		},
		workflow.NodeConfig{},
	)
	researchC := workflow.NewFunctionNode("research_C",
		func(_ agent.Context, _ any) (string, error) {
			return "Fact about carbon capture.", nil
		},
		workflow.NodeConfig{},
	)

	// workflow.NewJoinNode waits for all predecessors (research_A, research_B,
	// research_C) to complete and emits a map[nodeName]output to its successor.
	gatherNode := workflow.NewJoinNode("gather")

	// formatNode receives map[string]any from gatherNode and assembles a
	// combined prompt string.
	formatNode := workflow.NewFunctionNode("format",
		func(_ agent.Context, results map[string]any) (string, error) {
			return fmt.Sprintf("A: %v\nB: %v\nC: %v",
				results["research_A"],
				results["research_B"],
				results["research_C"],
			), nil
		},
		workflow.NodeConfig{},
	)

	synthesisNode := workflow.NewFunctionNode("synthesis",
		func(_ agent.Context, prompt string) (string, error) {
			return "Combined report: " + prompt, nil
		},
		workflow.NodeConfig{},
	)

	// EdgeBuilder.AddFanOut fans workflow.Start out to all three research nodes.
	// EdgeBuilder.AddFanIn routes all three research nodes into gatherNode.
	eb := workflow.NewEdgeBuilder()
	eb.AddFanOut(workflow.Start, researchA, researchB, researchC)
	eb.AddFanIn(gatherNode, researchA, researchB, researchC)
	eb.Add(gatherNode, formatNode)
	eb.Add(formatNode, synthesisNode)

	return workflowagent.New(workflowagent.Config{
		Name:        "fan_out_workflow",
		Description: "Parallel research fan-out with JoinNode barrier and synthesis.",
		Edges:       eb.Build(),
	})
}
