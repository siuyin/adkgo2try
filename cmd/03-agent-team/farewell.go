package main

import (
	"context"
	"log"

	"github.com/siuyin/dflt"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model/gemini"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"
	"google.golang.org/genai"
)

func farewellAgent() agent.Agent {
	model, err := gemini.NewModel(context.Background(),
		dflt.EnvString("GOODBYE_MODEL", "gemma-4-31b-it"), &genai.ClientConfig{})
	if err != nil {
		log.Fatal("could not create goodbye model.")
	}
	gba, err := llmagent.New(llmagent.Config{
		Name:        "goodbye_agent",
		Model:       model,
		Description: "Handles simple farewells and goodbyes using the say_goodbye tool.",
		Instruction: `You are the farewell agent. Your ONLY task is to provide a polite farewell message.
		Use the say_goodbye tool when the user indicates they are leaving or ending the conversation (eg. Bye or I'm out of here).
		Do not perform any other action.`,
		Tools: []tool.Tool{goodbyeTool()},
	})
	if err != nil {
		log.Fatal("could not create goodbye agent")
	}
	return gba
}

type NullArgs struct{}

func goodbyeTool() tool.Tool {
	gt, err := functiontool.New[NullArgs, map[string]any](
		functiontool.Config{
			Name:        "say_goodbye",
			Description: "Provides a simple farewell message to conclude the conversation.",
		},
		func(ctx agent.Context, args NullArgs) (map[string]any, error) {
			msg := "Goodbye. Have a great day!"
			return map[string]any{
				"status":  "success",
				"message": msg,
			}, nil

		},
	)
	if err != nil {
		log.Fatalf("Failed to create get_weather tool: %v", err)
	}

	return gt
}
