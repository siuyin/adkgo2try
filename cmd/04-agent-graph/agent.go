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
	a, err := newSequentialGetStarted()
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

// cityTime holds the data passed from the lookup step to the report step.
type cityTime struct {
	City     string
	TimeInfo string
}

type cityInfo struct {
	City string
	Info string
}

// newSequentialGetStarted builds a three-node sequential workflow using the
// v2 graph engine. Each node is a workflow.NewFunctionNode whose return value
// is automatically wrapped in session.Event.Output and forwarded to the next
// node as its typed input.
//
// This is the Go equivalent of the Python Workflow example:
//
//	root_agent = Workflow(
//	    name="root_agent",
//	    edges=[("START", city_generator_agent, lookup_time_function,
//	             city_report_agent, completed_message_function)],
//	)
func newSequentialGetStarted() (agent.Agent, error) {
	// Step 1: return a city name. The string is set as event.Output and
	// becomes the typed input of the next node.
	cityGeneratorNode := workflow.NewFunctionNode("city_generator_agent",
		func(_ agent.Context, _ any) (cityInfo, error) {
			return cityInfo{"Tokyo", "The capital city of Japan"}, nil
		},
		workflow.NodeConfig{},
	)

	// Step 2: receive the city name and return structured time data.
	lookupTimeNode := workflow.NewFunctionNode("lookup_time_function",
		func(_ agent.Context, ci cityInfo) (cityTime, error) {
			return cityTime{City: ci.City, TimeInfo: "10:10 AM"}, nil
		},
		workflow.NodeConfig{},
	)

	// Step 3: receive the cityTime struct and produce the final report string.
	cityReportNode := workflow.NewFunctionNode("city_report_agent",
		func(_ agent.Context, ct cityTime) (string, error) {
			return fmt.Sprintf("It is %s in %s right now.\nWORKFLOW COMPLETED.",
				ct.TimeInfo, ct.City), nil
		},
		workflow.NodeConfig{},
	)

	// workflow.Chain wires START → cityGeneratorNode → lookupTimeNode → cityReportNode.
	// Data flows through event.Output: no session state writes needed.
	return workflowagent.New(workflowagent.Config{
		Name:        "root_agent",
		Description: "Sequential workflow: generate city → look up time → report.",
		Edges:       workflow.Chain(workflow.Start, cityGeneratorNode, lookupTimeNode, cityReportNode),
	})
}
