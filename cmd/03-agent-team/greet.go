package main

import (
	"context"
	"fmt"
	"log"

	"github.com/siuyin/dflt"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model/gemini"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"
	"google.golang.org/genai"
)

func greetingAgent() agent.Agent {
	model, err := gemini.NewModel(context.Background(),
		dflt.EnvString("GREETING_MODEL", "gemma-4-31b-it"), &genai.ClientConfig{})
	if err != nil {
		log.Fatal("could not create greeting model")
	}
	ga, err := llmagent.New(llmagent.Config{
		Name:        "greeting_agent",
		Model:       model,
		Description: "Handles simple greetings and hellos with the say_hello tool.",
		Instruction: `You are the greeting agent. Your ONLY task is to provide a friendly greeting to the user.
		Use the say_hello tool to generate the greeting.
		If the user provides their name, make sure to pass it to the tool.
		Do not engage in any other conversation or tasks.`,
		Tools: []tool.Tool{helloTool()},
	})
	if err != nil {
		log.Fatal("could not create greeting agent")
	}
	return ga
}

type NameArgs struct {
	Name string `json:"name,omitempty"`
}

func helloTool() tool.Tool {
	ht, err := functiontool.New[NameArgs, map[string]any](
		functiontool.Config{
			Name:        "say_hello",
			Description: "Provides a simple greeting. If a name is provided it will be used.",
		},
		func(ctx agent.Context, args NameArgs) (map[string]any, error) {
			greeting := "Hello there!"
			if args.Name != "" {
				greeting = fmt.Sprintf("Hello %s!", args.Name)
			}
			return map[string]any{
				"status":   "success",
				"greeting": greeting,
			}, nil

		},
	)
	if err != nil {
		log.Fatalf("Failed to create get_weather tool: %v", err)
	}

	return ht
}
