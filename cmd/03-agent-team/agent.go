package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/siuyin/dflt"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/cmd/launcher"
	"google.golang.org/adk/v2/cmd/launcher/full"
	"google.golang.org/adk/v2/model/gemini"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"
	"google.golang.org/genai"
)

func main() {
	ctx := context.Background()
	modelName := dflt.EnvString("MODEL", "gemma-4-31b-it")
	model, err := gemini.NewModel(ctx, modelName, &genai.ClientConfig{
		APIKey: os.Getenv("GOOGLE_API_KEY"),
	})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	a, err := llmagent.New(llmagent.Config{
		Name:        "weather_agent",
		Model:       model,
		Description: "The main coordinator agent. Handles weather questions and delegates greetings and farewells to specialists",
		Instruction: `You are the main weather agent coordinating a team. Your primary responsibility to to provide weather info.
		Use the weather tool ONLY for specific weather requests (eg. What is the weather in London?).
		You have specialized sub-agents:
		  greeting_agent: Handles simple greeetings like "hi", "hello". Delegate to it for these.
		  farewell_agent: Handles simple farewells like "Bye", "goodbye". Delegate to it for these.
		Analyse the user's query. If it is greeting, delegate to greeting_agent.
		If it is a farewell, delegate to farewell_agent.
		If it is a weather question, handle it yourself by using the get_weather tool.`,
		Tools: []tool.Tool{
			weatherTool(),
		},
		SubAgents: []agent.Agent{
			greetingAgent(),
			farewellAgent(),
		},
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

type CityArgs struct {
	City string `json:"city"`
}

func weatherTool() tool.Tool {
	wt, err := functiontool.New[CityArgs, map[string]any](
		functiontool.Config{
			Name:        "get_weather",
			Description: "Retrieves the current weather report for a specified city.",
		},
		func(ctx agent.Context, args CityArgs) (map[string]any, error) {
			if strings.EqualFold(args.City, "new york") {
				return map[string]any{
					"status": "success",
					"report": "The weather in New York is sunny with a temperature of 25 degrees Celsius (77 degrees Fahrenheit).",
				}, nil
			}
			return map[string]any{
				"status":        "error",
				"error_message": "Weather information for '" + args.City + "' is not available.",
			}, nil
		},
	)
	if err != nil {
		log.Fatalf("Failed to create get_weather tool: %v", err)
	}
	return wt
}
